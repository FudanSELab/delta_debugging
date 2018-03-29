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
	"sort"
	"strings"
	"time"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
)

const (
	// DefaultAccessLog is the name of the log channel (stdout in docker environment)
	DefaultAccessLog = "/dev/stdout"

	// DefaultLbType defines the default load balancer policy
	DefaultLbType = LbTypeRoundRobin

	// LDSName is the name of listener-discovery-service (LDS) cluster
	LDSName = "lds"

	// RDSName is the name of route-discovery-service (RDS) cluster
	RDSName = "rds"

	// SDSName is the name of service-discovery-service (SDS) cluster
	SDSName = "sds"

	// CDSName is the name of cluster-discovery-service (CDS) cluster
	CDSName = "cds"

	// RDSAll is the special name for HTTP PROXY route
	RDSAll = "http_proxy"

	// VirtualListenerName is the name for traffic capture listener
	VirtualListenerName = "virtual"

	// ClusterTypeStrictDNS name for clusters of type 'strict_dns'
	ClusterTypeStrictDNS = "strict_dns"

	// ClusterTypeStatic name for clusters of type 'static'
	ClusterTypeStatic = "static"

	// ClusterTypeOriginalDST name for clusters of type 'original_dst'
	ClusterTypeOriginalDST = "original_dst"

	// ClusterTypeSDS name for clusters of type 'sds'
	ClusterTypeSDS = "sds"

	// LbTypeRoundRobin is the name for round-robin LB
	LbTypeRoundRobin = "round_robin"

	// LbTypeLeastRequest is the name for least request LB
	LbTypeLeastRequest = "least_request"

	// LbTypeRingHash is the name for ring hash LB
	LbTypeRingHash = "ring_hash"

	// LbTypeRandom is the name for random LB
	LbTypeRandom = "random"

	// LbTypeOriginalDST is the name for LB of original_dst
	LbTypeOriginalDST = "original_dst_lb"

	// ClusterFeatureHTTP2 is the feature to use HTTP/2 for a cluster
	ClusterFeatureHTTP2 = "http2"

	// HTTPConnectionManager is the name of HTTP filter.
	HTTPConnectionManager = "http_connection_manager"

	// TCPProxyFilter is the name of the TCP Proxy network filter.
	TCPProxyFilter = "tcp_proxy"

	// CORSFilter is the name of the CORS network filter
	CORSFilter = "cors"

	// MongoProxyFilter is the name of the Mongo Proxy network filter.
	MongoProxyFilter = "mongo_proxy"

	// RedisProxyFilter is the name of the Redis Proxy network filter.
	RedisProxyFilter = "redis_proxy"

	// RedisDefaultOpTimeout is the op timeout used for Redis Proxy filter
	// Currently it is set to 30s (conversion happens in the filter)
	// TODO - Allow this to be configured.
	RedisDefaultOpTimeout = 30 * time.Second

	// WildcardAddress binds to all IP addresses
	WildcardAddress = "0.0.0.0"

	// LocalhostAddress for local binding
	LocalhostAddress = "127.0.0.1"

	// EgressTraceOperation denotes the name of trace operation for Envoy
	EgressTraceOperation = "egress"

	// IngressTraceOperation denotes the name of trace operation for Envoy
	IngressTraceOperation = "ingress"

	// ZipkinTraceDriverType denotes the Zipkin HTTP trace driver
	ZipkinTraceDriverType = "zipkin"

	// ZipkinCollectorCluster denotes the cluster where zipkin server is running
	ZipkinCollectorCluster = "zipkin"

	// ZipkinCollectorEndpoint denotes the REST endpoint where Envoy posts Zipkin spans
	ZipkinCollectorEndpoint = "/api/v1/spans"

	// MaxClusterNameLength is the maximum cluster name length
	MaxClusterNameLength = 189 // TODO: use MeshConfig.StatNameLength instead

	// headers with special meaning in Envoy
	headerMethod    = ":method"
	headerAuthority = ":authority"
	headerScheme    = ":scheme"

	router  = "router"
	auto    = "auto"
	decoder = "decoder"
	read    = "read"
	both    = "both"
)

// ListenersALPNProtocols denotes the the list of ALPN protocols that the listener
// should expose
var ListenersALPNProtocols = []string{"h2", "http/1.1"}

// convertDuration converts to golang duration and logs errors
func convertDuration(d *duration.Duration) time.Duration {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - convertDuration()")

	if d == nil {
		return 0
	}
	dur, err := ptypes.Duration(d)
	if err != nil {
		log.Warnf("error converting duration %#v, using 0: %v", d, err)
	}
	return dur
}

func protoDurationToMS(dur *duration.Duration) int64 {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - protoDurationToMS()")

	return int64(convertDuration(dur) / time.Millisecond)
}

// Config defines the schema for Envoy JSON configuration format
type Config struct {
	RootRuntime        *RootRuntime   `json:"runtime,omitempty"`
	Listeners          Listeners      `json:"listeners"`
	LDS                *LDSCluster    `json:"lds,omitempty"`
	Admin              Admin          `json:"admin"`
	ClusterManager     ClusterManager `json:"cluster_manager"`
	StatsdUDPIPAddress string         `json:"statsd_udp_ip_address,omitempty"`
	Tracing            *Tracing       `json:"tracing,omitempty"`

	// Special value used to hash all referenced values (e.g. TLS secrets)
	Hash []byte `json:"-"`
}

// Tracing definition
type Tracing struct {
	HTTPTracer HTTPTracer `json:"http"`
}

// HTTPTracer definition
type HTTPTracer struct {
	HTTPTraceDriver HTTPTraceDriver `json:"driver"`
}

// HTTPTraceDriver definition
type HTTPTraceDriver struct {
	HTTPTraceDriverType   string                `json:"type"`
	HTTPTraceDriverConfig HTTPTraceDriverConfig `json:"config"`
}

// HTTPTraceDriverConfig definition
type HTTPTraceDriverConfig struct {
	CollectorCluster  string `json:"collector_cluster"`
	CollectorEndpoint string `json:"collector_endpoint"`
}

// RootRuntime definition.
// See https://envoyproxy.github.io/envoy/configuration/overview/overview.html
type RootRuntime struct {
	SymlinkRoot          string `json:"symlink_root"`
	Subdirectory         string `json:"subdirectory"`
	OverrideSubdirectory string `json:"override_subdirectory,omitempty"`
}

// AbortFilter definition
type AbortFilter struct {
	Percent    int `json:"abort_percent,omitempty"`
	HTTPStatus int `json:"http_status,omitempty"`
}

// DelayFilter definition
type DelayFilter struct {
	Type     string `json:"type,omitempty"`
	Percent  int    `json:"fixed_delay_percent,omitempty"`
	Duration int64  `json:"fixed_duration_ms,omitempty"`
}

// AppendedHeader definition
type AppendedHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Header definition
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Regex bool   `json:"regex,omitempty"`
}

// FilterFaultConfig definition
type FilterFaultConfig struct {
	Abort           *AbortFilter `json:"abort,omitempty"`
	Delay           *DelayFilter `json:"delay,omitempty"`
	Headers         Headers      `json:"headers,omitempty"`
	UpstreamCluster string       `json:"upstream_cluster,omitempty"`
}

// FilterRouterConfig definition
type FilterRouterConfig struct {
	// DynamicStats defaults to true
	DynamicStats bool `json:"dynamic_stats,omitempty"`
}

// HTTPFilter definition
type HTTPFilter struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

// Runtime definition
type Runtime struct {
	Key     string `json:"key"`
	Default int    `json:"default"`
}

// Decorator definition
type Decorator struct {
	Operation string `json:"operation"`
}

// HTTPRoute definition
type HTTPRoute struct {
	Runtime *Runtime `json:"runtime,omitempty"`

	Path   string `json:"path,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regex  string `json:"regex,omitempty"`

	PrefixRewrite string `json:"prefix_rewrite,omitempty"`
	HostRewrite   string `json:"host_rewrite,omitempty"`

	PathRedirect string `json:"path_redirect,omitempty"`
	HostRedirect string `json:"host_redirect,omitempty"`

	Cluster          string           `json:"cluster,omitempty"`
	WeightedClusters *WeightedCluster `json:"weighted_clusters,omitempty"`

	Headers      Headers           `json:"headers,omitempty"`
	TimeoutMS    int64             `json:"timeout_ms"`
	RetryPolicy  *RetryPolicy      `json:"retry_policy,omitempty"`
	OpaqueConfig map[string]string `json:"opaque_config,omitempty"`

	AutoHostRewrite  bool `json:"auto_host_rewrite,omitempty"`
	WebsocketUpgrade bool `json:"use_websocket,omitempty"`

	ShadowCluster *ShadowCluster `json:"shadow,omitempty"`

	HeadersToAdd []AppendedHeader `json:"request_headers_to_add,omitempty"`

	CORSPolicy *CORSPolicy `json:"cors,omitempty"`

	Decorator *Decorator `json:"decorator,omitempty"`

	// clusters contains the set of referenced clusters in the route; the field is special
	// and used only to aggregate cluster information after composing routes
	clusters Clusters

	// faults contains the set of referenced faults in the route; the field is special
	// and used only to aggregate fault filter information after composing routes
	faults []*HTTPFilter
}

// Redirect returns true if route contains redirect logic
func (route *HTTPRoute) Redirect() bool {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Redirect()")

	return route.HostRedirect != "" || route.PathRedirect != ""
}

// CatchAll returns true if the route matches all requests
func (route *HTTPRoute) CatchAll() bool {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - CatchAll()")

	return len(route.Headers) == 0 && route.Path == "" && route.Prefix == "/"
}

// CombinePathPrefix checks that the route applies for a given path and prefix
// match and updates the path and the prefix in the route. If the route is
// incompatible with the path or the prefix, returns nil.  Either path or
// prefix must be set but not both.  The resulting route must match exactly the
// requests that match both the original route and the supplied path and
// prefix.
func (route *HTTPRoute) CombinePathPrefix(path, prefix string) *HTTPRoute {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - CombinePathPrefix()")

	switch {
	case path == "" && route.Path == "" && strings.HasPrefix(route.Prefix, prefix):
		// pick the longest prefix if both are prefix matches
		return route
	case path == "" && route.Path == "" && strings.HasPrefix(prefix, route.Prefix):
		route.Prefix = prefix
		return route
	case prefix == "" && route.Prefix == "" && route.Path == path:
		// pick only if path matches if both are path matches
		return route
	case path == "" && route.Prefix == "" && strings.HasPrefix(route.Path, prefix):
		// if mixed, pick if route path satisfies the prefix
		return route
	case prefix == "" && route.Path == "" && strings.HasPrefix(path, route.Prefix):
		// if mixed, pick if route prefix satisfies the path and change route to path
		route.Path = path
		route.Prefix = ""
		return route
	default:
		return nil
	}
}

// CORSPolicy definition
// See: https://www.envoyproxy.io/envoy/configuration/http_filters/cors_filter.html#config-http-filters-cors
type CORSPolicy struct {
	Enabled          bool     `json:"enabled,omitempty"`
	AllowCredentials bool     `json:"allow_credentials,omitempty"`
	AllowMethods     string   `json:"allow_methods,omitempty"`
	AllowHeaders     string   `json:"allow_headers,omitempty"`
	ExposeHeaders    string   `json:"expose_headers,omitempty"`
	MaxAge           string   `json:"max_age,omitempty"`
	AllowOrigin      []string `json:"allow_origin,omitempty"`
}

// RetryPolicy definition
// See: https://lyft.github.io/envoy/docs/configuration/http_conn_man/route_config/route.html#retry-policy
type RetryPolicy struct {
	Policy          string `json:"retry_on"` //if unset, set to 5xx,connect-failure,refused-stream
	NumRetries      int    `json:"num_retries,omitempty"`
	PerTryTimeoutMS int64  `json:"per_try_timeout_ms,omitempty"`
}

// ShadowCluster definition
// See: https://www.envoyproxy.io/envoy/configuration/http_conn_man/route_config/route.html?
// highlight=shadow#config-http-conn-man-route-table-route-shadow
type ShadowCluster struct {
	Cluster string `json:"cluster"`
}

// WeightedCluster definition
// See https://envoyproxy.github.io/envoy/configuration/http_conn_man/route_config/route.html
type WeightedCluster struct {
	Clusters         []*WeightedClusterEntry `json:"clusters"`
	RuntimeKeyPrefix string                  `json:"runtime_key_prefix,omitempty"`
}

// WeightedClusterEntry definition. Describes the format of each entry in the WeightedCluster
type WeightedClusterEntry struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

// VirtualHost definition
type VirtualHost struct {
	Name    string       `json:"name"`
	Domains []string     `json:"domains"`
	Routes  []*HTTPRoute `json:"routes"`
}

func (host *VirtualHost) clusters() Clusters {
	out := make(Clusters, 0)
	for _, route := range host.Routes {
		out = append(out, route.clusters...)
	}
	return out
}

// HTTPRouteConfig definition
type HTTPRouteConfig struct {
	VirtualHosts []*VirtualHost `json:"virtual_hosts"`
}

// HTTPRouteConfigs is a map from the port number to the route config
type HTTPRouteConfigs map[int]*HTTPRouteConfig

// EnsurePort creates a route config if necessary
func (routes HTTPRouteConfigs) EnsurePort(port int) *HTTPRouteConfig {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - EnsurePort()")

	config, ok := routes[port]
	if !ok {
		config = &HTTPRouteConfig{}
		routes[port] = config
	}
	return config
}

func (routes HTTPRouteConfigs) clusters() Clusters {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - clusters()")

	out := make(Clusters, 0)
	for _, config := range routes {
		out = append(out, config.clusters()...)
	}
	return out
}

func (routes HTTPRouteConfigs) normalize() HTTPRouteConfigs {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - normalize()")

	out := make(HTTPRouteConfigs)

	// sort HTTP routes by virtual hosts, rest should be deterministic
	for port, routeConfig := range routes {
		out[port] = routeConfig.normalize()
	}

	return out
}

// combine creates a new route config that is the union of all HTTP routes.
// note that the virtual hosts without an explicit port suffix (IP:PORT) are stripped
// for all routes except the route for port 80.
func (routes HTTPRouteConfigs) combine() *HTTPRouteConfig {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - combine()")

	out := &HTTPRouteConfig{}
	for port, config := range routes {
		for _, host := range config.VirtualHosts {
			vhost := &VirtualHost{
				Name:   host.Name,
				Routes: host.Routes,
			}
			for _, domain := range host.Domains {
				if port == 80 || strings.Contains(domain, ":") {
					vhost.Domains = append(vhost.Domains, domain)
				}
			}

			if len(vhost.Domains) > 0 {
				out.VirtualHosts = append(out.VirtualHosts, vhost)
			}
		}
	}
	return out.normalize()
}

// faults aggregates fault filters across virtual hosts in single http_conn_man
func (rc *HTTPRouteConfig) faults() []*HTTPFilter {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - faults()")

	out := make([]*HTTPFilter, 0)
	for _, host := range rc.VirtualHosts {
		for _, route := range host.Routes {
			out = append(out, route.faults...)
		}
	}
	return out
}

func (rc *HTTPRouteConfig) clusters() Clusters {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - clusters()")

	out := make(Clusters, 0)
	for _, host := range rc.VirtualHosts {
		out = append(out, host.clusters()...)
	}
	return out
}

func (rc *HTTPRouteConfig) normalize() *HTTPRouteConfig {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - normalize()")

	hosts := make([]*VirtualHost, len(rc.VirtualHosts))
	copy(hosts, rc.VirtualHosts)
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].Name < hosts[j].Name })
	return &HTTPRouteConfig{VirtualHosts: hosts}
}

// AccessLog definition.
type AccessLog struct {
	Path   string `json:"path"`
	Format string `json:"format,omitempty"`
	Filter string `json:"filter,omitempty"`
}

// HTTPFilterConfig definition
type HTTPFilterConfig struct {
	CodecType         string                 `json:"codec_type"`
	StatPrefix        string                 `json:"stat_prefix"`
	GenerateRequestID bool                   `json:"generate_request_id,omitempty"`
	UseRemoteAddress  bool                   `json:"use_remote_address,omitempty"`
	Tracing           *HTTPFilterTraceConfig `json:"tracing,omitempty"`
	RouteConfig       *HTTPRouteConfig       `json:"route_config,omitempty"`
	RDS               *RDS                   `json:"rds,omitempty"`
	Filters           []HTTPFilter           `json:"filters"`
	AccessLog         []AccessLog            `json:"access_log"`
}

// IsNetworkFilterConfig marks HTTPFilterConfig as an implementation of NetworkFilterConfig
func (*HTTPFilterConfig) IsNetworkFilterConfig() {}

// HTTPFilterTraceConfig definition
type HTTPFilterTraceConfig struct {
	OperationName string `json:"operation_name"`
}

// TCPRoute definition
type TCPRoute struct {
	Cluster           string   `json:"cluster"`
	DestinationIPList []string `json:"destination_ip_list,omitempty"`
	DestinationPorts  string   `json:"destination_ports,omitempty"`
	SourceIPList      []string `json:"source_ip_list,omitempty"`
	SourcePorts       string   `json:"source_ports,omitempty"`

	// special value to retain dependent cluster definition for TCP routes.
	clusterRef *Cluster
}

// TCPRouteByRoute sorts TCP routes over all route sub fields.
type TCPRouteByRoute []*TCPRoute

func (r TCPRouteByRoute) Len() int {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Len()")

	return len(r)
}

func (r TCPRouteByRoute) Swap(i, j int) {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Swap()")

	r[i], r[j] = r[j], r[i]
}

func (r TCPRouteByRoute) Less(i, j int) bool {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Less()")

	if r[i].Cluster != r[j].Cluster {
		return r[i].Cluster < r[j].Cluster
	}

	compare := func(a, b []string) bool {
		lenA, lenB := len(a), len(b)
		min := lenA
		if min > lenB {
			min = lenB
		}
		for k := 0; k < min; k++ {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return lenA < lenB
	}

	if less := compare(r[i].DestinationIPList, r[j].DestinationIPList); less {
		return less
	}
	if r[i].DestinationPorts != r[j].DestinationPorts {
		return r[i].DestinationPorts < r[j].DestinationPorts
	}
	if less := compare(r[i].SourceIPList, r[j].SourceIPList); less {
		return less
	}
	if r[i].SourcePorts != r[j].SourcePorts {
		return r[i].SourcePorts < r[j].SourcePorts
	}
	return false
}

// TCPProxyFilterConfig definition
type TCPProxyFilterConfig struct {
	StatPrefix  string          `json:"stat_prefix"`
	RouteConfig *TCPRouteConfig `json:"route_config"`
}

// IsNetworkFilterConfig marks TCPProxyFilterConfig as an implementation of NetworkFilterConfig
func (*TCPProxyFilterConfig) IsNetworkFilterConfig() {}

// TCPRouteConfig (or generalize as RouteConfig or L4RouteConfig for TCP/UDP?)
type TCPRouteConfig struct {
	Routes []*TCPRoute `json:"routes"`
}

// MongoProxyFilterConfig definition
type MongoProxyFilterConfig struct {
	StatPrefix string `json:"stat_prefix"`
}

// IsNetworkFilterConfig marks MongoProxyFilterConfig as an implementation of NetworkFilterConfig
func (*MongoProxyFilterConfig) IsNetworkFilterConfig() {}

// CORSFilterConfig definition
// See: https://www.envoyproxy.io/envoy/configuration/http_filters/cors_filter.html#config-http-filters-cors
type CORSFilterConfig struct{}

// IsNetworkFilterConfig marks CORSFilterConfig as an implementation of NetworkFilterConfig
func (*CORSFilterConfig) IsNetworkFilterConfig() {}

// RedisConnPool definition
type RedisConnPool struct {
	OperationTimeoutMS int64 `json:"op_timeout_ms"`
}

// RedisProxyFilterConfig definition
type RedisProxyFilterConfig struct {
	ClusterName string         `json:"cluster_name"`
	ConnPool    *RedisConnPool `json:"conn_pool"`
	StatPrefix  string         `json:"stat_prefix"`
}

// IsNetworkFilterConfig marks RedisProxyFilterConfig as an implementation of NetworkFilterConfig
func (*RedisProxyFilterConfig) IsNetworkFilterConfig() {}

// NetworkFilter definition
type NetworkFilter struct {
	Type   string              `json:"type"`
	Name   string              `json:"name"`
	Config NetworkFilterConfig `json:"config"`
}

// NetworkFilterConfig is a marker interface
type NetworkFilterConfig interface {
	IsNetworkFilterConfig()
}

// Listener definition
type Listener struct {
	Address        string           `json:"address"`
	Name           string           `json:"name,omitempty"`
	Filters        []*NetworkFilter `json:"filters"`
	SSLContext     *SSLContext      `json:"ssl_context,omitempty"`
	BindToPort     bool             `json:"bind_to_port"`
	UseOriginalDst bool             `json:"use_original_dst,omitempty"`
}

// Listeners is a collection of listeners
type Listeners []*Listener

// normalize sorts and de-duplicates listeners by address
func (listeners Listeners) normalize() Listeners {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - normalize()")

	out := make(Listeners, 0, len(listeners))
	set := make(map[string]bool)
	for _, listener := range listeners {
		if !set[listener.Address] {
			set[listener.Address] = true
			out = append(out, listener)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Address < out[j].Address })
	return out
}

// GetByAddress returns a listener by its address
func (listeners Listeners) GetByAddress(addr string) *Listener {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - GetByAddress()")

	for _, listener := range listeners {
		if listener.Address == addr {
			return listener
		}
	}
	return nil
}

// SSLContext definition
type SSLContext struct {
	CertChainFile            string `json:"cert_chain_file"`
	PrivateKeyFile           string `json:"private_key_file"`
	CaCertFile               string `json:"ca_cert_file,omitempty"`
	RequireClientCertificate bool   `json:"require_client_certificate"`
	ALPNProtocols            string `json:"alpn_protocols,omitempty"`
}

// SSLContextExternal definition
type SSLContextExternal struct {
	CaCertFile string `json:"ca_cert_file,omitempty"`
}

// SSLContextWithSAN definition, VerifySubjectAltName cannot be nil.
type SSLContextWithSAN struct {
	CertChainFile        string   `json:"cert_chain_file"`
	PrivateKeyFile       string   `json:"private_key_file"`
	CaCertFile           string   `json:"ca_cert_file,omitempty"`
	VerifySubjectAltName []string `json:"verify_subject_alt_name"`
}

// Admin definition
type Admin struct {
	AccessLogPath string `json:"access_log_path"`
	Address       string `json:"address"`
}

// Host definition
type Host struct {
	URL string `json:"url"`
}

// Cluster definition
type Cluster struct {
	Name                     string            `json:"name"`
	ServiceName              string            `json:"service_name,omitempty"`
	ConnectTimeoutMs         int64             `json:"connect_timeout_ms"`
	Type                     string            `json:"type"`
	LbType                   string            `json:"lb_type"`
	MaxRequestsPerConnection int               `json:"max_requests_per_connection,omitempty"`
	Hosts                    []Host            `json:"hosts,omitempty"`
	SSLContext               interface{}       `json:"ssl_context,omitempty"`
	Features                 string            `json:"features,omitempty"`
	CircuitBreaker           *CircuitBreaker   `json:"circuit_breakers,omitempty"`
	OutlierDetection         *OutlierDetection `json:"outlier_detection,omitempty"`

	// special values used by the post-processing passes for outbound mesh-local clusters
	outbound bool
	hostname string
	port     *model.Port
	labels   model.Labels
}

// CircuitBreaker definition
// See: https://lyft.github.io/envoy/docs/configuration/cluster_manager/cluster_circuit_breakers.html#circuit-breakers
type CircuitBreaker struct {
	Default DefaultCBPriority `json:"default"`
}

// DefaultCBPriority defines the circuit breaker for default cluster priority
type DefaultCBPriority struct {
	MaxConnections     int `json:"max_connections,omitempty"`
	MaxPendingRequests int `json:"max_pending_requests,omitempty"`
	MaxRequests        int `json:"max_requests,omitempty"`
	MaxRetries         int `json:"max_retries,omitempty"`
}

// OutlierDetection definition
// See: https://lyft.github.io/envoy/docs/configuration/cluster_manager/cluster_runtime.html#outlier-detection
type OutlierDetection struct {
	ConsecutiveErrors  int   `json:"consecutive_5xx,omitempty"`
	IntervalMS         int64 `json:"interval_ms,omitempty"`
	BaseEjectionTimeMS int64 `json:"base_ejection_time_ms,omitempty"`
	MaxEjectionPercent int   `json:"max_ejection_percent,omitempty"`
}

// Clusters is a collection of clusters
type Clusters []*Cluster

// normalize deduplicates and sorts clusters by name
func (clusters Clusters) normalize() Clusters {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - normalize()")

	out := make(Clusters, 0, len(clusters))
	set := make(map[string]bool)
	for _, cluster := range clusters {
		if !set[cluster.Name] {
			set[cluster.Name] = true
			out = append(out, cluster)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// RoutesByPath sorts routes by their path and/or prefix, such that:
// - Exact path routes are "less than" than prefix path routes
// - Exact path routes are sorted lexicographically
// - Prefix path routes are sorted anti-lexicographically
//
// This order ensures that prefix path routes do not shadow more
// specific routes which share the same prefix.
type RoutesByPath []*HTTPRoute

func (r RoutesByPath) Len() int {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Len()")

	return len(r)
}

func (r RoutesByPath) Swap(i, j int) {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Swap()")

	r[i], r[j] = r[j], r[i]
}

func (r RoutesByPath) Less(i, j int) bool {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Less()")

	if r[i].Path != "" {
		if r[j].Path != "" {
			// i and j are both path
			return r[i].Path < r[j].Path
		}
		// i is path and j is prefix => i is "less than" j
		return true
	}
	if r[j].Path != "" {
		// i is prefix nad j is path => j is "less than" i
		return false
	}
	// i and j are both prefix
	return r[i].Prefix > r[j].Prefix
}

// Headers sorts headers
type Headers []Header

func (s Headers) Len() int {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Len()")

	return len(s)
}

func (s Headers) Swap(i, j int) {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Swap()")

	s[i], s[j] = s[j], s[i]
}

func (s Headers) Less(i, j int) bool {

	log.Infof("[调试标记 - pilot - pkg - proxy - envoy - v1 - resources.go - Less()")

	if s[i].Name == s[j].Name {
		if s[i].Regex == s[j].Regex {
			return s[i].Value < s[j].Value
		}
		// true is less, false is more
		return s[i].Regex
	}
	return s[i].Name < s[j].Name
}

// DiscoveryCluster is a service discovery service definition
type DiscoveryCluster struct {
	Cluster        *Cluster `json:"cluster"`
	RefreshDelayMs int64    `json:"refresh_delay_ms"`
}

// LDSCluster is a reference to LDS cluster by name
type LDSCluster struct {
	Cluster        string `json:"cluster"`
	RefreshDelayMs int64  `json:"refresh_delay_ms"`
}

// RDS definition
type RDS struct {
	Cluster         string `json:"cluster"`
	RouteConfigName string `json:"route_config_name"`
	RefreshDelayMs  int64  `json:"refresh_delay_ms"`
}

// ClusterManager definition
type ClusterManager struct {
	Clusters Clusters          `json:"clusters"`
	SDS      *DiscoveryCluster `json:"sds,omitempty"`
	CDS      *DiscoveryCluster `json:"cds,omitempty"`
}
