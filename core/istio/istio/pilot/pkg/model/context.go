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

package model

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"
	multierror "github.com/hashicorp/go-multierror"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/log"
)

// Environment provides an aggregate environmental API for Pilot
type Environment struct {
	// Discovery interface for listing services and instances
	ServiceDiscovery

	// Accounts interface for listing service accounts
	ServiceAccounts

	// Config interface for listing routing rules
	IstioConfigStore

	// Mesh is the mesh config (to be merged into the config store)
	Mesh *meshconfig.MeshConfig

	// Mixer subject alternate name for mutual TLS
	MixerSAN []string
}

// Proxy defines the proxy attributes used by xDS identification
type Proxy struct {
	// Type specifies the node type
	Type NodeType

	// IPAddress is the IP address of the proxy used to identify it and its
	// co-located service instances. Example: "10.60.1.6"
	IPAddress string

	// ID is the unique platform-specific sidecar proxy ID
	ID string

	// Domain defines the DNS domain suffix for short hostnames (e.g.
	// "default.svc.cluster.local")
	Domain string
}

// NodeType decides the responsibility of the proxy serves in the mesh
type NodeType string

const (
	// Sidecar type is used for sidecar proxies in the application containers
	Sidecar NodeType = "sidecar"

	// Ingress type is used for cluster ingress proxies
	Ingress NodeType = "ingress"

	// Router type is used for standalone proxies acting as L7/L4 routers
	Router NodeType = "router"
)

// IsApplicationNodeType verifies that the NodeType is one of the declared constants in the model
func IsApplicationNodeType(nType NodeType) bool {
	switch nType {
	case Sidecar, Ingress, Router:
		return true
	default:
		return false
	}
}

// ServiceNode encodes the proxy node attributes into a URI-acceptable string
func (node Proxy) ServiceNode() string {
	return strings.Join([]string{
		string(node.Type), node.IPAddress, node.ID, node.Domain,
	}, serviceNodeSeparator)

}

// ParseServiceNode is the inverse of service node function
func ParseServiceNode(s string) (Proxy, error) {
	parts := strings.Split(s, serviceNodeSeparator)
	out := Proxy{}

	if len(parts) != 4 {
		return out, fmt.Errorf("missing parts in the service node %q", s)
	}

	out.Type = NodeType(parts[0])

	switch out.Type {
	case Sidecar, Ingress, Router:
	default:
		return out, fmt.Errorf("invalid node type (valid types: ingress, sidecar, router in the service node %q", s)
	}
	out.IPAddress = parts[1]

	// Does query from ingress or router have to carry valid IP address?
	if net.ParseIP(out.IPAddress) == nil && out.Type == Sidecar {
		return out, fmt.Errorf("invalid IP address %q in the service node %q", out.IPAddress, s)
	}

	out.ID = parts[2]
	out.Domain = parts[3]
	return out, nil
}

const (
	serviceNodeSeparator = "~"

	// IngressCertsPath is the path location for ingress certificates
	IngressCertsPath = "/etc/istio/ingress-certs/"

	// AuthCertsPath is the path location for mTLS certificates
	AuthCertsPath = "/etc/certs/"

	// CertChainFilename is mTLS chain file
	CertChainFilename = "cert-chain.pem"

	// KeyFilename is mTLS private key
	KeyFilename = "key.pem"

	// RootCertFilename is mTLS root cert
	RootCertFilename = "root-cert.pem"

	// IngressCertFilename is the ingress cert file name
	IngressCertFilename = "tls.crt"

	// IngressKeyFilename is the ingress private key file name
	IngressKeyFilename = "tls.key"

	// ConfigPathDir config directory for storing envoy json config files.
	// It also stores core files as per
	// https://github.com/istio/istio/blob/master/install/kubernetes/templates/istio-sidecar-injector-configmap-debug.yaml.tmpl#L27
	ConfigPathDir = "/etc/istio/proxy"

	// BinaryPathFilename envoy binary location
	BinaryPathFilename = "/usr/local/bin/envoy"

	// ServiceClusterName service cluster name used in xDS calls
	ServiceClusterName = "istio-proxy"

	// DiscoveryServerAddress discovery IP address:port
	DiscoveryServerAddress = "istio-pilot:15003"
)

// DefaultProxyConfig for individual proxies
func DefaultProxyConfig() meshconfig.ProxyConfig {
	return meshconfig.ProxyConfig{
		ConfigPath:             ConfigPathDir,
		BinaryPath:             BinaryPathFilename,
		ServiceCluster:         ServiceClusterName,
		AvailabilityZone:       "", //no service zone by default, i.e. AZ-aware routing is disabled
		DrainDuration:          ptypes.DurationProto(2 * time.Second),
		ParentShutdownDuration: ptypes.DurationProto(3 * time.Second),
		DiscoveryAddress:       DiscoveryServerAddress,
		DiscoveryRefreshDelay:  ptypes.DurationProto(1 * time.Second),
		ZipkinAddress:          "",
		ConnectTimeout:         ptypes.DurationProto(1 * time.Second),
		StatsdUdpAddress:       "",
		ProxyAdminPort:         15000,
		ControlPlaneAuthPolicy: meshconfig.AuthenticationPolicy_NONE,
		CustomConfigFile:       "",
		Concurrency:            0,
	}
}

// DefaultMeshConfig configuration
func DefaultMeshConfig() meshconfig.MeshConfig {
	config := DefaultProxyConfig()
	return meshconfig.MeshConfig{
		// TODO(mixeraddress is deprecated. Remove)
		MixerAddress:          "",
		MixerCheckServer:      "",
		MixerReportServer:     "",
		DisablePolicyChecks:   false,
		ProxyListenPort:       15001,
		ConnectTimeout:        ptypes.DurationProto(1 * time.Second),
		IngressClass:          "istio",
		IngressControllerMode: meshconfig.MeshConfig_STRICT,
		AuthPolicy:            meshconfig.MeshConfig_NONE,
		RdsRefreshDelay:       ptypes.DurationProto(1 * time.Second),
		EnableTracing:         true,
		AccessLogFile:         "/dev/stdout",
		DefaultConfig:         &config,
	}
}

// ApplyMeshConfigDefaults returns a new MeshConfig decoded from the
// input YAML with defaults applied to omitted configuration values.
func ApplyMeshConfigDefaults(yaml string) (*meshconfig.MeshConfig, error) {
	out := DefaultMeshConfig()
	if err := ApplyYAML(yaml, &out); err != nil {
		return nil, multierror.Prefix(err, "failed to convert to proto.")
	}

	// Reset the default ProxyConfig as jsonpb.UnmarshalString doesn't
	// handled nested decode properly for our use case.
	prevDefaultConfig := out.DefaultConfig
	defaultProxyConfig := DefaultProxyConfig()
	out.DefaultConfig = &defaultProxyConfig

	// Re-apply defaults to ProxyConfig if they were defined in the
	// original input MeshConfig.ProxyConfig.
	if prevDefaultConfig != nil {
		origProxyConfigYAML, err := ToYAML(prevDefaultConfig)
		if err != nil {
			return nil, multierror.Prefix(err, "failed to re-encode default proxy config")
		}
		if err := ApplyYAML(origProxyConfigYAML, out.DefaultConfig); err != nil {
			return nil, multierror.Prefix(err, "failed to convert to proto.")
		}
	}

	// Backward compat option: if mixer address is set but
	// mixer_check_server and mixer_report_server are unset, copy the value
	// into these two config vars.
	if out.MixerAddress != "" && out.MixerCheckServer == "" && out.MixerReportServer == "" {
		out.MixerCheckServer = out.MixerAddress
		out.MixerReportServer = out.MixerAddress
	}

	if err := ValidateMeshConfig(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ParsePort extracts port number from a valid proxy address
func ParsePort(addr string) int {
	port, err := strconv.Atoi(addr[strings.Index(addr, ":")+1:])
	if err != nil {
		log.Warna(err)
	}

	return port
}
