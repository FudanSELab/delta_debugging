// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	restful "github.com/emicklei/go-restful"
	_ "github.com/golang/glog" // TODO(nmittler): Remove this
	multierror "github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/util"
	"istio.io/istio/pkg/version"
)

const (
	metricsNamespace     = "pilot"
	metricsSubsystem     = "discovery"
	metricLabelCacheName = "cache_name"
	metricLabelMethod    = "method"
	metricBuildVersion   = "build_version"
)

var (
	// Save the build version information.
	buildVersion = version.Info.String()

	cacheSizeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "cache_size",
			Help:      "Current size (in bytes) of a single cache within Pilot",
		}, []string{metricLabelCacheName, metricBuildVersion})
	cacheHitCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "cache_hit",
			Help:      "Count of cache hits for a particular cache within Pilot",
		}, []string{metricLabelCacheName, metricBuildVersion})
	cacheMissCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "cache_miss",
			Help:      "Count of cache misses for a particular cache within Pilot",
		}, []string{metricLabelCacheName, metricBuildVersion})
	callCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "calls",
			Help:      "Counter of individual method calls in Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})
	errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "errors",
			Help:      "Counter of errors encountered during a given method call within Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})
	webhookCallCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "webhook_calls",
			Help:      "Counter of individual webhook calls made in Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})
	webhookErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "webhook_errors",
			Help:      "Counter of errors encountered when invoking the webhook endpoint within Pilot",
		}, []string{metricLabelMethod, metricBuildVersion})

	resourceBuckets = []float64{0, 10, 20, 30, 40, 50, 75, 100, 150, 250, 500, 1000, 10000}
	resourceCounter = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "resources",
			Help:      "Histogram of returned resource counts per method by Pilot",
			Buckets:   resourceBuckets,
		}, []string{metricLabelMethod, metricBuildVersion})
)

var (
	// Variables associated with clear cache squashing.
	lastClearCache     time.Time
	clearCacheTimerSet bool
	clearCacheMutex    sync.Mutex
	clearCacheTime     = 1
)

func init() {
	prometheus.MustRegister(cacheSizeGauge)
	prometheus.MustRegister(cacheHitCounter)
	prometheus.MustRegister(cacheMissCounter)
	prometheus.MustRegister(callCounter)
	prometheus.MustRegister(errorCounter)
	prometheus.MustRegister(resourceCounter)

	cacheSquash := os.Getenv("PILOT_CACHE_SQUASH")
	if len(cacheSquash) > 0 {
		t, err := strconv.Atoi(cacheSquash)
		if err == nil {
			clearCacheTime = t
		}
	}
}

// DiscoveryService publishes services, clusters, and routes for all proxies
type DiscoveryService struct {
	model.Environment
	server          *http.Server
	webhookClient   *http.Client
	webhookEndpoint string
	// TODO Profile and optimize cache eviction policy to avoid
	// flushing the entire cache when any route, service, or endpoint
	// changes. An explicit cache expiration policy should be
	// considered with this change to avoid memory exhaustion as the
	// entire cache will no longer be periodically flushed and stale
	// entries can linger in the cache indefinitely.
	sdsCache *discoveryCache
	cdsCache *discoveryCache
	rdsCache *discoveryCache
	ldsCache *discoveryCache
}

type discoveryCacheStatEntry struct {
	Hit  uint64 `json:"hit"`
	Miss uint64 `json:"miss"`
}

type discoveryCacheStats struct {
	Stats map[string]*discoveryCacheStatEntry `json:"cache_stats"`
}

type discoveryCacheEntry struct {
	data          []byte
	hit           uint64 // atomic
	miss          uint64 // atomic
	resourceCount uint32
}

type discoveryCache struct {
	name     string
	disabled bool
	mu       sync.RWMutex
	cache    map[string]*discoveryCacheEntry
}

func newDiscoveryCache(name string, enabled bool) *discoveryCache {
	return &discoveryCache{
		name:     name,
		disabled: !enabled,
		cache:    make(map[string]*discoveryCacheEntry),
	}
}

func (c *discoveryCache) cachedDiscoveryResponse(key string) ([]byte, uint32, bool) {
	if c.disabled {
		return nil, 0, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Miss - entry.miss is updated in updateCachedDiscoveryResponse
	entry, ok := c.cache[key]
	if !ok || entry.data == nil {
		return nil, 0, false
	}

	// Hit
	atomic.AddUint64(&entry.hit, 1)
	cacheHitCounter.With(c.cacheSizeLabels()).Inc()
	return entry.data, entry.resourceCount, true
}

func (c *discoveryCache) updateCachedDiscoveryResponse(key string, resourceCount uint32, data []byte) {
	if c.disabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.cache[key]
	var cacheSizeDelta float64
	if !ok {
		entry = &discoveryCacheEntry{}
		c.cache[key] = entry
		cacheSizeDelta = float64(len(key) + len(data))
	} else if entry.data != nil {
		cacheSizeDelta = float64(len(data) - len(entry.data))
		log.Warnf("Overriding cached data for entry %v", key)
	}
	entry.resourceCount = resourceCount
	entry.data = data
	atomic.AddUint64(&entry.miss, 1)
	cacheMissCounter.With(c.cacheSizeLabels()).Inc()
	cacheSizeGauge.With(c.cacheSizeLabels()).Add(cacheSizeDelta)
}

func (c *discoveryCache) clear() {
	// Reset the cache size metric for this cache.
	cacheSizeGauge.Delete(c.cacheSizeLabels())

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range c.cache {
		v.data = nil
	}
}

func (c *discoveryCache) resetStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.cache {
		atomic.StoreUint64(&v.hit, 0)
		atomic.StoreUint64(&v.miss, 0)
	}
}

func (c *discoveryCache) stats() map[string]*discoveryCacheStatEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]*discoveryCacheStatEntry, len(c.cache))
	for k, v := range c.cache {
		stats[k] = &discoveryCacheStatEntry{
			Hit:  atomic.LoadUint64(&v.hit),
			Miss: atomic.LoadUint64(&v.miss),
		}
	}
	return stats
}

func (c *discoveryCache) cacheSizeLabels() prometheus.Labels {
	return prometheus.Labels{
		metricLabelCacheName: c.name,
		metricBuildVersion:   buildVersion,
	}
}

type hosts struct {
	Hosts []*host `json:"hosts"`
}

type host struct {
	Address string `json:"ip_address"`
	Port    int    `json:"port"`
	Tags    *tags  `json:"tags,omitempty"`
}

type tags struct {
	AZ     string `json:"az,omitempty"`
	Canary bool   `json:"canary,omitempty"`

	// Weight is an integer in the range [1, 100] or empty
	Weight int `json:"load_balancing_weight,omitempty"`
}

type ldsResponse struct {
	Listeners Listeners `json:"listeners"`
}

type keyAndService struct {
	Key   string  `json:"service-key"`
	Hosts []*host `json:"hosts"`
}

// Request parameters for discovery services
const (
	ServiceKey      = "service-key"
	ServiceCluster  = "service-cluster"
	ServiceNode     = "service-node"
	RouteConfigName = "route-config-name"
)

// DiscoveryServiceOptions contains options for create a new discovery
// service instance.
type DiscoveryServiceOptions struct {
	Port            int
	MonitoringPort  int
	EnableProfiling bool
	EnableCaching   bool
	WebhookEndpoint string
}

// NewDiscoveryService creates an Envoy discovery service on a given port
func NewDiscoveryService(ctl model.Controller, configCache model.ConfigStoreCache,
	environment model.Environment, o DiscoveryServiceOptions) (*DiscoveryService, error) {
	out := &DiscoveryService{
		Environment: environment,
		sdsCache:    newDiscoveryCache("sds", o.EnableCaching),
		cdsCache:    newDiscoveryCache("cds", o.EnableCaching),
		rdsCache:    newDiscoveryCache("rds", o.EnableCaching),
		ldsCache:    newDiscoveryCache("lds", o.EnableCaching),
	}

	container := restful.NewContainer()
	if o.EnableProfiling {
		container.ServeMux.HandleFunc("/debug/pprof/", pprof.Index)
		container.ServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		container.ServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		container.ServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		container.ServeMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	out.Register(container)

	out.webhookEndpoint, out.webhookClient = util.NewWebHookClient(o.WebhookEndpoint)

	out.server = &http.Server{Addr: ":" + strconv.Itoa(o.Port), Handler: container}

	// Flush cached discovery responses whenever services, service
	// instances, or routing configuration changes.
	serviceHandler := func(*model.Service, model.Event) { out.clearCache() }
	if err := ctl.AppendServiceHandler(serviceHandler); err != nil {
		return nil, err
	}
	instanceHandler := func(*model.ServiceInstance, model.Event) { out.clearCache() }
	if err := ctl.AppendInstanceHandler(instanceHandler); err != nil {
		return nil, err
	}

	if configCache != nil {
		// TODO: changes should not trigger a full recompute of LDS/RDS/CDS/EDS
		// (especially mixerclient HTTP and quota)
		configHandler := func(model.Config, model.Event) { out.clearCache() }
		for _, descriptor := range model.IstioConfigTypes {
			configCache.RegisterEventHandler(descriptor.Type, configHandler)
		}
	}

	return out, nil
}

// Register adds routes a web service container. This is visible for testing purposes only.
func (ds *DiscoveryService) Register(container *restful.Container) {
	ws := &restful.WebService{}
	ws.Produces(restful.MIME_JSON)

	// List all known services (informational, not invoked by Envoy)
	ws.Route(ws.
		GET("/v1/registration").
		To(ds.ListAllEndpoints).
		Doc("Services in SDS"))

	// This route makes discovery act as an Envoy Service discovery service (SDS).
	// See https://envoyproxy.github.io/envoy/intro/arch_overview/service_discovery.html#service-discovery-service-sds
	ws.Route(ws.
		GET(fmt.Sprintf("/v1/registration/{%s}", ServiceKey)).
		To(ds.ListEndpoints).
		Doc("SDS registration").
		Param(ws.PathParameter(ServiceKey, "tuple of service name and tag name").DataType("string")))

	// This route makes discovery act as an Envoy Cluster discovery service (CDS).
	// See https://envoyproxy.github.io/envoy/configuration/cluster_manager/cds.html#config-cluster-manager-cds
	ws.Route(ws.
		GET(fmt.Sprintf("/v1/clusters/{%s}/{%s}", ServiceCluster, ServiceNode)).
		To(ds.ListClusters).
		Doc("CDS registration").
		Param(ws.PathParameter(ServiceCluster, "client proxy service cluster").DataType("string")).
		Param(ws.PathParameter(ServiceNode, "client proxy service node").DataType("string")))

	// This route makes discovery act as an Envoy Route discovery service (RDS).
	// See https://lyft.github.io/envoy/docs/configuration/http_conn_man/rds.html
	ws.Route(ws.
		GET(fmt.Sprintf("/v1/routes/{%s}/{%s}/{%s}", RouteConfigName, ServiceCluster, ServiceNode)).
		To(ds.ListRoutes).
		Doc("RDS registration").
		Param(ws.PathParameter(RouteConfigName, "route configuration name").DataType("string")).
		Param(ws.PathParameter(ServiceCluster, "client proxy service cluster").DataType("string")).
		Param(ws.PathParameter(ServiceNode, "client proxy service node").DataType("string")))

	// This route responds to LDS requests
	// See https://lyft.github.io/envoy/docs/configuration/listeners/lds.html
	ws.Route(ws.
		GET(fmt.Sprintf("/v1/listeners/{%s}/{%s}", ServiceCluster, ServiceNode)).
		To(ds.ListListeners).
		Doc("LDS registration").
		Param(ws.PathParameter(ServiceCluster, "client proxy service cluster").DataType("string")).
		Param(ws.PathParameter(ServiceNode, "client proxy service node").DataType("string")))

	// This route retrieves the Availability Zone of the service node requested
	ws.Route(ws.
		GET(fmt.Sprintf("/v1/az/{%s}/{%s}", ServiceCluster, ServiceNode)).
		To(ds.AvailabilityZone).
		Doc("AZ for service node").
		Param(ws.PathParameter(ServiceCluster, "client proxy service cluster").DataType("string")).
		Param(ws.PathParameter(ServiceNode, "client proxy service node").DataType("string")))

	ws.Route(ws.
		GET("/cache_stats").
		To(ds.GetCacheStats).
		Doc("Get discovery service cache stats").
		Writes(discoveryCacheStats{}))

	ws.Route(ws.
		POST("/cache_stats_delete").
		To(ds.ClearCacheStats).
		Doc("Clear discovery service cache stats"))

	container.Add(ws)
}

// Start starts the Pilot discovery service on the port specified in DiscoveryServiceOptions. If Port == 0, a
// port number is automatically chosen. This method returns the address on which the server is listening for incoming
// connections. Content serving is started by this method, but is executed asynchronously. Serving can be cancelled
// at any time by closing the provided stop channel.
func (ds *DiscoveryService) Start(stop chan struct{}) (net.Addr, error) {
	addr := ds.server.Addr
	if addr == "" {
		addr = ":http"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			if err := ds.server.Serve(listener); err != nil {
				log.Warna(err)
			}
		}()

		// Wait for the stop notification and shutdown the server.
		<-stop
		err := ds.server.Close()
		if err != nil {
			log.Warna(err)
		}
	}()

	log.Infof("Discovery service started at %s", listener.Addr().String())
	return listener.Addr(), nil
}

// GetCacheStats returns the statistics for cached discovery responses.
func (ds *DiscoveryService) GetCacheStats(_ *restful.Request, response *restful.Response) {
	stats := make(map[string]*discoveryCacheStatEntry)
	for k, v := range ds.sdsCache.stats() {
		stats[k] = v
	}
	for k, v := range ds.cdsCache.stats() {
		stats[k] = v
	}
	for k, v := range ds.rdsCache.stats() {
		stats[k] = v
	}
	for k, v := range ds.ldsCache.stats() {
		stats[k] = v
	}
	if err := response.WriteEntity(discoveryCacheStats{stats}); err != nil {
		log.Warna(err)
	}
}

// ClearCacheStats clear the statistics for cached discovery responses.
func (ds *DiscoveryService) ClearCacheStats(_ *restful.Request, _ *restful.Response) {
	ds.sdsCache.resetStats()
	ds.cdsCache.resetStats()
	ds.rdsCache.resetStats()
	ds.ldsCache.resetStats()
}

// clearCache will clear all envoy caches. Called by service, instance and config handlers.
// This will impact the performance, since envoy will need to recalculate.
func (ds *DiscoveryService) clearCache() {
	clearCacheMutex.Lock()
	defer clearCacheMutex.Unlock()

	if time.Since(lastClearCache) < time.Duration(clearCacheTime)*time.Second {
		if !clearCacheTimerSet {
			clearCacheTimerSet = true
			time.AfterFunc(time.Duration(clearCacheTime)*time.Second, func() {
				clearCacheTimerSet = false
				ds.clearCache() // it's after time - so will clear the cache
			})
		}
		return
	}
	// TODO: clear the RDS few seconds after CDS !!
	lastClearCache = time.Now()
	log.Infof("Cleared discovery service cache")
	ds.sdsCache.clear()
	ds.cdsCache.clear()
	ds.rdsCache.clear()
	ds.ldsCache.clear()
}

// ListAllEndpoints responds with all Services and is not restricted to a single service-key
func (ds *DiscoveryService) ListAllEndpoints(_ *restful.Request, response *restful.Response) {
	methodName := "ListAllEndpoints"
	incCalls(methodName)

	services := make([]*keyAndService, 0)

	svcs, err := ds.Services()
	if err != nil {
		// If client experiences an error, 503 error will tell envoy to keep its current
		// cache and try again later
		errorResponse(methodName, response, http.StatusServiceUnavailable, "EDS "+err.Error())
		return
	}

	for _, service := range svcs {
		if !service.External() {
			for _, port := range service.Ports {
				hosts := make([]*host, 0)
				instances, err := ds.Instances(service.Hostname, []string{port.Name}, nil)
				if err != nil {
					// If client experiences an error, 503 error will tell envoy to keep its current
					// cache and try again later
					errorResponse(methodName, response, http.StatusInternalServerError, "EDS "+err.Error())
					return
				}
				for _, instance := range instances {
					// Only set tags if theres an AZ to set, ensures nil tags when there isnt
					var t *tags
					if instance.AvailabilityZone != "" {
						t = &tags{AZ: instance.AvailabilityZone}
					}
					hosts = append(hosts, &host{
						Address: instance.Endpoint.Address,
						Port:    instance.Endpoint.Port,
						Tags:    t,
					})
				}
				services = append(services, &keyAndService{
					Key:   service.Key(port, nil),
					Hosts: hosts,
				})
			}
		}
	}

	// Sort servicesArray.  This is not strictly necessary, but discovery_test.go will
	// be comparing against a golden example using test/util/diff.go which does a textual comparison
	sort.Slice(services, func(i, j int) bool { return services[i].Key < services[j].Key })

	if err := response.WriteEntity(services); err != nil {
		incErrors(methodName)
		log.Warna(err)
	} else {
		observeResources(methodName, uint32(len(services)))
	}
}

// ListEndpoints responds to EDS requests
func (ds *DiscoveryService) ListEndpoints(request *restful.Request, response *restful.Response) {
	methodName := "ListEndpoints"
	incCalls(methodName)

	key := request.Request.URL.String()
	out, resourceCount, cached := ds.sdsCache.cachedDiscoveryResponse(key)
	if !cached {
		hostname, ports, tags := model.ParseServiceKey(request.PathParameter(ServiceKey))
		// envoy expects an empty array if no hosts are available
		hostArray := make([]*host, 0)
		endpoints, err := ds.Instances(hostname, ports.GetNames(), tags)
		if err != nil {
			// If client experiences an error, 503 error will tell envoy to keep its current
			// cache and try again later
			errorResponse(methodName, response, http.StatusServiceUnavailable, "EDS "+err.Error())
			return
		}
		for _, ep := range endpoints {
			hostArray = append(hostArray, &host{
				Address: ep.Endpoint.Address,
				Port:    ep.Endpoint.Port,
			})
		}
		if out, err = json.MarshalIndent(hosts{Hosts: hostArray}, " ", " "); err != nil {
			errorResponse(methodName, response, http.StatusInternalServerError, "EDS "+err.Error())
			return
		}
		resourceCount = uint32(len(endpoints))
		if resourceCount > 0 {
			ds.sdsCache.updateCachedDiscoveryResponse(key, resourceCount, out)
		}
	}
	observeResources(methodName, resourceCount)
	writeResponse(response, out)
}

func (ds *DiscoveryService) parseDiscoveryRequest(request *restful.Request) (model.Proxy, error) {
	nodeInfo := request.PathParameter(ServiceNode)
	svcNode, err := model.ParseServiceNode(nodeInfo)
	if err != nil {
		return svcNode, multierror.Prefix(err, fmt.Sprintf("unexpected %s: ", ServiceNode))
	}
	return svcNode, nil
}

// AvailabilityZone responds to requests for an AZ for the given cluster node
func (ds *DiscoveryService) AvailabilityZone(request *restful.Request, response *restful.Response) {
	methodName := "AvailabilityZone"
	incCalls(methodName)

	svcNode, err := ds.parseDiscoveryRequest(request)
	if err != nil {
		errorResponse(methodName, response, http.StatusNotFound, "AvailabilityZone "+err.Error())
		return
	}
	proxyInstances, err := ds.GetProxyServiceInstances(svcNode)
	if err != nil {
		errorResponse(methodName, response, http.StatusNotFound, "AvailabilityZone "+err.Error())
		return
	}
	if len(proxyInstances) <= 0 {
		errorResponse(methodName, response, http.StatusNotFound, "AvailabilityZone couldn't find the given cluster node")
		return
	}
	// All instances are going to have the same IP addr therefore will all be in the same AZ
	writeResponse(response, []byte(proxyInstances[0].AvailabilityZone))
}

// ListClusters responds to CDS requests for all outbound clusters
func (ds *DiscoveryService) ListClusters(request *restful.Request, response *restful.Response) {
	methodName := "ListClusters"
	incCalls(methodName)

	key := request.Request.URL.String()
	out, resourceCount, cached := ds.cdsCache.cachedDiscoveryResponse(key)
	transformedOutput := out
	if !cached {
		svcNode, err := ds.parseDiscoveryRequest(request)
		if err != nil {
			errorResponse(methodName, response, http.StatusNotFound, "CDS "+err.Error())
			return
		}

		clusters, err := buildClusters(ds.Environment, svcNode)
		if err != nil {
			// If client experiences an error, 503 error will tell envoy to keep its current
			// cache and try again later
			errorResponse(methodName, response, http.StatusServiceUnavailable, "CDS "+err.Error())
			return
		}
		if out, err = json.MarshalIndent(ClusterManager{Clusters: clusters}, " ", " "); err != nil {
			errorResponse(methodName, response, http.StatusInternalServerError, "CDS "+err.Error())
			return
		}

		transformedOutput, err = ds.invokeWebhook(request.Request.URL.Path, out, "webhook"+methodName)
		if err != nil {
			// Use whatever we generated.
			transformedOutput = out
		}

		// TODO: this is wrong as it doesn't take into account clusters added by webhook
		resourceCount = uint32(len(clusters))
		// TODO: BUG. if resourceCount is 0, but transformedOutput has added resources, the cache wont update
		if resourceCount > 0 {
			ds.cdsCache.updateCachedDiscoveryResponse(key, resourceCount, transformedOutput)
		}
	}

	observeResources(methodName, resourceCount)
	writeResponse(response, transformedOutput)
}

// ListListeners responds to LDS requests
func (ds *DiscoveryService) ListListeners(request *restful.Request, response *restful.Response) {
	methodName := "ListListeners"
	incCalls(methodName)

	key := request.Request.URL.String()
	out, resourceCount, cached := ds.ldsCache.cachedDiscoveryResponse(key)
	transformedOutput := out
	if !cached {
		svcNode, err := ds.parseDiscoveryRequest(request)
		if err != nil {
			errorResponse(methodName, response, http.StatusNotFound, "LDS "+err.Error())
			return
		}

		listeners, err := buildListeners(ds.Environment, svcNode)
		if err != nil {
			// If client experiences an error, 503 error will tell envoy to keep its current
			// cache and try again later
			errorResponse(methodName, response, http.StatusServiceUnavailable, "LDS "+err.Error())
			return
		}
		out, err = json.MarshalIndent(ldsResponse{Listeners: listeners}, " ", " ")
		if err != nil {
			errorResponse(methodName, response, http.StatusInternalServerError, "LDS "+err.Error())
			return
		}

		transformedOutput, err = ds.invokeWebhook(request.Request.URL.Path, out, "webhook"+methodName)
		if err != nil {
			// Use whatever we generated.
			log.Errorf("error invoking webhook: %v", err)
			transformedOutput = out
		}

		// TODO: This does not take into account listeners added by webhook
		resourceCount = uint32(len(listeners))
		// TODO: Bug. If resourceCount is 0 but transformedOutput adds listeners, cache wont update
		if resourceCount > 0 {
			ds.ldsCache.updateCachedDiscoveryResponse(key, resourceCount, transformedOutput)
		}
	}
	observeResources(methodName, resourceCount)
	writeResponse(response, transformedOutput)
}

// ListRoutes responds to RDS requests, used by HTTP routes
// Routes correspond to HTTP routes and use the listener port as the route name
// to identify HTTP filters in the config. Service node value holds the local proxy identity.
func (ds *DiscoveryService) ListRoutes(request *restful.Request, response *restful.Response) {
	methodName := "ListRoutes"
	incCalls(methodName)

	key := request.Request.URL.String()
	out, resourceCount, cached := ds.rdsCache.cachedDiscoveryResponse(key)
	transformedOutput := out
	if !cached {
		svcNode, err := ds.parseDiscoveryRequest(request)
		if err != nil {
			errorResponse(methodName, response, http.StatusNotFound, "RDS "+err.Error())
			return
		}

		routeConfigName := request.PathParameter(RouteConfigName)
		routeConfig, err := buildRDSRoute(ds.Mesh, svcNode, routeConfigName,
			ds.ServiceDiscovery, ds.IstioConfigStore)
		if err != nil {
			// If client experiences an error, 503 error will tell envoy to keep its current
			// cache and try again later
			errorResponse(methodName, response, http.StatusServiceUnavailable, "RDS "+err.Error())
			return
		}
		if out, err = json.MarshalIndent(routeConfig, " ", " "); err != nil {
			errorResponse(methodName, response, http.StatusInternalServerError, "RDS "+err.Error())
			return
		}

		transformedOutput, err = ds.invokeWebhook(request.Request.URL.Path, out, "webhook"+methodName)
		if err != nil {
			// Use whatever we generated.
			transformedOutput = out
		}

		if routeConfig != nil && routeConfig.VirtualHosts != nil { //TODO: fix same bug as above.
			resourceCount = uint32(len(routeConfig.VirtualHosts))
			if resourceCount > 0 {
				ds.rdsCache.updateCachedDiscoveryResponse(key, resourceCount, transformedOutput)
			}
		}
	}
	observeResources(methodName, resourceCount)
	writeResponse(response, transformedOutput)
}

func (ds *DiscoveryService) invokeWebhook(path string, payload []byte, methodName string) ([]byte, error) {
	if ds.webhookClient == nil {
		return payload, nil
	}

	incWebhookCalls(methodName)
	resp, err := ds.webhookClient.Post(ds.webhookEndpoint+path, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		incWebhookErrors(methodName)
		return nil, err
	}

	defer resp.Body.Close() // nolint: errcheck

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		incWebhookErrors(methodName)
	}

	return out, err
}

func incCalls(methodName string) {
	callCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func incErrors(methodName string) {
	errorCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func incWebhookCalls(methodName string) {
	webhookCallCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func incWebhookErrors(methodName string) {
	webhookErrorCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Inc()
}

func observeResources(methodName string, count uint32) {
	resourceCounter.With(prometheus.Labels{
		metricLabelMethod:  methodName,
		metricBuildVersion: buildVersion,
	}).Observe(float64(count))
}

func errorResponse(methodName string, r *restful.Response, status int, msg string) {
	incErrors(methodName)
	log.Warn(msg)
	if err := r.WriteErrorString(status, msg); err != nil {
		log.Warna(err)
	}
}

func writeResponse(r *restful.Response, data []byte) {
	r.WriteHeader(http.StatusOK)
	if _, err := r.Write(data); err != nil {
		log.Warna(err)
	}
}
