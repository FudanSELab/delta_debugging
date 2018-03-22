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

package bootstrap

import (
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"code.cloudfoundry.org/copilot"
	"github.com/davecgh/go-spew/spew"
	multierror "github.com/hashicorp/go-multierror"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	durpb "github.com/golang/protobuf/ptypes/duration"
	"k8s.io/client-go/kubernetes"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/cmd"
	configaggregate "istio.io/istio/pilot/pkg/config/aggregate"
	"istio.io/istio/pilot/pkg/config/clusterregistry"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/config/kube/crd/file"
	"istio.io/istio/pilot/pkg/config/kube/ingress"
	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pilot/pkg/kube/admit"
	"istio.io/istio/pilot/pkg/model"
	envoy "istio.io/istio/pilot/pkg/proxy/envoy/v1"
	"istio.io/istio/pilot/pkg/proxy/envoy/v1/mock"
	"istio.io/istio/pilot/pkg/serviceregistry"
	"istio.io/istio/pilot/pkg/serviceregistry/aggregate"
	"istio.io/istio/pilot/pkg/serviceregistry/cloudfoundry"
	"istio.io/istio/pilot/pkg/serviceregistry/consul"
	"istio.io/istio/pilot/pkg/serviceregistry/eureka"
	"istio.io/istio/pilot/pkg/serviceregistry/kube"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/version"
)

// ServiceRegistry is an expansion of the platform.ServiceRegistry enum that adds a mock registry.
type ServiceRegistry string

const (
	// MockRegistry environment flag
	MockRegistry ServiceRegistry = "Mock"
	// KubernetesRegistry environment flag
	KubernetesRegistry ServiceRegistry = "Kubernetes"
	// ConsulRegistry environment flag
	ConsulRegistry ServiceRegistry = "Consul"
	// EurekaRegistry environment flag
	EurekaRegistry ServiceRegistry = "Eureka"
	// CloudFoundryRegistry environment flag
	CloudFoundryRegistry ServiceRegistry = "CloudFoundry"
)

var (
	// TODO: use model.IstioConfigTypes once model.IngressRule is deprecated
	configDescriptor = model.ConfigDescriptor{
		model.RouteRule,
		model.V1alpha2RouteRule,
		model.Gateway,
		model.EgressRule,
		model.ExternalService,
		model.DestinationPolicy,
		model.DestinationRule,
		model.HTTPAPISpec,
		model.HTTPAPISpecBinding,
		model.QuotaSpec,
		model.QuotaSpecBinding,
		model.EndUserAuthenticationPolicySpec,
		model.EndUserAuthenticationPolicySpecBinding,
	}
)

// MeshArgs provide configuration options for the mesh. If ConfigFile is provided, an attempt will be made to
// load the mesh from the file. Otherwise, a default mesh will be used with optional overrides.
type MeshArgs struct {
	ConfigFile      string
	MixerAddress    string
	RdsRefreshDelay *durpb.Duration
}

// ConfigArgs provide configuration options for the configuration controller. If FileDir is set, that directory will
// be monitored for CRD yaml files and will update the controller as those files change (This is used for testing
// purposes). Otherwise, a CRD client is created based on the configuration.
type ConfigArgs struct {
	ClusterRegistriesDir string
	KubeConfig           string
	CFConfig             string
	ControllerOptions    kube.ControllerOptions
	FileDir              string
}

// ConsulArgs provides configuration for the Consul service registry.
type ConsulArgs struct {
	Config    string
	ServerURL string
	Interval  time.Duration
}

// EurekaArgs provides configuration for the Eureka service registry
type EurekaArgs struct {
	ServerURL string
	Interval  time.Duration
}

// ServiceArgs provides the composite configuration for all service registries in the system.
type ServiceArgs struct {
	Registries []string
	Consul     ConsulArgs
	Eureka     EurekaArgs
}

// AdmissionArgs provides configuration options for the admission controller. This is a partial duplicate of
// admit.ControllerOptions (other fields are filled out before constructing the admission controller). Only
// used if running with k8s, Consul, or Eureka (not in a mock environment).
type AdmissionArgs struct {
	// ExternalAdmissionWebhookName is the name of the
	// ExternalAdmissionHook which describes he external admission
	// webhook and resources and operations it applies to.
	ExternalAdmissionWebhookName string

	// ServiceName is the service name of the webhook.
	ServiceName string

	// SecretName is the name of k8s secret that contains the webhook
	// server key/cert and corresponding CA cert that signed them. The
	// server key/cert are used to serve the webhook and the CA cert
	// is provided to k8s apiserver during admission controller
	// registration.
	SecretName string

	// Port where the webhook is served. Per k8s admission
	// registration requirements this should be 443 unless there is
	// only a single port for the service.
	Port int

	// RegistrationDelay controls how long admission registration
	// occurs after the webhook is started. This is used to avoid
	// potential races where registration completes and k8s apiserver
	// invokes the webhook before the HTTP server is started.
	RegistrationDelay time.Duration
}

// PilotArgs provides all of the configuration parameters for the Pilot discovery service.
type PilotArgs struct {
	DiscoveryOptions envoy.DiscoveryServiceOptions
	Namespace        string
	Mesh             MeshArgs
	Config           ConfigArgs
	Service          ServiceArgs
	Admission        AdmissionArgs
}

// Server contains the runtime configuration for the Pilot discovery service.
type Server struct {
	mesh              *meshconfig.MeshConfig
	serviceController *aggregate.Controller
	configController  model.ConfigStoreCache
	mixerSAN          []string
	kubeClient        kubernetes.Interface
	startFuncs        []startFunc
	listeningAddr     net.Addr
	clusterStore      *clusterregistry.ClusterStore
}

// NewServer creates a new Server instance based on the provided arguments.
func NewServer(args PilotArgs) (*Server, error) {
	// If the namespace isn't set, try looking it up from the environment.
	if args.Namespace == "" {
		args.Namespace = os.Getenv("POD_NAMESPACE")
	}

	s := &Server{}

	// Apply the arguments to the configuration.
	if err := s.initMonitor(&args); err != nil {
		return nil, err
	}
	if err := s.initMesh(&args); err != nil {
		return nil, err
	}
	if err := s.initClusterRegistries(&args); err != nil {
		return nil, err
	}
	if err := s.initKubeClient(&args); err != nil {
		return nil, err
	}
	if err := s.initAdmissionController(&args); err != nil {
		return nil, err
	}
	if err := s.initMixerSan(&args); err != nil {
		return nil, err
	}
	if err := s.initConfigController(&args); err != nil {
		return nil, err
	}
	if err := s.initServiceControllers(&args); err != nil {
		return nil, err
	}
	if err := s.initDiscoveryService(&args); err != nil {
		return nil, err
	}
	return s, nil
}

// Start starts all components of the Pilot discovery service on the port specified in DiscoveryServiceOptions.
// If Port == 0, a port number is automatically chosen. This method returns the address on which the server is
// listening for incoming connections. Content serving is started by this method, but is executed asynchronously.
// Serving can be cancelled at any time by closing the provided stop channel.
func (s *Server) Start(stop chan struct{}) (net.Addr, error) {
	// Now start all of the components.
	for _, fn := range s.startFuncs {
		if err := fn(stop); err != nil {
			return nil, err
		}
	}

	return s.listeningAddr, nil
}

// startFunc defines a function that will be used to start one or more components of the Pilot discovery service.
type startFunc func(stop chan struct{}) error

// initMonitor initializes the configuration for the pilot monitoring server.
func (s *Server) initMonitor(args *PilotArgs) error {
	s.addStartFunc(func(stop chan struct{}) error {
		monitor, err := startMonitor(args.DiscoveryOptions.MonitoringPort)
		if err != nil {
			return err
		}

		go func() {
			<-stop
			err := monitor.Close()
			log.Debugf("Monitoring server terminated: %v", err)
		}()
		return nil
	})
	return nil
}

func (s *Server) initClusterRegistries(args *PilotArgs) (err error) {
	if args.Config.ClusterRegistriesDir != "" {
		s.clusterStore, err = clusterregistry.ReadClusters(args.Config.ClusterRegistriesDir)
		if s.clusterStore != nil {
			log.Infof("clusters configuration %s", spew.Sdump(s.clusterStore))
		}
	}
	return err
}

// initMesh creates the mesh in the pilotConfig from the input arguments.
func (s *Server) initMesh(args *PilotArgs) error {
	// If a config file was specified, use it.
	var mesh *meshconfig.MeshConfig
	if args.Mesh.ConfigFile != "" {
		fileMesh, err := cmd.ReadMeshConfig(args.Mesh.ConfigFile)
		if err != nil {
			log.Warnf("failed to read mesh configuration, using default: %v", err)
		} else {
			mesh = fileMesh
		}
	}

	if mesh == nil {
		// Config file either wasn't specified or failed to load - use a default mesh.
		defaultMesh := model.DefaultMeshConfig()
		mesh = &defaultMesh

		// Allow some overrides for testing purposes.
		if args.Mesh.MixerAddress != "" {
			mesh.MixerCheckServer = args.Mesh.MixerAddress
			mesh.MixerReportServer = args.Mesh.MixerAddress
		}
		if args.Mesh.RdsRefreshDelay != nil {
			mesh.RdsRefreshDelay = args.Mesh.RdsRefreshDelay
		}
	}

	log.Infof("mesh configuration %s", spew.Sdump(mesh))
	log.Infof("version %s", version.Info.String())
	log.Infof("flags %s", spew.Sdump(args))

	s.mesh = mesh
	return nil
}

// initMixerSan configures the mixerSAN configuration item. The mesh must already have been configured.
func (s *Server) initMixerSan(args *PilotArgs) error {
	if s.mesh == nil {
		return fmt.Errorf("the mesh has not been configured before configuring mixer san")
	}
	if s.mesh.DefaultConfig.ControlPlaneAuthPolicy == meshconfig.AuthenticationPolicy_MUTUAL_TLS {
		s.mixerSAN = envoy.GetMixerSAN(args.Config.ControllerOptions.DomainSuffix, args.Namespace)
	}
	return nil
}

func (s *Server) getKubeCfgFile(args *PilotArgs) (kubeCfgFile string) {
	// If the cluster store is configured, get pilot's kubeconfig from there
	if s.clusterStore != nil {
		if kubeCfgFile = s.clusterStore.GetPilotAccessConfig(); kubeCfgFile != "" {
			kubeCfgFile = path.Join(args.Config.ClusterRegistriesDir, kubeCfgFile)
		}
	}
	if kubeCfgFile == "" {
		kubeCfgFile = args.Config.KubeConfig
	}
	return
}

// initKubeClient creates the k8s client if running in an k8s environment.
func (s *Server) initKubeClient(args *PilotArgs) error {
	needToCreateClient := false
	for _, r := range args.Service.Registries {
		switch ServiceRegistry(r) {
		case KubernetesRegistry:
			needToCreateClient = true
		case ConsulRegistry:
			needToCreateClient = true
		case EurekaRegistry:
			needToCreateClient = true
		}
	}

	if needToCreateClient {
		var client kubernetes.Interface
		var kuberr error

		kubeCfgFile := s.getKubeCfgFile(args)
		_, client, kuberr = kube.CreateInterface(kubeCfgFile)
		if kuberr != nil {
			return multierror.Prefix(kuberr, "failed to connect to Kubernetes API.")
		}
		s.kubeClient = client
	}
	return nil
}

type mockController struct{}

func (c *mockController) AppendServiceHandler(f func(*model.Service, model.Event)) error {
	return nil
}

func (c *mockController) AppendInstanceHandler(f func(*model.ServiceInstance, model.Event)) error {
	return nil
}

func (c *mockController) Run(<-chan struct{}) {}

// initConfigController creates the config controller in the pilotConfig.
func (s *Server) initConfigController(args *PilotArgs) error {
	var configController model.ConfigStoreCache
	if args.Config.FileDir != "" {
		store := memory.Make(configDescriptor)
		configController = memory.NewController(store)
		fileMonitor := file.NewMonitor(configController, args.Config.FileDir, configDescriptor)

		// Defer starting the file monitor until after the service is created.
		s.addStartFunc(func(stop chan struct{}) error {
			fileMonitor.Start(stop)
			return nil
		})
	} else {
		kubeCfgFile := s.getKubeCfgFile(args)
		configClient, err := crd.NewClient(kubeCfgFile, configDescriptor,
			args.Config.ControllerOptions.DomainSuffix)
		if err != nil {
			return multierror.Prefix(err, "failed to open a config client.")
		}

		if err = configClient.RegisterResources(); err != nil {
			return multierror.Prefix(err, "failed to register custom resources.")
		}

		configController = crd.NewController(configClient, args.Config.ControllerOptions)
	}

	// Defer starting the controller until after the service is created.
	s.configController = configController
	s.addStartFunc(func(stop chan struct{}) error {
		go s.configController.Run(stop)
		return nil
	})
	return nil
}

// createK8sServiceControllers creates all the k8s service controllers under this pilot
func (s *Server) createK8sServiceControllers(serviceControllers *aggregate.Controller, args *PilotArgs) (err error) {
	kubectl := kube.NewController(s.kubeClient, args.Config.ControllerOptions)
	serviceControllers.AddRegistry(
		aggregate.Registry{
			Name:             serviceregistry.ServiceRegistry(KubernetesRegistry),
			ServiceDiscovery: kubectl,
			ServiceAccounts:  kubectl,
			Controller:       kubectl,
		})

	// Add clusters under the same pilot
	if s.clusterStore != nil {
		clusters := s.clusterStore.GetPilotClusters()
		for _, cluster := range clusters {
			kubeconfig := clusterregistry.GetClusterAccessConfig(cluster)
			kubeCfgFile := path.Join(args.Config.ClusterRegistriesDir, kubeconfig)
			log.Infof("Cluster name: %s, AccessConfigFile: %s", clusterregistry.GetClusterName(cluster), kubeCfgFile)
			_, client, kuberr := kube.CreateInterface(kubeCfgFile)
			if kuberr != nil {
				err = multierror.Append(err, multierror.Prefix(kuberr, fmt.Sprintf("failed to connect to Access API with accessconfig: %s", kubeCfgFile)))
			}

			kubectl := kube.NewController(client, args.Config.ControllerOptions)
			serviceControllers.AddRegistry(
				aggregate.Registry{
					Name:             serviceregistry.ServiceRegistry(KubernetesRegistry),
					ClusterName:      clusterregistry.GetClusterName(cluster),
					ServiceDiscovery: kubectl,
					ServiceAccounts:  kubectl,
					Controller:       kubectl,
				})
		}
	}
	return
}

// initServiceControllers creates and initializes the service controllers
func (s *Server) initServiceControllers(args *PilotArgs) error {
	serviceControllers := aggregate.NewController()
	registered := make(map[ServiceRegistry]bool)
	for _, r := range args.Service.Registries {
		serviceRegistry := ServiceRegistry(r)
		if _, exists := registered[serviceRegistry]; exists {
			log.Warnf("%s registry specified multiple times.", r)
			continue
		}
		registered[serviceRegistry] = true
		log.Infof("Adding %s registry adapter", serviceRegistry)
		switch serviceRegistry {
		case MockRegistry:
			discovery1 := mock.NewDiscovery(
				map[string]*model.Service{
					mock.HelloService.Hostname: mock.HelloService,
				}, 2)

			discovery2 := mock.NewDiscovery(
				map[string]*model.Service{
					mock.WorldService.Hostname: mock.WorldService,
				}, 2)

			registry1 := aggregate.Registry{
				Name:             serviceregistry.ServiceRegistry("mockAdapter1"),
				ServiceDiscovery: discovery1,
				ServiceAccounts:  discovery1,
				Controller:       &mockController{},
			}

			registry2 := aggregate.Registry{
				Name:             serviceregistry.ServiceRegistry("mockAdapter2"),
				ServiceDiscovery: discovery2,
				ServiceAccounts:  discovery2,
				Controller:       &mockController{},
			}
			serviceControllers.AddRegistry(registry1)
			serviceControllers.AddRegistry(registry2)
		case KubernetesRegistry:
			if err := s.createK8sServiceControllers(serviceControllers, args); err != nil {
				return err
			}
			if s.mesh.IngressControllerMode != meshconfig.MeshConfig_OFF {
				// Wrap the config controller with a cache.
				configController, err := configaggregate.MakeCache([]model.ConfigStoreCache{
					s.configController,
					ingress.NewController(s.kubeClient, s.mesh, args.Config.ControllerOptions),
				})
				if err != nil {
					return err
				}

				// Update the config controller
				s.configController = configController

				if ingressSyncer, errSyncer := ingress.NewStatusSyncer(s.mesh, s.kubeClient,
					args.Namespace, args.Config.ControllerOptions); errSyncer != nil {
					log.Warnf("Disabled ingress status syncer due to %v", errSyncer)
				} else {
					s.addStartFunc(func(stop chan struct{}) error {
						go ingressSyncer.Run(stop)
						return nil
					})
				}
			}
		case ConsulRegistry:
			log.Infof("Consul url: %v", args.Service.Consul.ServerURL)
			conctl, conerr := consul.NewController(
				args.Service.Consul.ServerURL, args.Service.Consul.Interval)
			if conerr != nil {
				return fmt.Errorf("failed to create Consul controller: %v", conerr)
			}
			serviceControllers.AddRegistry(
				aggregate.Registry{
					Name:             serviceregistry.ServiceRegistry(r),
					ServiceDiscovery: conctl,
					ServiceAccounts:  conctl,
					Controller:       conctl,
				})
		case EurekaRegistry:
			log.Infof("Eureka url: %v", args.Service.Eureka.ServerURL)
			eurekaClient := eureka.NewClient(args.Service.Eureka.ServerURL)
			serviceControllers.AddRegistry(
				aggregate.Registry{
					Name:             serviceregistry.ServiceRegistry(r),
					Controller:       eureka.NewController(eurekaClient, args.Service.Eureka.Interval),
					ServiceDiscovery: eureka.NewServiceDiscovery(eurekaClient),
					ServiceAccounts:  eureka.NewServiceAccounts(),
				})

		case CloudFoundryRegistry:
			cfConfig, err := cloudfoundry.LoadConfig(args.Config.CFConfig)
			if err != nil {
				return multierror.Prefix(err, "loading cloud foundry config")
			}
			tlsConfig, err := cfConfig.ClientTLSConfig()
			if err != nil {
				return multierror.Prefix(err, "creating cloud foundry client tls config")
			}
			client, err := copilot.NewIstioClient(cfConfig.Copilot.Address, tlsConfig)
			if err != nil {
				return multierror.Prefix(err, "creating cloud foundry client")
			}
			serviceControllers.AddRegistry(aggregate.Registry{
				Name: serviceregistry.ServiceRegistry(r),
				Controller: &cloudfoundry.Controller{
					Ticker: cloudfoundry.NewTicker(cfConfig.Copilot.PollInterval),
					Client: client,
				},
				ServiceDiscovery: &cloudfoundry.ServiceDiscovery{
					Client:      client,
					ServicePort: cfConfig.ServicePort,
				},
				ServiceAccounts: cloudfoundry.NewServiceAccounts(),
			})

		default:
			return multierror.Prefix(nil, "Service registry "+r+" is not supported.")
		}
	}

	s.serviceController = serviceControllers

	// Defer running of the service controllers.
	s.addStartFunc(func(stop chan struct{}) error {
		go s.serviceController.Run(stop)
		return nil
	})

	return nil
}

func (s *Server) initDiscoveryService(args *PilotArgs) error {
	environment := model.Environment{
		Mesh:             s.mesh,
		IstioConfigStore: model.MakeIstioStore(s.configController),
		ServiceDiscovery: s.serviceController,
		ServiceAccounts:  s.serviceController,
		MixerSAN:         s.mixerSAN,
	}

	// Set up discovery service
	discovery, err := envoy.NewDiscoveryService(
		s.serviceController,
		s.configController,
		environment,
		args.DiscoveryOptions)
	if err != nil {
		return fmt.Errorf("failed to create discovery service: %v", err)
	}

	s.addStartFunc(func(stop chan struct{}) error {
		addr, err := discovery.Start(stop)
		if err == nil {
			// Store the listening address in the output config.
			s.listeningAddr = addr
		}
		return err
	})

	return nil
}

// initAdmissionController creates and initializes the k8s admission controller if running in a k8s environment.
func (s *Server) initAdmissionController(args *PilotArgs) error {
	if s.kubeClient == nil {
		// Not running in a k8s environment - do nothing.
		return nil
	}

	// Create the arguments for the admission controller
	admissionArgs := admit.ControllerOptions{
		ExternalAdmissionWebhookName: args.Admission.ExternalAdmissionWebhookName,
		ServiceName:                  args.Admission.ServiceName,
		SecretName:                   args.Admission.SecretName,
		Port:                         args.Admission.Port,
		RegistrationDelay:            args.Admission.RegistrationDelay,
		Descriptor:                   configDescriptor,
		ServiceNamespace:             args.Namespace,
		DomainSuffix:                 args.Config.ControllerOptions.DomainSuffix,
		ValidateNamespaces: []string{
			args.Config.ControllerOptions.WatchedNamespace,
			args.Namespace,
		},
	}

	admissionController, err := admit.NewController(s.kubeClient, admissionArgs)
	if err != nil {
		return err
	}

	// Defer running the admission controller.
	s.addStartFunc(func(stop chan struct{}) error {
		go admissionController.Run(stop)
		return nil
	})
	return nil
}

func (s *Server) addStartFunc(fn startFunc) {
	s.startFuncs = append(s.startFuncs, fn)
}
