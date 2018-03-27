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

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/spf13/cobra"

	"github.com/spf13/cobra/doc"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/cmd"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/proxy"
	envoy "istio.io/istio/pilot/pkg/proxy/envoy/v1"
	"istio.io/istio/pilot/pkg/serviceregistry"
	"istio.io/istio/pkg/collateral"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/version"
)

var (
	role     model.Proxy
	registry serviceregistry.ServiceRegistry

	// proxy config flags (named identically)
	configPath             string
	binaryPath             string
	serviceCluster         string
	availabilityZone       string
	drainDuration          time.Duration
	parentShutdownDuration time.Duration
	discoveryAddress       string
	discoveryRefreshDelay  time.Duration
	zipkinAddress          string
	connectTimeout         time.Duration
	statsdUDPAddress       string
	proxyAdminPort         int
	controlPlaneAuthPolicy string
	customConfigFile       string
	proxyLogLevel          string
	concurrency            int
	bootstrapv2            bool

	loggingOptions = log.NewOptions()

	rootCmd = &cobra.Command{
		Use:   "pilot-agent",
		Short: "Istio Pilot agent",
		Long:  "Istio Pilot provides management plane functionality to the Istio service mesh and Istio Mixer.",
	}

	proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "Envoy proxy agent",
		RunE: func(c *cobra.Command, args []string) error {
			if err := log.Configure(loggingOptions); err != nil {
				return err
			}
			log.Infof("Version %s", version.Info.String())
			role.Type = model.Sidecar
			if len(args) > 0 {
				role.Type = model.NodeType(args[0])
			}

			// set values from registry platform
			if role.IPAddress == "" {
				if registry == serviceregistry.KubernetesRegistry {
					role.IPAddress = os.Getenv("INSTANCE_IP")
				} else {
					ipAddr := "127.0.0.1"
					if ok := proxy.WaitForPrivateNetwork(); ok {
						ipAddr = proxy.GetPrivateIP().String()
						log.Infof("Obtained private IP %v", ipAddr)
					}

					role.IPAddress = ipAddr
				}
			}
			if role.ID == "" {
				if registry == serviceregistry.KubernetesRegistry {
					role.ID = os.Getenv("POD_NAME") + "." + os.Getenv("POD_NAMESPACE")
				} else if registry == serviceregistry.ConsulRegistry {
					role.ID = role.IPAddress + ".service.consul"
				} else {
					role.ID = role.IPAddress
				}
			}
			pilotDomain := role.Domain
			if role.Domain == "" {
				if registry == serviceregistry.KubernetesRegistry {
					role.Domain = os.Getenv("POD_NAMESPACE") + ".svc.cluster.local"
					pilotDomain = "cluster.local"
				} else if registry == serviceregistry.ConsulRegistry {
					role.Domain = "service.consul"
				} else {
					role.Domain = ""
				}
			}

			log.Infof("Proxy role: %#v", role)

			proxyConfig := meshconfig.ProxyConfig{}

			// set all flags
			proxyConfig.AvailabilityZone = availabilityZone
			proxyConfig.CustomConfigFile = customConfigFile
			proxyConfig.ConfigPath = configPath
			proxyConfig.BinaryPath = binaryPath
			proxyConfig.ServiceCluster = serviceCluster
			proxyConfig.DrainDuration = ptypes.DurationProto(drainDuration)
			proxyConfig.ParentShutdownDuration = ptypes.DurationProto(parentShutdownDuration)
			proxyConfig.DiscoveryAddress = discoveryAddress
			proxyConfig.DiscoveryRefreshDelay = ptypes.DurationProto(discoveryRefreshDelay)
			proxyConfig.ZipkinAddress = zipkinAddress
			proxyConfig.ConnectTimeout = ptypes.DurationProto(connectTimeout)
			proxyConfig.StatsdUdpAddress = statsdUDPAddress
			proxyConfig.ProxyAdminPort = int32(proxyAdminPort)
			proxyConfig.Concurrency = int32(concurrency)

			var pilotSAN []string
			switch controlPlaneAuthPolicy {
			case meshconfig.AuthenticationPolicy_NONE.String():
				proxyConfig.ControlPlaneAuthPolicy = meshconfig.AuthenticationPolicy_NONE
			case meshconfig.AuthenticationPolicy_MUTUAL_TLS.String():
				var ns string
				proxyConfig.ControlPlaneAuthPolicy = meshconfig.AuthenticationPolicy_MUTUAL_TLS
				if registry == serviceregistry.KubernetesRegistry {
					partDiscoveryAddress := strings.Split(discoveryAddress, ":")
					discoveryHostname := partDiscoveryAddress[0]
					parts := strings.Split(discoveryHostname, ".")
					if len(parts) == 1 {
						// namespace of pilot is not part of discovery address use
						// pod namespace e.g. istio-pilot:15003
						ns = os.Getenv("POD_NAMESPACE")
					} else {
						// namespace is found in the discovery address
						// e.g. istio-pilot.istio-system:15003
						ns = parts[1]
					}
				}
				pilotSAN = envoy.GetPilotSAN(pilotDomain, ns)
			}

			// resolve statsd address
			if proxyConfig.StatsdUdpAddress != "" {
				addr, err := proxy.ResolveAddr(proxyConfig.StatsdUdpAddress)
				if err == nil {
					proxyConfig.StatsdUdpAddress = addr
				}
				// If istio-mixer.istio-system can't be resolved, skip generating the statsd config.
				// (instead of crashing). Mixer is optional.
			}

			if err := model.ValidateProxyConfig(&proxyConfig); err != nil {
				return err
			}

			if out, err := model.ToYAML(&proxyConfig); err != nil {
				log.Infof("Failed to serialize to YAML: %v", err)
			} else {
				log.Infof("Effective config: %s", out)
			}

			certs := []envoy.CertSource{
				{
					Directory: model.AuthCertsPath,
					Files:     []string{model.CertChainFilename, model.KeyFilename, model.RootCertFilename},
				},
			}

			if role.Type == model.Ingress {
				certs = append(certs, envoy.CertSource{
					Directory: model.IngressCertsPath,
					Files:     []string{model.IngressCertFilename, model.IngressKeyFilename},
				})
			}

			log.Infof("Monitored certs: %#v", certs)

			var envoyProxy proxy.Proxy
			if bootstrapv2 {
				// Using a different constructor - the code will likely be refactored / split from the v1,
				// but may expose same interface to minimize risks
				envoyProxy = envoy.NewV2Proxy(proxyConfig, role.ServiceNode(), proxyLogLevel, pilotSAN)
			} else {
				envoyProxy = envoy.NewProxy(proxyConfig, role.ServiceNode(), proxyLogLevel)
			}
			agent := proxy.NewAgent(envoyProxy, proxy.DefaultRetry)
			watcher := envoy.NewWatcher(proxyConfig, agent, role, certs, pilotSAN)
			ctx, cancel := context.WithCancel(context.Background())
			go watcher.Run(ctx)

			stop := make(chan struct{})
			cmd.WaitSignal(stop)
			<-stop
			cancel()
			return nil
		},
	}
)

func timeDuration(dur *duration.Duration) time.Duration {
	out, err := ptypes.Duration(dur)
	if err != nil {
		log.Warna(err)
	}
	return out
}

func init() {
	proxyCmd.PersistentFlags().StringVar((*string)(&registry), "serviceregistry",
		string(serviceregistry.KubernetesRegistry),
		fmt.Sprintf("Select the platform for service registry, options are {%s, %s, %s}",
			serviceregistry.KubernetesRegistry, serviceregistry.ConsulRegistry, serviceregistry.EurekaRegistry))
	proxyCmd.PersistentFlags().StringVar(&role.IPAddress, "ip", "",
		"Proxy IP address. If not provided uses ${INSTANCE_IP} environment variable.")
	proxyCmd.PersistentFlags().StringVar(&role.ID, "id", "",
		"Proxy unique ID. If not provided uses ${POD_NAME}.${POD_NAMESPACE} from environment variables")
	proxyCmd.PersistentFlags().StringVar(&role.Domain, "domain", "",
		"DNS domain suffix. If not provided uses ${POD_NAMESPACE}.svc.cluster.local")

	// Flags for proxy configuration
	values := model.DefaultProxyConfig()
	proxyCmd.PersistentFlags().StringVar(&configPath, "configPath", values.ConfigPath,
		"Path to the generated configuration file directory")
	proxyCmd.PersistentFlags().StringVar(&binaryPath, "binaryPath", values.BinaryPath,
		"Path to the proxy binary")
	proxyCmd.PersistentFlags().StringVar(&serviceCluster, "serviceCluster", values.ServiceCluster,
		"Service cluster")
	proxyCmd.PersistentFlags().StringVar(&availabilityZone, "availabilityZone", values.AvailabilityZone,
		"Availability zone")
	proxyCmd.PersistentFlags().DurationVar(&drainDuration, "drainDuration",
		timeDuration(values.DrainDuration),
		"The time in seconds that Envoy will drain connections during a hot restart")
	proxyCmd.PersistentFlags().DurationVar(&parentShutdownDuration, "parentShutdownDuration",
		timeDuration(values.ParentShutdownDuration),
		"The time in seconds that Envoy will wait before shutting down the parent process during a hot restart")
	proxyCmd.PersistentFlags().StringVar(&discoveryAddress, "discoveryAddress", values.DiscoveryAddress,
		"Address of the discovery service exposing xDS (e.g. istio-pilot:8080)")
	proxyCmd.PersistentFlags().DurationVar(&discoveryRefreshDelay, "discoveryRefreshDelay",
		timeDuration(values.DiscoveryRefreshDelay),
		"Polling interval for service discovery (used by EDS, CDS, LDS, but not RDS)")
	proxyCmd.PersistentFlags().StringVar(&zipkinAddress, "zipkinAddress", values.ZipkinAddress,
		"Address of the Zipkin service (e.g. zipkin:9411)")
	proxyCmd.PersistentFlags().DurationVar(&connectTimeout, "connectTimeout",
		timeDuration(values.ConnectTimeout),
		"Connection timeout used by Envoy for supporting services")
	proxyCmd.PersistentFlags().StringVar(&statsdUDPAddress, "statsdUdpAddress", values.StatsdUdpAddress,
		"IP Address and Port of a statsd UDP listener (e.g. 10.75.241.127:9125)")
	proxyCmd.PersistentFlags().IntVar(&proxyAdminPort, "proxyAdminPort", int(values.ProxyAdminPort),
		"Port on which Envoy should listen for administrative commands")
	proxyCmd.PersistentFlags().StringVar(&controlPlaneAuthPolicy, "controlPlaneAuthPolicy",
		values.ControlPlaneAuthPolicy.String(), "Control Plane Authentication Policy")
	proxyCmd.PersistentFlags().StringVar(&customConfigFile, "customConfigFile", values.CustomConfigFile,
		"Path to the generated configuration file directory")
	// Log levels are provided by the library https://github.com/gabime/spdlog, used by Envoy.
	proxyCmd.PersistentFlags().StringVar(&proxyLogLevel, "proxyLogLevel", "info",
		fmt.Sprintf("The log level used to start the Envoy proxy (choose from {%s, %s, %s, %s, %s, %s, %s})",
			"trace", "debug", "info", "warn", "err", "critical", "off"))
	proxyCmd.PersistentFlags().IntVar(&concurrency, "concurrency", int(values.Concurrency),
		"number of worker threads to run")
	proxyCmd.PersistentFlags().BoolVar(&bootstrapv2, "bootstrapv2", true,
		"Use bootstrap v2")

	// Attach the Istio logging options to the command.
	loggingOptions.AttachCobraFlags(rootCmd)

	cmd.AddFlags(rootCmd)

	rootCmd.AddCommand(proxyCmd)
	rootCmd.AddCommand(version.CobraCommand())

	rootCmd.AddCommand(collateral.CobraCommand(rootCmd, &doc.GenManHeader{
		Title:   "Istio Pilot Agent",
		Section: "pilot-agent CLI",
		Manual:  "Istio Pilot Agent",
	}))
}

func main() {
	// Needed to avoid "logging before flag.Parse" error with glog.
	cmd.SupressGlogWarnings()

	if err := rootCmd.Execute(); err != nil {
		log.Errora(err)
		os.Exit(-1)
	}
}
