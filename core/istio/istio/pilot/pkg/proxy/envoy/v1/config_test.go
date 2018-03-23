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
	"crypto/sha1"
	"io/ioutil"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/test/util"
)

func TestRoutesByPath(t *testing.T) {
	cases := []struct {
		in       []*HTTPRoute
		expected []*HTTPRoute
	}{

		// Case 2: Prefix before path
		{
			in: []*HTTPRoute{
				{Prefix: "/api"},
				{Path: "/api/v1"},
			},
			expected: []*HTTPRoute{
				{Path: "/api/v1"},
				{Prefix: "/api"},
			},
		},

		// Case 3: Longer prefix before shorter prefix
		{
			in: []*HTTPRoute{
				{Prefix: "/api"},
				{Prefix: "/api/v1"},
			},
			expected: []*HTTPRoute{
				{Prefix: "/api/v1"},
				{Prefix: "/api"},
			},
		},
	}

	// Function to determine if two *Route slices
	// are the same (same Routes, same order)
	sameOrder := func(r1, r2 []*HTTPRoute) bool {
		for i, r := range r1 {
			if r.Path != r2[i].Path || r.Prefix != r2[i].Prefix {
				return false
			}
		}
		return true
	}

	for i, c := range cases {
		sort.Sort(RoutesByPath(c.in))
		if !sameOrder(c.in, c.expected) {
			t.Errorf("Invalid sort order for case %d", i)
		}
	}
}

func TestTCPRouteConfigByRoute(t *testing.T) {
	cases := []struct {
		name string
		in   []*TCPRoute
		want []*TCPRoute
	}{
		{
			name: "sorted by cluster",
			in: []*TCPRoute{{
				Cluster:           "cluster-b",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.2/32", "192.168.1.1/32"},
				DestinationPorts:  "5000",
			}},
			want: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.2/32", "192.168.1.1/32"},
				DestinationPorts:  "5000",
			}, {
				Cluster:           "cluster-b",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
			}},
		},
		{
			name: "sorted by DestinationIPList",
			in: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.2.1/32", "192.168.2.2/32"},
				DestinationPorts:  "5000",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
			}},
			want: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.2.1/32", "192.168.2.2/32"},
				DestinationPorts:  "5000",
			}},
		},
		{
			name: "sorted by DestinationPorts",
			in: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5001",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
			}},
			want: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5001",
			}},
		},
		{
			name: "sorted by SourceIPList",
			in: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.3.1/32", "192.168.3.2/32"},
				SourcePorts:       "5002",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.2.1/32", "192.168.2.2/32"},
				SourcePorts:       "5002",
			}},
			want: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.2.1/32", "192.168.2.2/32"},
				SourcePorts:       "5002",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.3.1/32", "192.168.3.2/32"},
				SourcePorts:       "5002",
			}},
		},
		{
			name: "sorted by SourcePorts",
			in: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.2.1/32", "192.168.2.2/32"},
				SourcePorts:       "5003",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.2.1/32", "192.168.2.2/32"},
				SourcePorts:       "5002",
			}},
			want: []*TCPRoute{{
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.2.1/32", "192.168.2.2/32"},
				SourcePorts:       "5002",
			}, {
				Cluster:           "cluster-a",
				DestinationIPList: []string{"192.168.1.1/32", "192.168.1.2/32"},
				DestinationPorts:  "5000",
				SourceIPList:      []string{"192.168.2.1/32", "192.168.2.2/32"},
				SourcePorts:       "5003",
			}},
		},
	}

	for _, c := range cases {
		sort.Sort(TCPRouteByRoute(c.in))
		if !reflect.DeepEqual(c.in, c.want) {
			t.Errorf("Invalid sort order for case %q:\n got  %#v\n want %#v", c.name, c.in, c.want)
		}
	}
}

type fileConfig struct {
	meta model.ConfigMeta
	file string
}

const (
	envoySidecarConfig     = "testdata/envoy-sidecar.json"
	envoySidecarAuthConfig = "testdata/envoy-sidecar-auth.json"
)

var (
	destinationRuleWorld = fileConfig{
		meta: model.ConfigMeta{Type: model.DestinationRule.Type, Name: "destination-world"},
		file: "testdata/destination-world-v1alpha2.yaml.golden",
	}

	destinationRuleWorldCB = fileConfig{
		meta: model.ConfigMeta{Type: model.DestinationRule.Type, Name: "destination-world-cb"},
		file: "testdata/destination-world-cb-v1alpha2.yaml.golden",
	}

	destinationRuleHello = fileConfig{
		meta: model.ConfigMeta{Type: model.DestinationRule.Type, Name: "destination-hello"},
		file: "testdata/destination-hello-v1alpha2.yaml.golden",
	}

	destinationRuleExternal = fileConfig{
		meta: model.ConfigMeta{Type: model.DestinationRule.Type, Name: "destination-google"},
		file: "testdata/subset-google-v1alpha2.yaml.golden",
	}

	cbPolicy = fileConfig{
		meta: model.ConfigMeta{Type: model.DestinationPolicy.Type, Name: "circuit-breaker"},
		file: "testdata/cb-policy.yaml.golden",
	}

	cbRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "circuit-breaker"},
		file: "testdata/cb-route-rule-v1alpha2.yaml.golden",
	}

	timeoutRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "timeout"},
		file: "testdata/timeout-route-rule.yaml.golden",
	}

	timeoutRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "timeout"},
		file: "testdata/timeout-route-rule-v1alpha2.yaml.golden",
	}

	weightedRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "weighted"},
		file: "testdata/weighted-route.yaml.golden",
	}

	weightedRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "weighted"},
		file: "testdata/weighted-route-v1alpha2.yaml.golden",
	}

	gatewayWeightedRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "gateway-weighted"},
		file: "testdata/gateway-weighted-route.yaml",
	}

	gatewayRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "gateway-simple"},
		file: "testdata/gateway-route.yaml",
	}

	gatewayWildcardRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "gateway-wildcard-simple"},
		file: "testdata/gateway-wildcard-route.yaml",
	}

	gatewayRouteRule2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "gateway-simple-2"},
		file: "testdata/gateway-route-2.yaml",
	}

	gatewayConfig = fileConfig{
		meta: model.ConfigMeta{Type: model.Gateway.Type, Name: "some-gateway"},
		file: "testdata/gateway.yaml",
	}

	gatewayConfig2 = fileConfig{
		meta: model.ConfigMeta{Type: model.Gateway.Type, Name: "some-gateway-2"},
		file: "testdata/gateway2.yaml",
	}

	gatewayWildcardConfig = fileConfig{
		meta: model.ConfigMeta{Type: model.Gateway.Type, Name: "some-gateway-wildcard"},
		file: "testdata/gateway-wildcard.yaml",
	}

	faultRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "fault"},
		file: "testdata/fault-route.yaml.golden",
	}

	faultRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "fault"},
		file: "testdata/fault-route-v1alpha2.yaml.golden",
	}

	multiMatchFaultRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "multi-match-fault"},
		file: "testdata/multi-match-fault-v1alpha2.yaml.golden",
	}

	redirectRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "redirect"},
		file: "testdata/redirect-route.yaml.golden",
	}

	redirectRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "redirect"},
		file: "testdata/redirect-route-v1alpha2.yaml.golden",
	}

	redirectRouteToEgressRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "redirect-to-egress"},
		file: "testdata/redirect-route-to-egress.yaml.golden",
	}

	rewriteRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "rewrite"},
		file: "testdata/rewrite-route.yaml.golden",
	}

	rewriteRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "rewrite"},
		file: "testdata/rewrite-route-v1alpha2.yaml.golden",
	}

	multiMatchRewriteRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "multi-match-rewrite"},
		file: "testdata/multi-match-rewrite-route-v1alpha2.yaml.golden",
	}

	googleTimeoutRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "egress-timeout"}, // FIXME: rename after switch to v1alpha2
		file: "testdata/google-timeout-rule-v1alpha2.yaml.golden",
	}

	websocketRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "websocket"},
		file: "testdata/websocket-route.yaml.golden",
	}

	websocketRouteRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "websocket"},
		file: "testdata/websocket-route-v1alpha2.yaml.golden",
	}

	egressRule = fileConfig{
		meta: model.ConfigMeta{Type: model.EgressRule.Type, Name: "google"},
		file: "testdata/egress-rule.yaml.golden",
	}

	externalServiceRule = fileConfig{
		meta: model.ConfigMeta{Type: model.ExternalService.Type, Name: "google"},
		file: "testdata/external-service-rule.yaml.golden",
	}

	externalServiceRuleDNS = fileConfig{
		meta: model.ConfigMeta{Type: model.ExternalService.Type, Name: "google"},
		file: "testdata/external-service-rule-dns.yaml.golden",
	}

	externalServiceRuleStatic = fileConfig{
		meta: model.ConfigMeta{Type: model.ExternalService.Type, Name: "google"},
		file: "testdata/external-service-rule-static.yaml.golden",
	}

	externalServiceRuleTCP = fileConfig{
		meta: model.ConfigMeta{Type: model.ExternalService.Type, Name: "google"},
		file: "testdata/external-service-rule-tcp.yaml.golden",
	}

	externalServiceRuleTCPDNS = fileConfig{
		meta: model.ConfigMeta{Type: model.ExternalService.Type, Name: "google"},
		file: "testdata/external-service-rule-tcp-dns.yaml.golden",
	}

	externalServiceRuleTCPStatic = fileConfig{
		meta: model.ConfigMeta{Type: model.ExternalService.Type, Name: "google"},
		file: "testdata/external-service-rule-tcp-static.yaml.golden",
	}

	externalServiceRouteRule = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "ext-route"},
		file: "testdata/external-service-route-rule.yaml.golden",
	}

	destinationRuleGoogleCB = fileConfig{
		meta: model.ConfigMeta{Type: model.DestinationRule.Type, Name: "google"},
		file: "testdata/subset-google-cb-v1alpha2.yaml.golden",
	}

	egressRuleCBPolicy = fileConfig{
		meta: model.ConfigMeta{Type: model.DestinationPolicy.Type, Name: "egress-circuit-breaker"},
		file: "testdata/egress-rule-cb-policy.yaml.golden",
	}

	egressRuleTimeoutRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "egress-timeout"},
		file: "testdata/egress-rule-timeout-route-rule.yaml.golden",
	}

	egressRuleTCP = fileConfig{
		meta: model.ConfigMeta{Type: model.EgressRule.Type, Name: "google-cloud-tcp"},
		file: "testdata/egress-rule-tcp.yaml.golden",
	}

	ingressRouteRule1 = fileConfig{
		meta: model.ConfigMeta{Type: model.IngressRule.Type, Name: "world"},
		file: "testdata/ingress-route-world.yaml.golden",
	}

	ingressRouteRule2 = fileConfig{
		meta: model.ConfigMeta{Type: model.IngressRule.Type, Name: "foo"},
		file: "testdata/ingress-route-foo.yaml.golden",
	}

	addHeaderRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "append-headers"},
		file: "testdata/addheaders-route.yaml.golden",
	}

	addHeaderRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "append-headers"},
		file: "testdata/addheaders-route-v1alpha2.yaml.golden",
	}

	corsPolicyRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "cors-policy"},
		file: "testdata/corspolicy-route.yaml.golden",
	}

	corsPolicyRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "cors-policy"},
		file: "testdata/corspolicy-route-v1alpha2.yaml.golden",
	}

	mirrorRule = fileConfig{
		meta: model.ConfigMeta{Type: model.RouteRule.Type, Name: "mirror-requests"},
		file: "testdata/mirror-route.yaml.golden",
	}

	mirrorRuleV2 = fileConfig{
		meta: model.ConfigMeta{Type: model.V1alpha2RouteRule.Type, Name: "mirror-requests"},
		file: "testdata/mirror-route-v1alpha2.yaml.golden",
	}

	// mixerclient service configuration
	mixerclientAPISpec = fileConfig{
		meta: model.ConfigMeta{Type: model.HTTPAPISpec.Type, Name: "api-spec"},
		file: "testdata/api-spec.yaml.golden",
	}

	mixerclientAPISpecBinding = fileConfig{
		meta: model.ConfigMeta{Type: model.HTTPAPISpecBinding.Type, Name: "api-spec-binding"},
		file: "testdata/api-spec-binding.yaml.golden",
	}

	mixerclientQuotaSpec = fileConfig{
		meta: model.ConfigMeta{Type: model.QuotaSpec.Type, Name: "quota-spec"},
		file: "testdata/quota-spec.yaml.golden",
	}

	mixerclientQuotaSpecBinding = fileConfig{
		meta: model.ConfigMeta{Type: model.QuotaSpecBinding.Type, Name: "quota-spec-binding"},
		file: "testdata/quota-spec-binding.yaml.golden",
	}

	mixerclientAuthSpec = fileConfig{
		meta: model.ConfigMeta{Type: model.EndUserAuthenticationPolicySpec.Type, Name: "auth-spec"},
		file: "testdata/auth-spec.yaml.golden",
	}

	mixerclientAuthSpecBinding = fileConfig{
		meta: model.ConfigMeta{Type: model.EndUserAuthenticationPolicySpecBinding.Type, Name: "auth-spec-binding"},
		file: "testdata/auth-spec-binding.yaml.golden",
	}
)

func addConfig(r model.ConfigStore, config fileConfig, t *testing.T) {
	schema, ok := model.IstioConfigTypes.GetByType(config.meta.Type)
	if !ok {
		t.Fatalf("missing schema for %q", config.meta.Type)
	}
	content, err := ioutil.ReadFile(config.file)
	if err != nil {
		t.Fatalf("reading %s: %s", config.file, err)
	}
	spec, err := schema.FromYAML(string(content))
	if err != nil {
		t.Fatalf("parsing yaml for %s: %s", config.file, err)
	}
	out := model.Config{
		ConfigMeta: config.meta,
		Spec:       spec,
	}

	// set default values for overriding
	out.ConfigMeta.Namespace = "default"
	out.ConfigMeta.Domain = "cluster.local"

	_, err = r.Create(out)
	if err != nil {
		t.Fatalf("create for %s: %s", config.file, err)
	}
}

func makeProxyConfig() meshconfig.ProxyConfig {
	proxyConfig := model.DefaultProxyConfig()
	proxyConfig.ZipkinAddress = "localhost:6000"
	proxyConfig.StatsdUdpAddress = "10.1.1.10:9125"
	proxyConfig.DiscoveryAddress = "istio-pilot.istio-system:15003"
	proxyConfig.DiscoveryRefreshDelay = ptypes.DurationProto(10 * time.Millisecond)
	return proxyConfig
}

var (
	pilotSAN = []string{"spiffe://cluster.local/ns/istio-system/sa/istio-pilot-service-account"}
)

func makeProxyConfigControlPlaneAuth() meshconfig.ProxyConfig {
	proxyConfig := makeProxyConfig()
	proxyConfig.ControlPlaneAuthPolicy = meshconfig.AuthenticationPolicy_MUTUAL_TLS
	return proxyConfig
}

func makeMeshConfig() meshconfig.MeshConfig {
	mesh := model.DefaultMeshConfig()
	mesh.MixerCheckServer = "istio-mixer.istio-system:9091"
	mesh.MixerReportServer = mesh.MixerCheckServer
	mesh.RdsRefreshDelay = ptypes.DurationProto(10 * time.Millisecond)
	return mesh
}

func TestProxyConfig(t *testing.T) {
	cases := []struct {
		envoyConfigFilename string
	}{
		{
			envoySidecarConfig,
		},
	}

	proxyConfig := makeProxyConfig()
	for _, c := range cases {
		config := BuildConfig(proxyConfig, nil)
		if config == nil {
			t.Fatal("Failed to generate config")
		}

		err := config.WriteFile(c.envoyConfigFilename)
		if err != nil {
			t.Fatalf(err.Error())
		}

		util.CompareYAML(c.envoyConfigFilename, t)
	}
}

func TestProxyConfigControlPlaneAuth(t *testing.T) {
	cases := []struct {
		envoyConfigFilename string
	}{
		{
			envoySidecarAuthConfig,
		},
	}

	proxyConfig := makeProxyConfigControlPlaneAuth()
	for _, c := range cases {
		config := BuildConfig(proxyConfig, pilotSAN)
		if config == nil {
			t.Fatal("Failed to generate config")
		}

		err := config.WriteFile(c.envoyConfigFilename)
		if err != nil {
			t.Fatalf(err.Error())
		}

		util.CompareYAML(c.envoyConfigFilename, t)
	}
}

func TestTruncateClusterName(t *testing.T) {
	data := make([]byte, MaxClusterNameLength+1)
	for i := range data {
		data[i] = byte('a' + i%26)
	}
	s := string(data) // the alphabet in lowercase, repeating...

	var trunc string
	less := s[:MaxClusterNameLength-1]
	trunc = truncateClusterName(less)
	if trunc != less {
		t.Errorf("Cluster name modified when truncating short cluster name:\nwant %s,\ngot %s", less, trunc)
	}
	eq := s[:MaxClusterNameLength]
	trunc = truncateClusterName(eq)
	if trunc != eq {
		t.Errorf("Cluster name modified when truncating cluster name:\nwant %s,\ngot %s", eq, trunc)
	}
	gt := s[:MaxClusterNameLength+1]
	trunc = truncateClusterName(gt)
	if len(trunc) != MaxClusterNameLength {
		t.Errorf("Cluster name length is not expected: want %d, got %d", MaxClusterNameLength, len(trunc))
	}
	prefixLen := MaxClusterNameLength - sha1.Size*2
	if gt[:prefixLen] != trunc[:prefixLen] {
		t.Errorf("Unexpected prefix:\nwant %s,\ngot %s", gt[:prefixLen], trunc[:prefixLen])
	}
}

func TestBuildJwksUriClusterNameAndAddress(t *testing.T) {
	cases := []struct {
		in          string
		wantAddress string
		wantName    string
		wantSSL     bool
		wantError   bool
	}{
		{
			in:          "https://www.googleapis.com/oauth2/v1/certs",
			wantAddress: "www.googleapis.com:443",
			wantName:    OutboundJWTURIClusterPrefix + "www.googleapis.com|443",
			wantSSL:     true,
		},
		{
			in:          "https://www.googleapis.com:443/oauth2/v1/certs",
			wantAddress: "www.googleapis.com:443",
			wantName:    OutboundJWTURIClusterPrefix + "www.googleapis.com|443",
			wantSSL:     true,
		},
		{
			in:          "http://example.com/oauth2/v1/certs",
			wantAddress: "example.com:80",
			wantName:    OutboundJWTURIClusterPrefix + "example.com|80",
			wantSSL:     false,
		},
		{
			in:        ":foo",
			wantError: true,
		},
	}
	for _, c := range cases {
		gotName, gotAddress, gotSSL, gotError := buildJWKSURIClusterNameAndAddress(c.in)
		if c.wantError != (gotError != nil) {
			t.Errorf("%s returned unexpected error: want %v got %v: %v",
				c.in, c.wantError, gotError != nil, gotError)
		} else {
			if gotAddress != c.wantAddress {
				t.Errorf("%s: gotAddress %v wantAddress %v", c.in, gotAddress, c.wantAddress)
			}
			if gotName != c.wantName {
				t.Errorf("%s: gotName %v wantName %v", c.in, gotName, c.wantName)
			}
			if gotSSL != c.wantSSL {
				t.Errorf("%s: gotSsl %v wantSSL %v", c.in, gotSSL, c.wantSSL)
			}
		}
	}
}

/*
var (
	ingressCertFile = "testdata/tls.crt"
	ingressKeyFile  = "testdata/tls.key"
)

func compareFile(filename string, golden []byte, t *testing.T) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Error loading %s: %s", filename, err.Error())
	}
	if string(content) != string(golden) {
		t.Errorf("Failed validating file %s, got %s", filename, string(content))
	}
	err = os.Remove(filename)
	if err != nil {
		t.Errorf("Failed cleaning up temporary file %s", filename)
	}
}
*/
