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
	"fmt"
	"os"
	"time"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra/doc"

	"istio.io/istio/pilot/cmd"
	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/pkg/collateral"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/version"
)

var (
	serverArgs bootstrap.PilotArgs

	loggingOptions = log.NewOptions()

	rootCmd = &cobra.Command{
		Use:   "pilot-discovery",
		Short: "Istio Pilot",
		Long:  "Istio Pilot provides fleet-wide traffic management capabilities in the Istio Service Mesh.",
	}

	discoveryCmd = &cobra.Command{
		Use:   "discovery",
		Short: "Start Istio proxy discovery service",
		RunE: func(c *cobra.Command, args []string) error {
			if err := log.Configure(loggingOptions); err != nil {
				return err
			}

			// Create the stop channel for all of the servers.
			stop := make(chan struct{})

			// Create the server for the discovery service.
			discoveryServer, err := bootstrap.NewServer(serverArgs)
			if err != nil {
				return fmt.Errorf("failed to create discovery service: %v", err)
			}

			// Start the server
			_, err = discoveryServer.Start(stop)
			if err != nil {
				return fmt.Errorf("failed to start discovery service: %v", err)
			}

			cmd.WaitSignal(stop)
			return nil
		},
	}
)

func init() {
	discoveryCmd.PersistentFlags().StringSliceVar(&serverArgs.Service.Registries, "registries",
		[]string{string(bootstrap.KubernetesRegistry)},
		fmt.Sprintf("Comma separated list of platform service registries to read from (choose one or more from {%s, %s, %s, %s, %s})",
			bootstrap.KubernetesRegistry, bootstrap.ConsulRegistry, bootstrap.EurekaRegistry, bootstrap.CloudFoundryRegistry, bootstrap.MockRegistry))
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Config.CFConfig, "cfConfig", "",
		"Cloud Foundry config file")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Config.ClusterRegistriesDir, "clusterRegistriesDir", "",
		"Directory for a file-based cluster config store")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Config.KubeConfig, "kubeconfig", "",
		"Use a Kubernetes configuration file instead of in-cluster configuration")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Mesh.ConfigFile, "meshConfig", "/etc/istio/config/mesh",
		fmt.Sprintf("File name for Istio mesh configuration. If not specified, a default mesh will be used."))
	discoveryCmd.PersistentFlags().StringVarP(&serverArgs.Namespace, "namespace", "n", "",
		"Select a namespace where the controller resides. If not set, uses ${POD_NAMESPACE} environment variable")

	// Config Controller options
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Config.FileDir, "configDir", "",
		"Directory to watch for updates to config yaml files. If specified, the files will be used as the source of config, rather than a CRD client.")
	discoveryCmd.PersistentFlags().StringVarP(&serverArgs.Config.ControllerOptions.WatchedNamespace, "appNamespace",
		"a", metav1.NamespaceAll,
		"Restrict the applications namespace the controller manages; if not set, controller watches all namespaces")
	discoveryCmd.PersistentFlags().DurationVar(&serverArgs.Config.ControllerOptions.ResyncPeriod, "resync", 60*time.Second,
		"Controller resync interval")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Config.ControllerOptions.DomainSuffix, "domain", "cluster.local",
		"DNS domain suffix")

	discoveryCmd.PersistentFlags().IntVar(&serverArgs.DiscoveryOptions.Port, "port", 8080,
		"Discovery service port")
	discoveryCmd.PersistentFlags().IntVar(&serverArgs.DiscoveryOptions.MonitoringPort, "monitoringPort", 9093,
		"HTTP port to use for the exposing pilot self-monitoring information")
	discoveryCmd.PersistentFlags().BoolVar(&serverArgs.DiscoveryOptions.EnableProfiling, "profile", true,
		"Enable profiling via web interface host:port/debug/pprof")
	discoveryCmd.PersistentFlags().BoolVar(&serverArgs.DiscoveryOptions.EnableCaching, "discovery_cache", true,
		"Enable caching discovery service responses")
	// TODO (rshriram): Need v1/v2 endpoints and option to selectively
	// enable webhook for specific xDS config (cds/lds/etc).
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.DiscoveryOptions.WebhookEndpoint, "webhookEndpoint", "",
		"Webhook API endpoint (supports http://sockethost, and unix:///absolute/path/to/socket")

	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Service.Consul.Config, "consulconfig", "",
		"Consul Config file for discovery")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Service.Consul.ServerURL, "consulserverURL", "",
		"URL for the Consul server")
	discoveryCmd.PersistentFlags().DurationVar(&serverArgs.Service.Consul.Interval, "consulserverInterval", 2*time.Second,
		"Interval (in seconds) for polling the Consul service registry")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Service.Eureka.ServerURL, "eurekaserverURL", "",
		"URL for the Eureka server")
	discoveryCmd.PersistentFlags().DurationVar(&serverArgs.Service.Eureka.Interval, "eurekaserverInterval", 2*time.Second,
		"Interval (in seconds) for polling the Eureka service registry")

	// Admission controller arguments.
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Admission.ExternalAdmissionWebhookName,
		"admission-webhook-name", "pilot-webhook.istio.io", "Webhook name for Pilot admission controller")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Admission.ServiceName,
		"admission-service", "istio-pilot",
		"Service name the admission controller uses during registration")
	discoveryCmd.PersistentFlags().IntVar(&serverArgs.Admission.Port, "admission-service-port", 443,
		"HTTPS port of the admission service. Must be 443 if service has more than one port ")
	discoveryCmd.PersistentFlags().StringVar(&serverArgs.Admission.SecretName, "admission-secret", "pilot-webhook",
		"Name of k8s secret for pilot webhook certs")
	discoveryCmd.PersistentFlags().DurationVar(&serverArgs.Admission.RegistrationDelay,
		"admission-registration-delay", 0*time.Second,
		"Time to delay webhook registration after starting webhook server")

	// Attach the Istio logging options to the command.
	loggingOptions.AttachCobraFlags(rootCmd)

	cmd.AddFlags(rootCmd)

	rootCmd.AddCommand(discoveryCmd)
	rootCmd.AddCommand(version.CobraCommand())
	rootCmd.AddCommand(collateral.CobraCommand(rootCmd, &doc.GenManHeader{
		Title:   "Istio Pilot Discovery",
		Section: "pilot-discovery CLI",
		Manual:  "Istio Pilot Discovery",
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
