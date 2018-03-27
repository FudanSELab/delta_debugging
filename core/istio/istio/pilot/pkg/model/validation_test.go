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
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	multierror "github.com/hashicorp/go-multierror"

	meshconfig "istio.io/api/mesh/v1alpha1"
	mpb "istio.io/api/mixer/v1"
	mccpb "istio.io/api/mixer/v1/config/client"
	routing "istio.io/api/routing/v1alpha1"
	routingv2 "istio.io/api/routing/v1alpha2"
	"istio.io/istio/pilot/pkg/model/test"
)

func TestConfigDescriptorValidate(t *testing.T) {
	badLabel := strings.Repeat("a", dns1123LabelMaxLength+1)
	goodLabel := strings.Repeat("a", dns1123LabelMaxLength-1)

	cases := []struct {
		name       string
		descriptor ConfigDescriptor
		wantErr    bool
	}{{
		name:       "Valid ConfigDescriptor (IstioConfig)",
		descriptor: IstioConfigTypes,
		wantErr:    false,
	}, {
		name: "Invalid DNS11234Label in ConfigDescriptor",
		descriptor: ConfigDescriptor{ProtoSchema{
			Type:        badLabel,
			MessageName: RouteRule.MessageName,
		}},
		wantErr: true,
	}, {
		name: "Bad MessageName in ProtoMessage",
		descriptor: ConfigDescriptor{ProtoSchema{
			Type:        goodLabel,
			MessageName: "nonexistent",
		}},
		wantErr: true,
	}, {
		name: "Missing key function",
		descriptor: ConfigDescriptor{ProtoSchema{
			Type:        RouteRule.Type,
			MessageName: RouteRule.MessageName,
		}},
		wantErr: true,
	}, {
		name:       "Duplicate type and message",
		descriptor: ConfigDescriptor{RouteRule, RouteRule},
		wantErr:    true,
	}}

	for _, c := range cases {
		if err := c.descriptor.Validate(); (err != nil) != c.wantErr {
			t.Errorf("%v failed: got %v but wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestConfigDescriptorValidateConfig(t *testing.T) {
	cases := []struct {
		name    string
		typ     string
		config  interface{}
		wantErr bool
	}{
		{
			name:    "bad configuration object",
			typ:     RouteRule.Type,
			config:  nil,
			wantErr: true,
		},
		{
			name:    "undeclared kind",
			typ:     "special-type",
			config:  &routing.RouteRule{},
			wantErr: true,
		},
		{
			name:    "non-proto object configuration",
			typ:     RouteRule.Type,
			config:  "non-proto objection configuration",
			wantErr: true,
		},
		{
			name:    "message type and kind mismatch",
			typ:     RouteRule.Type,
			config:  &routing.DestinationPolicy{},
			wantErr: true,
		},
		{
			name:    "ProtoSchema validation1",
			typ:     RouteRule.Type,
			config:  &routing.RouteRule{},
			wantErr: true,
		},
		{
			name:    "Successful validation",
			typ:     MockConfig.Type,
			config:  &test.MockConfig{Key: "test"},
			wantErr: false,
		},
	}

	types := append(IstioConfigTypes, MockConfig)

	for _, c := range cases {
		if err := types.ValidateConfig(c.typ, c.config); (err != nil) != c.wantErr {
			t.Errorf("%v failed: got error=%v but wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

var (
	endpoint1 = NetworkEndpoint{
		Address:     "192.168.1.1",
		Port:        10001,
		ServicePort: &Port{Name: "http", Port: 81, Protocol: ProtocolHTTP},
	}

	service1 = &Service{
		Hostname: "one.service.com",
		Address:  "192.168.3.1", // VIP
		Ports: PortList{
			&Port{Name: "http", Port: 81, Protocol: ProtocolHTTP},
			&Port{Name: "http-alt", Port: 8081, Protocol: ProtocolHTTP},
		},
	}
)

func TestServiceInstanceValidate(t *testing.T) {
	cases := []struct {
		name     string
		instance *ServiceInstance
		valid    bool
	}{
		{
			name: "nil service",
			instance: &ServiceInstance{
				Labels:   Labels{},
				Endpoint: endpoint1,
			},
		},
		{
			name: "bad label",
			instance: &ServiceInstance{
				Service:  service1,
				Labels:   Labels{"*": "-"},
				Endpoint: endpoint1,
			},
		},
		{
			name: "invalid service",
			instance: &ServiceInstance{
				Service: &Service{},
			},
		},
		{
			name: "invalid endpoint port and service port",
			instance: &ServiceInstance{
				Service: service1,
				Endpoint: NetworkEndpoint{
					Address: "192.168.1.2",
					Port:    -80,
				},
			},
		},
		{
			name: "endpoint missing service port",
			instance: &ServiceInstance{
				Service: service1,
				Endpoint: NetworkEndpoint{
					Address: "192.168.1.2",
					Port:    service1.Ports[1].Port,
					ServicePort: &Port{
						Name:     service1.Ports[1].Name + "-extra",
						Port:     service1.Ports[1].Port,
						Protocol: service1.Ports[1].Protocol,
					},
				},
			},
		},
		{
			name: "endpoint port and protocol mismatch",
			instance: &ServiceInstance{
				Service: service1,
				Endpoint: NetworkEndpoint{
					Address: "192.168.1.2",
					Port:    service1.Ports[1].Port,
					ServicePort: &Port{
						Name:     "http",
						Port:     service1.Ports[1].Port + 1,
						Protocol: ProtocolGRPC,
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Log("running case " + c.name)
		if got := c.instance.Validate(); (got == nil) != c.valid {
			t.Errorf("%s failed: got valid=%v but wanted valid=%v: %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestServiceValidate(t *testing.T) {
	ports := PortList{
		{Name: "http", Port: 80, Protocol: ProtocolHTTP},
		{Name: "http-alt", Port: 8080, Protocol: ProtocolHTTP},
	}
	badPorts := PortList{
		{Port: 80, Protocol: ProtocolHTTP},
		{Name: "http-alt^", Port: 8080, Protocol: ProtocolHTTP},
		{Name: "http", Port: -80, Protocol: ProtocolHTTP},
	}

	address := "192.168.1.1"

	cases := []struct {
		name    string
		service *Service
		valid   bool
	}{
		{
			name:    "empty hostname",
			service: &Service{Hostname: "", Address: address, Ports: ports},
		},
		{
			name:    "invalid hostname",
			service: &Service{Hostname: "hostname.^.com", Address: address, Ports: ports},
		},
		{
			name:    "empty ports",
			service: &Service{Hostname: "hostname", Address: address},
		},
		{
			name:    "bad ports",
			service: &Service{Hostname: "hostname", Address: address, Ports: badPorts},
		},
	}
	for _, c := range cases {
		if got := c.service.Validate(); (got == nil) != c.valid {
			t.Errorf("%s failed: got valid=%v but wanted valid=%v: %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestLabelsValidate(t *testing.T) {
	cases := []struct {
		name  string
		tags  Labels
		valid bool
	}{
		{
			name:  "empty tags",
			valid: true,
		},
		{
			name: "bad tag",
			tags: Labels{"^": "^"},
		},
		{
			name:  "good tag",
			tags:  Labels{"key": "value"},
			valid: true,
		},
	}
	for _, c := range cases {
		if got := c.tags.Validate(); (got == nil) != c.valid {
			t.Errorf("%s failed: got valid=%v but wanted valid=%v: %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateFQDN(t *testing.T) {
	if ValidateFQDN(strings.Repeat("x", 256)) == nil {
		t.Error("expected error on long FQDN")
	}
	if ValidateFQDN("") == nil {
		t.Error("expected error on empty FQDN")
	}
}

func TestValidateWildcardDomain(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"empty", "", "empty"},
		{"too long", strings.Repeat("x", 256), "too long"},
		{"happy", strings.Repeat("x", 63), ""},
		{"wildcard", "*", ""},
		{"wildcard multi-segment", "*.bar.com", ""},
		{"wildcard single segment", "*foo", ""},
		{"wildcard prefix", "*foo.bar.com", ""},
		{"wildcard prefix dash", "*-foo.bar.com", ""},
		{"bad wildcard", "foo.*.com", "invalid"},
		{"bad wildcard", "foo*.bar.com", "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWildcardDomain(tt.in)
			if err == nil && tt.out != "" {
				t.Fatalf("ValidateWildcardDomain(%v) = nil, wanted %q", tt.in, tt.out)
			} else if err != nil && tt.out == "" {
				t.Fatalf("ValidateWildcardDomain(%v) = %v, wanted nil", tt.in, err)
			} else if err != nil && !strings.Contains(err.Error(), tt.out) {
				t.Fatalf("ValidateWildcardDomain(%v) = %v, wanted %q", tt.in, err, tt.out)
			}
		})
	}
}

func TestValidateRouteAndIngressRule(t *testing.T) {
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{name: "empty destination policy", in: &routing.DestinationPolicy{}, valid: false},
		{name: "empty route rule", in: &routing.RouteRule{}, valid: false},
		{name: "route rule w destination", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
		},
			valid: true},
		{name: "route rule bad destination", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar?"},
		},
			valid: false},
		{name: "route rule bad destination", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar", Labels: Labels{"version": "v1"}},
		},
			valid: false},
		{name: "route rule bad match source", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Match:       &routing.MatchCondition{Source: &routing.IstioService{Name: "somehost!"}},
		},
			valid: false},
		{name: "route rule bad weight", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Route: []*routing.DestinationWeight{
				{Weight: -1},
			},
		},
			valid: false},
		{name: "route rule no weight", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Route: []*routing.DestinationWeight{
				{Labels: map[string]string{"a": "b"}},
			},
		},
			valid: true},
		{name: "route rule two destinationweights", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Route: []*routing.DestinationWeight{
				{Labels: map[string]string{"a": "b"}, Weight: 50},
				{Labels: map[string]string{"a": "c"}, Weight: 50},
			},
		},
			valid: true},
		{name: "route rule two destinationweights 99", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Route: []*routing.DestinationWeight{
				{Labels: map[string]string{"a": "b"}, Weight: 50},
				{Labels: map[string]string{"a": "c"}, Weight: 49},
			},
		},
			valid: false},
		{name: "route rule bad route tags", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Route: []*routing.DestinationWeight{
				{Labels: map[string]string{"a": "?"}},
			},
		},
			valid: false},
		{name: "route rule bad timeout", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpReqTimeout: &routing.HTTPTimeout{
				TimeoutPolicy: &routing.HTTPTimeout_SimpleTimeout{
					SimpleTimeout: &routing.HTTPTimeout_SimpleTimeoutPolicy{
						Timeout: &duration.Duration{Seconds: -1}},
				},
			},
		},
			valid: false},
		{name: "route rule bad retry attempts", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpReqRetries: &routing.HTTPRetry{
				RetryPolicy: &routing.HTTPRetry_SimpleRetry{
					SimpleRetry: &routing.HTTPRetry_SimpleRetryPolicy{
						Attempts: -1, PerTryTimeout: &duration.Duration{Seconds: 0}},
				},
			},
		},
			valid: false},
		{name: "route rule bad delay fixed seconds", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpFault: &routing.HTTPFaultInjection{
				Delay: &routing.HTTPFaultInjection_Delay{
					Percent: -1,
					HttpDelayType: &routing.HTTPFaultInjection_Delay_FixedDelay{
						FixedDelay: &duration.Duration{Seconds: 3}},
				},
			},
		},
			valid: false},
		{name: "route rule bad delay fixed seconds", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpFault: &routing.HTTPFaultInjection{
				Delay: &routing.HTTPFaultInjection_Delay{
					Percent: 100,
					HttpDelayType: &routing.HTTPFaultInjection_Delay_FixedDelay{
						FixedDelay: &duration.Duration{Seconds: -1}},
				},
			},
		},
			valid: false},
		{name: "route rule bad abort percent", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpFault: &routing.HTTPFaultInjection{
				Abort: &routing.HTTPFaultInjection_Abort{
					Percent:   -1,
					ErrorType: &routing.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: 500},
				},
			},
		},
			valid: false},
		{name: "route rule bad abort status", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpFault: &routing.HTTPFaultInjection{
				Abort: &routing.HTTPFaultInjection_Abort{
					Percent:   100,
					ErrorType: &routing.HTTPFaultInjection_Abort_HttpStatus{HttpStatus: -1},
				},
			},
		},
			valid: false},
		{name: "route rule bad unsupported status", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpFault: &routing.HTTPFaultInjection{
				Abort: &routing.HTTPFaultInjection_Abort{
					Percent:   100,
					ErrorType: &routing.HTTPFaultInjection_Abort_GrpcStatus{GrpcStatus: "test"},
				},
			},
		},
			valid: false},
		{name: "route rule bad delay exp seconds", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			HttpFault: &routing.HTTPFaultInjection{
				Delay: &routing.HTTPFaultInjection_Delay{
					Percent: 101,
					HttpDelayType: &routing.HTTPFaultInjection_Delay_ExponentialDelay{
						ExponentialDelay: &duration.Duration{Seconds: -1}},
				},
			},
		},
			valid: false},
		{name: "route rule bad throttle after seconds", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			L4Fault: &routing.L4FaultInjection{
				Throttle: &routing.L4FaultInjection_Throttle{
					Percent:            101,
					DownstreamLimitBps: -1,
					UpstreamLimitBps:   -1,
					ThrottleAfter: &routing.L4FaultInjection_Throttle_ThrottleAfterPeriod{
						ThrottleAfterPeriod: &duration.Duration{Seconds: -1}},
				},
				Terminate: &routing.L4FaultInjection_Terminate{
					Percent:              101,
					TerminateAfterPeriod: &duration.Duration{Seconds: -1},
				},
			},
		},
			valid: false},
		{name: "route rule bad throttle after bytes", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			L4Fault: &routing.L4FaultInjection{
				Throttle: &routing.L4FaultInjection_Throttle{
					Percent:            101,
					DownstreamLimitBps: -1,
					UpstreamLimitBps:   -1,
					ThrottleAfter: &routing.L4FaultInjection_Throttle_ThrottleAfterBytes{
						ThrottleAfterBytes: -1},
				},
			},
		},
			valid: false},
		{name: "route rule match valid subnets", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Match: &routing.MatchCondition{
				Tcp: &routing.L4MatchAttributes{
					SourceSubnet:      []string{"1.2.3.4"},
					DestinationSubnet: []string{"1.2.3.4/24"},
				},
			},
		},
			valid: true},
		{name: "route rule match invalid subnets", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Match: &routing.MatchCondition{
				Tcp: &routing.L4MatchAttributes{
					SourceSubnet:      []string{"foo", "1.2.3.4/banana"},
					DestinationSubnet: []string{"1.2.3.4/500", "1.2.3.4/-1"},
				},
				Udp: &routing.L4MatchAttributes{
					SourceSubnet:      []string{"1.2.3.4", "1.2.3.4/24", ""},
					DestinationSubnet: []string{"foo.2.3.4", "1.2.3"},
				},
			},
		},
			valid: false},
		{name: "route rule match invalid redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect: &routing.HTTPRedirect{
				Uri:       "",
				Authority: "",
			},
		},
			valid: false},
		{name: "route rule match valid host redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect: &routing.HTTPRedirect{
				Authority: "foo.bar.com",
			},
		},
			valid: true},
		{name: "route rule match valid path redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect: &routing.HTTPRedirect{
				Uri: "/new/path",
			},
		},
			valid: true},
		{name: "route rule match valid redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect: &routing.HTTPRedirect{
				Uri:       "/new/path",
				Authority: "foo.bar.com",
			},
		},
			valid: true},
		{name: "route rule match valid redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect: &routing.HTTPRedirect{
				Uri:       "/new/path",
				Authority: "foo.bar.com",
			},
			HttpFault: &routing.HTTPFaultInjection{},
		},
			valid: false},
		{name: "route rule match invalid redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect: &routing.HTTPRedirect{
				Uri: "/new/path",
			},
			Route: []*routing.DestinationWeight{
				{Labels: map[string]string{"version": "v1"}},
			},
		},
			valid: false},
		{name: "websocket upgrade invalid redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect: &routing.HTTPRedirect{
				Uri: "/new/path",
			},
			Route: []*routing.DestinationWeight{
				{Destination: &routing.IstioService{Name: "host"}, Weight: 100},
			},
			WebsocketUpgrade: true,
		},
			valid: false},
		{name: "route rule match invalid rewrite", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Rewrite:     &routing.HTTPRewrite{},
		},
			valid: false},
		{name: "route rule match valid host rewrite", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Rewrite: &routing.HTTPRewrite{
				Authority: "foo.bar.com",
			},
		},
			valid: true},
		{name: "route rule match rewrite and redirect", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Redirect:    &routing.HTTPRedirect{Uri: "/new/path"},
			Rewrite:     &routing.HTTPRewrite{Authority: "foo.bar.com"},
		},
			valid: false},
		{name: "route rule match valid prefix rewrite", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Rewrite: &routing.HTTPRewrite{
				Uri: "/new/path",
			},
		},
			valid: true},
		{name: "route rule match valid rewrite", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Rewrite: &routing.HTTPRewrite{
				Authority: "foo.bar.com",
				Uri:       "/new/path",
			},
		},
			valid: true},
		{name: "append headers", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			AppendHeaders: map[string]string{
				"name": "val",
			},
		},
			valid: true},
		{name: "append headers bad name", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			AppendHeaders: map[string]string{
				"": "val",
			},
		},
			valid: false},
		{name: "append headers bad val", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			AppendHeaders: map[string]string{
				"name": "",
			},
		},
			valid: false},
		{name: "mirror", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Mirror:      &routing.IstioService{Name: "barfoo"},
		},
			valid: true},
		{name: "mirror bad service", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			Mirror:      &routing.IstioService{},
		},
			valid: false},
		{name: "valid cors policy", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			CorsPolicy: &routing.CorsPolicy{
				MaxAge:           &duration.Duration{Seconds: 5},
				AllowOrigin:      []string{"http://foo.example"},
				AllowMethods:     []string{"POST", "GET", "OPTIONS"},
				AllowHeaders:     []string{"content-type"},
				AllowCredentials: &wrappers.BoolValue{Value: true},
				ExposeHeaders:    []string{"x-custom-header"},
			},
		},
			valid: true},
		{name: "cors policy invalid allow headers", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			CorsPolicy: &routing.CorsPolicy{
				MaxAge:           &duration.Duration{Seconds: 5},
				AllowOrigin:      []string{"http://foo.example"},
				AllowMethods:     []string{"POST", "GET", "OPTIONS"},
				AllowHeaders:     []string{"Content-Type"},
				AllowCredentials: &wrappers.BoolValue{Value: true},
				ExposeHeaders:    []string{"x-custom-header"},
			},
		},
			valid: false},
		{name: "cors policy invalid expose headers", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			CorsPolicy: &routing.CorsPolicy{
				MaxAge:           &duration.Duration{Seconds: 5},
				AllowOrigin:      []string{"http://foo.example"},
				AllowMethods:     []string{"POST", "GET", "OPTIONS"},
				AllowHeaders:     []string{"content-type"},
				AllowCredentials: &wrappers.BoolValue{Value: true},
				ExposeHeaders:    []string{"X-Custom-Header"},
			},
		},
			valid: false},
		{name: "invalid cors policy bad max age", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			CorsPolicy: &routing.CorsPolicy{
				MaxAge: &duration.Duration{Nanos: 1000000},
			},
		},
			valid: false},
		{name: "invalid cors policy invalid max age", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			CorsPolicy: &routing.CorsPolicy{
				MaxAge: &duration.Duration{Nanos: 100},
			},
		},
			valid: false},
		{name: "invalid cors policy bad allow method", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			CorsPolicy: &routing.CorsPolicy{
				MaxAge:           &duration.Duration{Seconds: 5},
				AllowOrigin:      []string{"http://foo.example"},
				AllowMethods:     []string{"POST", "GET", "UNSUPPORTED"},
				AllowHeaders:     []string{"content-type"},
				AllowCredentials: &wrappers.BoolValue{Value: true},
				ExposeHeaders:    []string{"x-custom-header"},
			},
		},
			valid: false},
		{name: "invalid cors policy bad allow method 2", in: &routing.RouteRule{
			Destination: &routing.IstioService{Name: "foobar"},
			CorsPolicy: &routing.CorsPolicy{
				MaxAge:           &duration.Duration{Seconds: 5},
				AllowOrigin:      []string{"http://foo.example"},
				AllowMethods:     []string{"POST", "get"},
				AllowHeaders:     []string{"content-type"},
				AllowCredentials: &wrappers.BoolValue{Value: true},
				ExposeHeaders:    []string{"x-custom-header"},
			},
		},
			valid: false},
	}
	for _, c := range cases {
		if got := ValidateRouteRule(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateRouteRule failed on %v: got valid=%v but wanted valid=%v: %v",
				c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateDestinationPolicy(t *testing.T) {
	cases := []struct {
		in    proto.Message
		valid bool
	}{
		{in: &routing.RouteRule{}, valid: false},
		{in: &routing.DestinationPolicy{}, valid: false},
		{in: &routing.DestinationPolicy{Destination: &routing.IstioService{Name: "foobar"}}, valid: true},
		{in: &routing.DestinationPolicy{
			Destination: &routing.IstioService{
				Name:      "?",
				Namespace: "?",
				Domain:    "a.?",
			},
			CircuitBreaker: &routing.CircuitBreaker{
				CbPolicy: &routing.CircuitBreaker_SimpleCb{
					SimpleCb: &routing.CircuitBreaker_SimpleCircuitBreakerPolicy{
						MaxConnections:               -1,
						HttpMaxPendingRequests:       -1,
						HttpMaxRequests:              -1,
						SleepWindow:                  &duration.Duration{Seconds: -1},
						HttpConsecutiveErrors:        -1,
						HttpDetectionInterval:        &duration.Duration{Seconds: -1},
						HttpMaxRequestsPerConnection: -1,
						HttpMaxEjectionPercent:       -1,
					},
				},
			},
		},
			valid: false},
		{in: &routing.DestinationPolicy{
			Destination: &routing.IstioService{Name: "ratings"},
			CircuitBreaker: &routing.CircuitBreaker{
				CbPolicy: &routing.CircuitBreaker_SimpleCb{
					SimpleCb: &routing.CircuitBreaker_SimpleCircuitBreakerPolicy{
						HttpMaxEjectionPercent: 101,
					},
				},
			},
		},
			valid: false},
		{in: &routing.DestinationPolicy{
			Destination: &routing.IstioService{Name: "foobar"},
			LoadBalancing: &routing.LoadBalancing{
				LbPolicy: &routing.LoadBalancing_Name{
					Name: 0,
				},
			},
		},
			valid: true},
		{in: &routing.DestinationPolicy{
			Destination:   &routing.IstioService{Name: "foobar"},
			LoadBalancing: &routing.LoadBalancing{},
		},
			valid: false},
		{in: &routing.DestinationPolicy{
			Source: &routing.IstioService{},
		},
			valid: false},
	}
	for _, c := range cases {
		if got := ValidateDestinationPolicy(c.in); (got == nil) != c.valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%v: %v", got == nil, c.valid, got)
		}
	}
}

func TestValidatePort(t *testing.T) {
	ports := map[int]bool{
		0:     false,
		65536: false,
		-1:    false,
		100:   true,
		1000:  true,
		65535: true,
	}
	for port, valid := range ports {
		if got := ValidatePort(port); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %d", got == nil, valid, got, port)
		}
	}
}

func TestValidateProxyAddress(t *testing.T) {
	addresses := map[string]bool{
		"istio-pilot:80":     true,
		"istio-pilot":        false,
		"isti..:80":          false,
		"10.0.0.100:9090":    true,
		"10.0.0.100":         false,
		"istio-pilot:port":   false,
		"istio-pilot:100000": false,
	}
	for addr, valid := range addresses {
		if got := ValidateProxyAddress(addr); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %s", got == nil, valid, got, addr)
		}
	}
}

func TestValidateDuration(t *testing.T) {
	durations := map[duration.Duration]bool{
		{Seconds: 1}:              true,
		{Seconds: 1, Nanos: -1}:   false,
		{Seconds: -11, Nanos: -1}: false,
		{Nanos: 1}:                false,
		{Seconds: 1, Nanos: 1}:    false,
	}
	for duration, valid := range durations {
		if got := ValidateDuration(&duration); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %v", got == nil, valid, got, duration)
		}
	}
}

func TestValidateParentAndDrain(t *testing.T) {
	type ParentDrainTime struct {
		Parent duration.Duration
		Drain  duration.Duration
		Valid  bool
	}

	combinations := []ParentDrainTime{
		{
			Parent: duration.Duration{Seconds: 2},
			Drain:  duration.Duration{Seconds: 1},
			Valid:  true,
		},
		{
			Parent: duration.Duration{Seconds: 1},
			Drain:  duration.Duration{Seconds: 1},
			Valid:  false,
		},
		{
			Parent: duration.Duration{Seconds: 1},
			Drain:  duration.Duration{Seconds: 2},
			Valid:  false,
		},
		{
			Parent: duration.Duration{Seconds: 2},
			Drain:  duration.Duration{Seconds: 1, Nanos: 1000000},
			Valid:  false,
		},
		{
			Parent: duration.Duration{Seconds: 2, Nanos: 1000000},
			Drain:  duration.Duration{Seconds: 1},
			Valid:  false,
		},
		{
			Parent: duration.Duration{Seconds: -2},
			Drain:  duration.Duration{Seconds: 1},
			Valid:  false,
		},
		{
			Parent: duration.Duration{Seconds: 2},
			Drain:  duration.Duration{Seconds: -1},
			Valid:  false,
		},
		{
			Parent: duration.Duration{Seconds: 1 + int64(time.Hour/time.Second)},
			Drain:  duration.Duration{Seconds: 10},
			Valid:  false,
		},
		{
			Parent: duration.Duration{Seconds: 10},
			Drain:  duration.Duration{Seconds: 1 + int64(time.Hour/time.Second)},
			Valid:  false,
		},
	}
	for _, combo := range combinations {
		if got := ValidateParentAndDrain(&combo.Drain, &combo.Parent); (got == nil) != combo.Valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for Parent:%v Drain:%v",
				got == nil, combo.Valid, got, combo.Parent, combo.Drain)
		}
	}
}

func TestValidateRefreshDelay(t *testing.T) {
	durations := map[duration.Duration]bool{
		{Seconds: 1}:     true,
		{Seconds: 36001}: false,
		{Nanos: 1}:       false,
	}
	for duration, valid := range durations {
		if got := ValidateRefreshDelay(&duration); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %v", got == nil, valid, got, duration)
		}
	}
}

func TestValidateConnectTimeout(t *testing.T) {
	durations := map[duration.Duration]bool{
		{Seconds: 1}:   true,
		{Seconds: 31}:  false,
		{Nanos: 99999}: false,
	}
	for duration, valid := range durations {
		if got := ValidateConnectTimeout(&duration); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %v", got == nil, valid, got, duration)
		}
	}
}

func TestValidateMeshConfig(t *testing.T) {
	if ValidateMeshConfig(&meshconfig.MeshConfig{}) == nil {
		t.Error("expected an error on an empty mesh config")
	}

	invalid := meshconfig.MeshConfig{
		MixerCheckServer:  "10.0.0.100",
		MixerReportServer: "10.0.0.100",
		ProxyListenPort:   0,
		ConnectTimeout:    ptypes.DurationProto(-1 * time.Second),
		AuthPolicy:        -1,
		RdsRefreshDelay:   ptypes.DurationProto(-1 * time.Second),
		DefaultConfig:     &meshconfig.ProxyConfig{},
	}

	err := ValidateMeshConfig(&invalid)
	if err == nil {
		t.Errorf("expected an error on invalid proxy mesh config: %v", invalid)
	} else {
		switch err.(type) {
		case *multierror.Error:
			// each field must cause an error in the field
			if len(err.(*multierror.Error).Errors) < 6 {
				t.Errorf("expected an error for each field %v", err)
			}
		default:
			t.Errorf("expected a multi error as output")
		}
	}
}

func TestValidateProxyConfig(t *testing.T) {
	if ValidateProxyConfig(&meshconfig.ProxyConfig{}) == nil {
		t.Error("expected an error on an empty proxy config")
	}

	invalid := meshconfig.ProxyConfig{
		ConfigPath:             "",
		BinaryPath:             "",
		DiscoveryAddress:       "10.0.0.100",
		ProxyAdminPort:         0,
		DrainDuration:          ptypes.DurationProto(-1 * time.Second),
		ParentShutdownDuration: ptypes.DurationProto(-1 * time.Second),
		DiscoveryRefreshDelay:  ptypes.DurationProto(-1 * time.Second),
		ConnectTimeout:         ptypes.DurationProto(-1 * time.Second),
		ServiceCluster:         "",
		StatsdUdpAddress:       "10.0.0.100",
		ZipkinAddress:          "10.0.0.100",
		ControlPlaneAuthPolicy: -1,
	}

	err := ValidateProxyConfig(&invalid)
	if err == nil {
		t.Errorf("expected an error on invalid proxy mesh config: %v", invalid)
	} else {
		switch err.(type) {
		case *multierror.Error:
			// each field must cause an error in the field
			if len(err.(*multierror.Error).Errors) < 12 {
				t.Errorf("expected an error for each field %v", err)
			}
		default:
			t.Errorf("expected a multi error as output")
		}
	}
}

func TestValidateIstioService(t *testing.T) {
	type IstioService struct {
		Service routing.IstioService
		Valid   bool
	}

	services := []IstioService{
		{
			Service: routing.IstioService{Name: "", Service: "", Domain: "", Namespace: ""},
			Valid:   false,
		},
		{
			Service: routing.IstioService{Service: "**cnn.com"},
			Valid:   false,
		},
		{
			Service: routing.IstioService{Service: "cnn.com", Labels: Labels{"*": ":"}},
			Valid:   false,
		},
		{
			Service: routing.IstioService{Service: "*cnn.com", Domain: "domain", Namespace: "namespace"},
			Valid:   false,
		},
		{
			Service: routing.IstioService{Service: "*cnn.com", Namespace: "namespace"},
			Valid:   false,
		},
		{
			Service: routing.IstioService{Service: "*cnn.com", Domain: "domain"},
			Valid:   false,
		},
		{
			Service: routing.IstioService{Name: "name", Service: "*cnn.com"},
			Valid:   false,
		},
		{
			Service: routing.IstioService{Service: "*cnn.com"},
			Valid:   true,
		},
		{
			Service: routing.IstioService{Name: "reviews", Domain: "svc.local", Namespace: "default"},
			Valid:   true,
		},
		{
			Service: routing.IstioService{Name: "reviews", Domain: "default", Namespace: "svc.local"},
			Valid:   false,
		},
	}

	for _, svc := range services {
		if got := ValidateIstioService(&svc.Service); (got == nil) != svc.Valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %s",
				got == nil, svc.Valid, got, svc.Service)
		}
	}
}

func TestValidateMatchCondition(t *testing.T) {
	for key, mc := range map[string]*routing.MatchCondition{
		"bad header key": {
			Request: &routing.MatchRequest{Headers: map[string]*routing.StringMatch{
				"XHeader": {MatchType: &routing.StringMatch_Exact{Exact: "test"}},
			}},
		},
		"bad header value": {
			Request: &routing.MatchRequest{Headers: map[string]*routing.StringMatch{
				"user-agent": {},
			}},
		},
		"uri header exact empty": {
			Request: &routing.MatchRequest{Headers: map[string]*routing.StringMatch{
				HeaderURI: {MatchType: &routing.StringMatch_Exact{}},
			}},
		},
		"uri header prefix empty": {
			Request: &routing.MatchRequest{Headers: map[string]*routing.StringMatch{
				HeaderURI: {MatchType: &routing.StringMatch_Prefix{}},
			}},
		},
		"uri header regex empty": {
			Request: &routing.MatchRequest{Headers: map[string]*routing.StringMatch{
				HeaderURI: {MatchType: &routing.StringMatch_Regex{}},
			}},
		},
	} {
		if ValidateMatchCondition(mc) == nil {
			t.Errorf("expected error %s for %#v", key, mc)
		}
	}

}

func TestValidateEgressRuleDomain(t *testing.T) {
	domains := map[string]bool{
		"cnn.com":    true,
		"cnn..com":   false,
		"10.0.0.100": true,
		"cnn.com:80": false,
		"*cnn.com":   true,
		"*.cnn.com":  true,
		"*-cnn.com":  true,
		"*.0.0.100":  true,
		"*0.0.100":   true,
		"*cnn*.com":  false,
		"cnn.*.com":  false,
		"*com":       true,
		"*0":         true,
		"**com":      false,
		"**0":        false,
		"*":          true,
		"":           false,
		"*.":         false,
	}

	for domain, valid := range domains {
		if got := ValidateEgressRuleDomain(domain); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %s", got == nil, valid, got, domain)
		}
	}
}

func TestValidateEgressRuleService(t *testing.T) {
	services := map[string]bool{
		"cnn.com":        true,
		"cnn..com":       false,
		"10.0.0.100":     true,
		"cnn.com:80":     false,
		"*cnn.com":       true,
		"*.cnn.com":      true,
		"*-cnn.com":      true,
		"*.0.0.100":      true,
		"*0.0.100":       true,
		"*cnn*.com":      false,
		"cnn.*.com":      false,
		"*com":           true,
		"*0":             true,
		"**com":          false,
		"**0":            false,
		"*":              true,
		"":               false,
		"*.":             false,
		"192.168.3.0/24": true,
		"50.1.2.3/32":    true,
		"10.15.0.0/16":   true,
		"10.15.0.0/0":    true,
		"10.15.0/16":     false,
		"10.15.0.0/33":   false,
	}

	for service, valid := range services {
		if got := ValidateEgressRuleService(service); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %s", got == nil, valid, got, service)
		}
	}
}

func TestValidateEgressRulePort(t *testing.T) {
	ports := map[*routing.EgressRule_Port]bool{
		{Port: 80, Protocol: "http"}:    true,
		{Port: 80, Protocol: "http2"}:   true,
		{Port: 80, Protocol: "grpc"}:    true,
		{Port: 443, Protocol: "https"}:  true,
		{Port: 80, Protocol: "https"}:   true,
		{Port: 443, Protocol: "http"}:   true,
		{Port: 1, Protocol: "http"}:     true,
		{Port: 2, Protocol: "https"}:    true,
		{Port: 80, Protocol: "tcp"}:     true,
		{Port: 1000, Protocol: "mongo"}: true,
		{Port: 80, Protocol: "udp"}:     false,
		{Port: 0, Protocol: "http"}:     false,
		{Port: 65536, Protocol: "http"}: false,
		{Port: 65535, Protocol: "http"}: true,
	}

	for port, valid := range ports {
		if got := ValidateEgressRulePort(port); (got == nil) != valid {
			t.Errorf("Failed: got valid=%t but wanted valid=%t: %v for %v", got == nil, valid, got, port)
		}
	}
}

func TestValidateIngressRule(t *testing.T) {
	cases := []struct {
		name string
		in   proto.Message
	}{
		{name: "nil egress rule"},
		{name: "empty egress rule", in: &routing.IngressRule{}},
		{name: "empty egress rule", in: &routing.IngressRule{
			Destination: &routing.IstioService{
				Service: "***", Labels: Labels{"version": "v1"},
			},
		}},
	}

	for _, c := range cases {
		if ValidateIngressRule(c.in) == nil {
			t.Errorf("ValidateIngressRule failed on %s: %#v", c.name, c.in)
		}
	}
}

func TestValidateEgressRule(t *testing.T) {
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{name: "nil egress rule"},
		{name: "empty egress rule", in: &routing.EgressRule{}, valid: false},
		{name: "valid egress rule",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{
					Service: "*cnn.com",
				},
				Ports: []*routing.EgressRule_Port{
					{Port: 80, Protocol: "http"},
					{Port: 443, Protocol: "https"},
				},
				UseEgressProxy: false},
			valid: true},
		{name: "valid egress rule with IP address",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{
					Service: "192.168.3.0",
				},
				Ports: []*routing.EgressRule_Port{
					{Port: 80, Protocol: "http"},
					{Port: 443, Protocol: "https"},
				},
				UseEgressProxy: false},
			valid: true},

		{name: "valid egress rule with tcp ports",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{
					Service: "192.168.3.0/24",
				},
				Ports: []*routing.EgressRule_Port{
					{Port: 80, Protocol: "tcp"},
					{Port: 443, Protocol: "tcp"},
				},
				UseEgressProxy: false},
			valid: true},
		{name: "egress rule with tcp ports, an http protocol",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{
					Service: "192.168.3.0/24",
				},
				Ports: []*routing.EgressRule_Port{
					{Port: 80, Protocol: "tcp"},
					{Port: 443, Protocol: "http"},
				},
				UseEgressProxy: false},
			valid: false},
		{name: "egress rule with use_egress_proxy = true, not yet implemented",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{
					Service: "*cnn.com",
				},
				Ports: []*routing.EgressRule_Port{
					{Port: 80, Protocol: "http"},
					{Port: 8080, Protocol: "http"},
				},
				UseEgressProxy: true},
			valid: false},
		{name: "empty destination",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{},
				Ports: []*routing.EgressRule_Port{
					{Port: 80, Protocol: "http"},
					{Port: 443, Protocol: "https"},
				},
				UseEgressProxy: false},
			valid: false},
		{name: "empty ports",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{
					Service: "*cnn.com",
				},
				Ports:          []*routing.EgressRule_Port{},
				UseEgressProxy: false},
			valid: false},
		{name: "duplicate port",
			in: &routing.EgressRule{
				Destination: &routing.IstioService{
					Service: "*cnn.com",
				},
				Ports: []*routing.EgressRule_Port{
					{Port: 80, Protocol: "http"},
					{Port: 443, Protocol: "https"},
					{Port: 80, Protocol: "https"},
				},
				UseEgressProxy: false},
			valid: false},
	}

	for _, c := range cases {
		if got := ValidateEgressRule(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateEgressRule failed on %v: got valid=%v but wanted valid=%v: %v",
				c.name, got == nil, c.valid, got)
		}
	}
}

var (
	validService    = &mccpb.IstioService{Service: "*cnn.com"}
	invalidService  = &mccpb.IstioService{Service: "^-foobar"}
	validAttributes = &mpb.Attributes{
		Attributes: map[string]*mpb.Attributes_AttributeValue{
			"api.service": {Value: &mpb.Attributes_AttributeValue_StringValue{"my-service"}},
		},
	}
	invalidAttributes = &mpb.Attributes{
		Attributes: map[string]*mpb.Attributes_AttributeValue{
			"api.service": {Value: &mpb.Attributes_AttributeValue_StringValue{""}},
		},
	}
)

func TestValidateMixerAttributes(t *testing.T) {
	cases := []struct {
		name  string
		in    *mpb.Attributes_AttributeValue
		valid bool
	}{
		{"happy string",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_StringValue{"my-service"}},
			true},
		{"invalid string",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_StringValue{""}},
			false},
		{"happy duration",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_DurationValue{&types.Duration{Seconds: 1}}},
			true},
		{"invalid duration",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_DurationValue{&types.Duration{Nanos: -1e9}}},
			false},
		{"happy bytes",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_BytesValue{[]byte{1, 2, 3}}},
			true},
		{"invalid bytes",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_BytesValue{[]byte{}}},
			false},
		{"happy timestamp",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_TimestampValue{&types.Timestamp{}}},
			true},
		{"invalid timestamp",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_TimestampValue{&types.Timestamp{Nanos: -1}}},
			false},
		{"nil timestamp",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_TimestampValue{nil}},
			false},
		{"happy stringmap",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_StringMapValue{
				&mpb.Attributes_StringMap{Entries: map[string]string{"foo": "bar"}}}},
			true},
		{"invalid stringmap",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_StringMapValue{
				&mpb.Attributes_StringMap{Entries: nil}}},
			false},
		{"nil stringmap",
			&mpb.Attributes_AttributeValue{Value: &mpb.Attributes_AttributeValue_StringMapValue{nil}},
			false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			attrs := &mpb.Attributes{
				Attributes: map[string]*mpb.Attributes_AttributeValue{"key": c.in},
			}
			if got := ValidateMixerAttributes(attrs); (got == nil) != c.valid {
				if c.valid {
					t.Fatal("got error, wanted none")
				} else {
					t.Fatal("got no error, wanted one")
				}
			}
		})
	}
}

func TestValidateHTTPAPISpec(t *testing.T) {
	var (
		validPattern = &mccpb.HTTPAPISpecPattern{
			Attributes: validAttributes,
			HttpMethod: "POST",
			Pattern: &mccpb.HTTPAPISpecPattern_UriTemplate{
				UriTemplate: "/pet/{id}",
			},
		}
		invalidPatternHTTPMethod = &mccpb.HTTPAPISpecPattern{
			Attributes: validAttributes,
			Pattern: &mccpb.HTTPAPISpecPattern_UriTemplate{
				UriTemplate: "/pet/{id}",
			},
		}
		invalidPatternURITemplate = &mccpb.HTTPAPISpecPattern{
			Attributes: validAttributes,
			HttpMethod: "POST",
			Pattern:    &mccpb.HTTPAPISpecPattern_UriTemplate{},
		}
		invalidPatternRegex = &mccpb.HTTPAPISpecPattern{
			Attributes: validAttributes,
			HttpMethod: "POST",
			Pattern:    &mccpb.HTTPAPISpecPattern_Regex{},
		}
		validAPIKey         = &mccpb.APIKey{Key: &mccpb.APIKey_Query{"api_key"}}
		invalidAPIKeyQuery  = &mccpb.APIKey{Key: &mccpb.APIKey_Query{}}
		invalidAPIKeyHeader = &mccpb.APIKey{Key: &mccpb.APIKey_Header{}}
		invalidAPIKeyCookie = &mccpb.APIKey{Key: &mccpb.APIKey_Cookie{}}
	)

	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{
			name: "missing pattern",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				ApiKeys:    []*mccpb.APIKey{validAPIKey},
			},
		},
		{
			name: "invalid pattern (bad attributes)",
			in: &mccpb.HTTPAPISpec{
				Attributes: invalidAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{validPattern},
				ApiKeys:    []*mccpb.APIKey{validAPIKey},
			},
		},
		{
			name: "invalid pattern (bad http_method)",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{invalidPatternHTTPMethod},
				ApiKeys:    []*mccpb.APIKey{validAPIKey},
			},
		},
		{
			name: "invalid pattern (missing uri_template)",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{invalidPatternURITemplate},
				ApiKeys:    []*mccpb.APIKey{validAPIKey},
			},
		},
		{
			name: "invalid pattern (missing regex)",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{invalidPatternRegex},
				ApiKeys:    []*mccpb.APIKey{validAPIKey},
			},
		},
		{
			name: "invalid api-key (missing query)",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{validPattern},
				ApiKeys:    []*mccpb.APIKey{invalidAPIKeyQuery},
			},
		},
		{
			name: "invalid api-key (missing header)",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{validPattern},
				ApiKeys:    []*mccpb.APIKey{invalidAPIKeyHeader},
			},
		},
		{
			name: "invalid api-key (missing cookie)",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{validPattern},
				ApiKeys:    []*mccpb.APIKey{invalidAPIKeyCookie},
			},
		},
		{
			name: "valid",
			in: &mccpb.HTTPAPISpec{
				Attributes: validAttributes,
				Patterns:   []*mccpb.HTTPAPISpecPattern{validPattern},
				ApiKeys:    []*mccpb.APIKey{validAPIKey},
			},
			valid: true,
		},
	}
	for _, c := range cases {
		if got := ValidateHTTPAPISpec(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateHTTPAPISpec(%v): got(%v) != want(%v): %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateHTTPAPISpecBinding(t *testing.T) {
	var (
		validHTTPAPISpecRef   = &mccpb.HTTPAPISpecReference{Name: "foo", Namespace: "bar"}
		invalidHTTPAPISpecRef = &mccpb.HTTPAPISpecReference{Name: "foo", Namespace: "--bar"}
	)
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{
			name: "no service",
			in: &mccpb.HTTPAPISpecBinding{
				Services: []*mccpb.IstioService{},
				ApiSpecs: []*mccpb.HTTPAPISpecReference{validHTTPAPISpecRef},
			},
		},
		{
			name: "invalid service",
			in: &mccpb.HTTPAPISpecBinding{
				Services: []*mccpb.IstioService{invalidService},
				ApiSpecs: []*mccpb.HTTPAPISpecReference{validHTTPAPISpecRef},
			},
		},
		{
			name: "no spec",
			in: &mccpb.HTTPAPISpecBinding{
				Services: []*mccpb.IstioService{validService},
				ApiSpecs: []*mccpb.HTTPAPISpecReference{},
			},
		},
		{
			name: "invalid spec",
			in: &mccpb.HTTPAPISpecBinding{
				Services: []*mccpb.IstioService{validService},
				ApiSpecs: []*mccpb.HTTPAPISpecReference{invalidHTTPAPISpecRef},
			},
		},
		{
			name: "valid",
			in: &mccpb.HTTPAPISpecBinding{
				Services: []*mccpb.IstioService{validService},
				ApiSpecs: []*mccpb.HTTPAPISpecReference{validHTTPAPISpecRef},
			},
			valid: true,
		},
	}
	for _, c := range cases {
		if got := ValidateHTTPAPISpecBinding(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateHTTPAPISpecBinding(%v): got(%v) != want(%v): %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateQuotaSpec(t *testing.T) {
	var (
		validMatch = &mccpb.AttributeMatch{
			Clause: map[string]*mccpb.StringMatch{
				"api.operation": {
					MatchType: &mccpb.StringMatch_Exact{
						Exact: "getPet",
					},
				},
			},
		}
		invalidMatchExact = &mccpb.AttributeMatch{
			Clause: map[string]*mccpb.StringMatch{
				"api.operation": {
					MatchType: &mccpb.StringMatch_Exact{Exact: ""},
				},
			},
		}
		invalidMatchPrefix = &mccpb.AttributeMatch{
			Clause: map[string]*mccpb.StringMatch{
				"api.operation": {
					MatchType: &mccpb.StringMatch_Prefix{Prefix: ""},
				},
			},
		}
		invalidMatchRegex = &mccpb.AttributeMatch{
			Clause: map[string]*mccpb.StringMatch{
				"api.operation": {
					MatchType: &mccpb.StringMatch_Regex{Regex: ""},
				},
			},
		}
		invalidQuota = &mccpb.Quota{
			Quota:  "",
			Charge: 0,
		}
		validQuota = &mccpb.Quota{
			Quota:  "myQuota",
			Charge: 2,
		}
	)
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{
			name: "no rules",
			in: &mccpb.QuotaSpec{
				Rules: []*mccpb.QuotaRule{{}},
			},
		},
		{
			name: "invalid match (exact)",
			in: &mccpb.QuotaSpec{
				Rules: []*mccpb.QuotaRule{{
					Match:  []*mccpb.AttributeMatch{invalidMatchExact},
					Quotas: []*mccpb.Quota{validQuota},
				}},
			},
		},
		{
			name: "invalid match (prefix)",
			in: &mccpb.QuotaSpec{
				Rules: []*mccpb.QuotaRule{{
					Match:  []*mccpb.AttributeMatch{invalidMatchPrefix},
					Quotas: []*mccpb.Quota{validQuota},
				}},
			},
		},
		{
			name: "invalid match (regex)",
			in: &mccpb.QuotaSpec{
				Rules: []*mccpb.QuotaRule{{
					Match:  []*mccpb.AttributeMatch{invalidMatchRegex},
					Quotas: []*mccpb.Quota{validQuota},
				}},
			},
		},
		{
			name: "no quota",
			in: &mccpb.QuotaSpec{
				Rules: []*mccpb.QuotaRule{{
					Match:  []*mccpb.AttributeMatch{validMatch},
					Quotas: []*mccpb.Quota{},
				}},
			},
		},
		{
			name: "invalid quota/charge",
			in: &mccpb.QuotaSpec{
				Rules: []*mccpb.QuotaRule{{
					Match:  []*mccpb.AttributeMatch{validMatch},
					Quotas: []*mccpb.Quota{invalidQuota},
				}},
			},
		},
		{
			name: "valid",
			in: &mccpb.QuotaSpec{
				Rules: []*mccpb.QuotaRule{{
					Match:  []*mccpb.AttributeMatch{validMatch},
					Quotas: []*mccpb.Quota{validQuota},
				}},
			},
			valid: true,
		},
	}
	for _, c := range cases {
		if got := ValidateQuotaSpec(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateQuotaSpec(%v): got(%v) != want(%v): %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateQuotaSpecBinding(t *testing.T) {
	var (
		validQuotaSpecRef   = &mccpb.QuotaSpecBinding_QuotaSpecReference{Name: "foo", Namespace: "bar"}
		invalidQuotaSpecRef = &mccpb.QuotaSpecBinding_QuotaSpecReference{Name: "foo", Namespace: "--bar"}
	)
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{
			name: "no service",
			in: &mccpb.QuotaSpecBinding{
				Services:   []*mccpb.IstioService{},
				QuotaSpecs: []*mccpb.QuotaSpecBinding_QuotaSpecReference{validQuotaSpecRef},
			},
		},
		{
			name: "invalid service",
			in: &mccpb.QuotaSpecBinding{
				Services:   []*mccpb.IstioService{invalidService},
				QuotaSpecs: []*mccpb.QuotaSpecBinding_QuotaSpecReference{validQuotaSpecRef},
			},
		},
		{
			name: "no spec",
			in: &mccpb.QuotaSpecBinding{
				Services:   []*mccpb.IstioService{validService},
				QuotaSpecs: []*mccpb.QuotaSpecBinding_QuotaSpecReference{},
			},
		},
		{
			name: "invalid spec",
			in: &mccpb.QuotaSpecBinding{
				Services:   []*mccpb.IstioService{validService},
				QuotaSpecs: []*mccpb.QuotaSpecBinding_QuotaSpecReference{invalidQuotaSpecRef},
			},
		},
		{
			name: "valid",
			in: &mccpb.QuotaSpecBinding{
				Services:   []*mccpb.IstioService{validService},
				QuotaSpecs: []*mccpb.QuotaSpecBinding_QuotaSpecReference{validQuotaSpecRef},
			},
			valid: true,
		},
	}
	for _, c := range cases {
		if got := ValidateQuotaSpecBinding(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateQuotaSpecBinding(%v): got(%v) != want(%v): %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateEndUserAuthenticationPolicySpec(t *testing.T) {
	var (
	//
	)
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{
			name: "no jwt",
			in:   &mccpb.EndUserAuthenticationPolicySpec{Jwts: []*mccpb.JWT{}},
		},
		{
			name: "invalid issuer",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Audiences:              []string{"audience_foo.example.com"},
					JwksUri:                "https://www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
		},
		{
			name: "invalid audiences",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{""},
					JwksUri:                "https://www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
		},
		{
			name: "missing jwks_uri",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{"audience_foo.example.com"},
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
		},
		{
			name: " jwks_uri with missing scheme",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{"audience_foo.example.com"},
					JwksUri:                "www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
		},
		{
			name: " jwks_uri invalid url",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{"audience_foo.example.com"},
					JwksUri:                ":foo",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
		},
		{
			name: "invalid duration",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:    "https://issuer.example.com",
					Audiences: []string{"audience_foo.example.com"},
					JwksUri:   "https://www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: &types.Duration{
						Seconds: -1,
					},
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
		},
		{
			name: "invalid location header",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{"audience_foo.example.com"},
					JwksUri:                "https://www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations:              []*mccpb.JWT_Location{{Scheme: &mccpb.JWT_Location_Header{}}},
				}},
			},
		},
		{
			name: "invalid location query",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{"audience_foo.example.com"},
					JwksUri:                "https://www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations:              []*mccpb.JWT_Location{{Scheme: &mccpb.JWT_Location_Query{}}},
				}},
			},
		},
		{
			name: "valid with https jwks_uri",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{"audience_foo.example.com"},
					JwksUri:                "https://www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
			valid: true,
		},
		{
			name: "valid with http jwks_uri",
			in: &mccpb.EndUserAuthenticationPolicySpec{
				Jwts: []*mccpb.JWT{{
					Issuer:                 "https://issuer.example.com",
					Audiences:              []string{"audience_foo.example.com"},
					JwksUri:                "http://www.example.com/oauth/v1/certs",
					PublicKeyCacheDuration: types.DurationProto(5 * time.Minute),
					Locations: []*mccpb.JWT_Location{{
						Scheme: &mccpb.JWT_Location_Header{Header: "x-goog-iap-jwt-assertion"},
					}},
				}},
			},
			valid: true,
		},
	}
	for _, c := range cases {
		if got := ValidateEndUserAuthenticationPolicySpec(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateEndUserAuthenticationPolicySpec(%v): got(%v) != want(%v): %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateEndUserAuthenticationPolicySpecBinding(t *testing.T) {
	var (
		validEndUserAuthenticationPolicySpecRef   = &mccpb.EndUserAuthenticationPolicySpecReference{Name: "foo", Namespace: "bar"}
		invalidEndUserAuthenticationPolicySpecRef = &mccpb.EndUserAuthenticationPolicySpecReference{Name: "foo", Namespace: "--bar"}
	)
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{
			name: "no service",
			in: &mccpb.EndUserAuthenticationPolicySpecBinding{
				Services: []*mccpb.IstioService{},
				Policies: []*mccpb.EndUserAuthenticationPolicySpecReference{validEndUserAuthenticationPolicySpecRef},
			},
		},
		{
			name: "invalid service",
			in: &mccpb.EndUserAuthenticationPolicySpecBinding{
				Services: []*mccpb.IstioService{invalidService},
				Policies: []*mccpb.EndUserAuthenticationPolicySpecReference{validEndUserAuthenticationPolicySpecRef},
			},
		},
		{
			name: "no spec",
			in: &mccpb.EndUserAuthenticationPolicySpecBinding{
				Services: []*mccpb.IstioService{validService},
				Policies: []*mccpb.EndUserAuthenticationPolicySpecReference{},
			},
		},
		{
			name: "invalid spec",
			in: &mccpb.EndUserAuthenticationPolicySpecBinding{
				Services: []*mccpb.IstioService{validService},
				Policies: []*mccpb.EndUserAuthenticationPolicySpecReference{invalidEndUserAuthenticationPolicySpecRef},
			},
		},
		{
			name: "valid",
			in: &mccpb.EndUserAuthenticationPolicySpecBinding{
				Services: []*mccpb.IstioService{validService},
				Policies: []*mccpb.EndUserAuthenticationPolicySpecReference{validEndUserAuthenticationPolicySpecRef},
			},
			valid: true,
		},
	}
	for _, c := range cases {
		if got := ValidateEndUserAuthenticationPolicySpecBinding(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateEndUserAuthenticationPolicySpecBinding(%v): got(%v) != want(%v): %v", c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateGateway(t *testing.T) {
	tests := []struct {
		name string
		in   proto.Message
		out  string
	}{
		{"empty", &routingv2.Gateway{}, "server"},
		{"invalid message", &routingv2.Server{}, "cannot cast"},
		{"happy domain",
			&routingv2.Gateway{
				Servers: []*routingv2.Server{{
					Hosts: []string{"foo.bar.com"},
					Port:  &routingv2.Port{Number: 7, Protocol: "http"},
				}},
			},
			""},
		{"happy ip",
			&routingv2.Gateway{
				Servers: []*routingv2.Server{{
					Hosts: []string{"192.168.0.1"},
					Port:  &routingv2.Port{Number: 7, Protocol: "http"},
				}},
			},
			""},
		{"happy cidr",
			&routingv2.Gateway{
				Servers: []*routingv2.Server{{
					Hosts: []string{"192.168.0.0/16"},
					Port:  &routingv2.Port{Number: 7, Protocol: "http"},
				}},
			},
			""},
		{"happy multiple servers",
			&routingv2.Gateway{
				Servers: []*routingv2.Server{
					{
						Hosts: []string{"foo.bar.com"},
						Port:  &routingv2.Port{Number: 7, Protocol: "http"},
					},
					{
						Hosts: []string{"192.168.0.0/16"},
						Port:  &routingv2.Port{Number: 18, Protocol: "redis"},
					}},
			},
			""},
		{"invalid port",
			&routingv2.Gateway{
				Servers: []*routingv2.Server{
					{
						Hosts: []string{"foo.bar.com"},
						Port:  &routingv2.Port{Number: 7, Protocol: "http"},
					},
					{
						Hosts: []string{"192.168.0.0/16"},
						Port:  &routingv2.Port{Number: 66000, Protocol: "redis"},
					}},
			},
			"port"},
		{"invalid domain",
			&routingv2.Gateway{
				Servers: []*routingv2.Server{
					{
						Hosts: []string{"foo.*.bar.com"},
						Port:  &routingv2.Port{Number: 7, Protocol: "http"},
					},
					{
						Hosts: []string{"192.168.0.0/33"},
						Port:  &routingv2.Port{Number: 66000, Protocol: "redis"},
					}},
			},
			"domain"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGateway(tt.in)
			if err == nil && tt.out != "" {
				t.Fatalf("ValidateGateway(%v) = nil, wanted %q", tt.in, tt.out)
			} else if err != nil && tt.out == "" {
				t.Fatalf("ValidateGateway(%v) = %v, wanted nil", tt.in, err)
			} else if err != nil && !strings.Contains(err.Error(), tt.out) {
				t.Fatalf("ValidateGateway(%v) = %v, wanted %q", tt.in, err, tt.out)
			}
		})
	}
}

func TestValidateServer(t *testing.T) {
	tests := []struct {
		name string
		in   *routingv2.Server
		out  string
	}{
		{"empty", &routingv2.Server{}, "host"},
		{"empty", &routingv2.Server{}, "port"},
		{"happy",
			&routingv2.Server{
				Hosts: []string{"foo.bar.com"},
				Port:  &routingv2.Port{Number: 7, Protocol: "http"},
			},
			""},
		{"invalid domain",
			&routingv2.Server{
				Hosts: []string{"foo.*.bar.com"},
				Port:  &routingv2.Port{Number: 7, Protocol: "http"},
			},
			"domain"},
		{"invalid port",
			&routingv2.Server{
				Hosts: []string{"foo.bar.com"},
				Port:  &routingv2.Port{Number: 66000, Protocol: "http"},
			},
			"port"},
		{"invalid tls options",
			&routingv2.Server{
				Hosts: []string{"foo.bar.com"},
				Port:  &routingv2.Port{Number: 1, Protocol: "http"},
				Tls:   &routingv2.Server_TLSOptions{Mode: routingv2.Server_TLSOptions_SIMPLE},
			},
			"TLS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServer(tt.in)
			if err == nil && tt.out != "" {
				t.Fatalf("validateServer(%v) = nil, wanted %q", tt.in, tt.out)
			} else if err != nil && tt.out == "" {
				t.Fatalf("validateServer(%v) = %v, wanted nil", tt.in, err)
			} else if err != nil && !strings.Contains(err.Error(), tt.out) {
				t.Fatalf("validateServer(%v) = %v, wanted %q", tt.in, err, tt.out)
			}
		})
	}
}

func TestValidateServerPort(t *testing.T) {
	tests := []struct {
		name string
		in   *routingv2.Port
		out  string
	}{
		{"empty", &routingv2.Port{}, "invalid protocol"},
		{"empty", &routingv2.Port{}, "port number"},
		{"happy",
			&routingv2.Port{
				Protocol: "http",
				Number:   1,
				Name:     "Henry",
			},
			""},
		{"invalid protocol",
			&routingv2.Port{
				Protocol: "kafka",
				Number:   1,
				Name:     "Henry",
			},
			"invalid protocol"},
		{"no port name/number",
			&routingv2.Port{
				Protocol: "http",
				Number:   0,
				Name:     "",
			},
			"either port number or name"},
		{"invalid number",
			&routingv2.Port{
				Protocol: "http",
				Number:   uint32(1 << 30),
				Name:     "",
			},
			"port number"},
		{"name, no number",
			&routingv2.Port{
				Protocol: "http",
				Number:   0,
				Name:     "Henry",
			},
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServerPort(tt.in)
			if err == nil && tt.out != "" {
				t.Fatalf("validateServerPort(%v) = nil, wanted %q", tt.in, tt.out)
			} else if err != nil && tt.out == "" {
				t.Fatalf("validateServerPort(%v) = %v, wanted nil", tt.in, err)
			} else if err != nil && !strings.Contains(err.Error(), tt.out) {
				t.Fatalf("validateServerPort(%v) = %v, wanted %q", tt.in, err, tt.out)
			}
		})
	}
}

func TestValidateTlsOptions(t *testing.T) {
	tests := []struct {
		name string
		in   *routingv2.Server_TLSOptions
		out  string
	}{
		{"empty", &routingv2.Server_TLSOptions{}, ""},
		{"simple",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_SIMPLE,
				ServerCertificate: "Captain Jean-Luc Picard"},
			""},
		{"simple with client bundle",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_SIMPLE,
				ServerCertificate: "Captain Jean-Luc Picard",
				CaCertificates:    "Commander William T. Riker"},
			""},
		{"simple no server cert",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_SIMPLE,
				ServerCertificate: ""},
			"server certificate"},
		{"mutual",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_MUTUAL,
				ServerCertificate: "Captain Jean-Luc Picard",
				CaCertificates:    "Commander William T. Riker"},
			""},
		{"mutual no server cert",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_MUTUAL,
				ServerCertificate: "",
				CaCertificates:    "Commander William T. Riker"},
			"server certificate"},
		{"mutual no client CA bundle",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_MUTUAL,
				ServerCertificate: "Captain Jean-Luc Picard",
				CaCertificates:    ""},
			"client CA bundle"},
		// this pair asserts we get errors about both client and server certs missing when in mutual mode
		// and both are absent, but requires less rewriting of the testing harness than merging the cases
		{"mutual no certs",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_MUTUAL,
				ServerCertificate: "",
				CaCertificates:    ""},
			"server certificate"},
		{"mutual no certs",
			&routingv2.Server_TLSOptions{
				Mode:              routingv2.Server_TLSOptions_MUTUAL,
				ServerCertificate: "",
				CaCertificates:    ""},
			"client CA bundle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTLSOptions(tt.in)
			if err == nil && tt.out != "" {
				t.Fatalf("validateTlsOptions(%v) = nil, wanted %q", tt.in, tt.out)
			} else if err != nil && tt.out == "" {
				t.Fatalf("validateTlsOptions(%v) = %v, wanted nil", tt.in, err)
			} else if err != nil && !strings.Contains(err.Error(), tt.out) {
				t.Fatalf("validateTlsOptions(%v) = %v, wanted %q", tt.in, err, tt.out)
			}
		})
	}
}

func TestValidateHTTPHeaderName(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
	}{
		{name: "header1", valid: true},
		{name: "HEADER2", valid: false},
	}

	for _, tc := range testCases {
		if got := ValidateHTTPHeaderName(tc.name); (got == nil) != tc.valid {
			t.Errorf("ValidateHTTPHeaderName(%q) => got valid=%v, want valid=%v",
				tc.name, got == nil, tc.valid)
		}
	}
}

func TestValidateCORSPolicy(t *testing.T) {
	testCases := []struct {
		name  string
		in    *routingv2.CorsPolicy
		valid bool
	}{
		{name: "valid", in: &routingv2.CorsPolicy{
			AllowMethods:  []string{"GET", "POST"},
			AllowHeaders:  []string{"header1", "header2"},
			ExposeHeaders: []string{"header3"},
			MaxAge:        &duration.Duration{Seconds: 2},
		}, valid: true},
		{name: "bad method", in: &routingv2.CorsPolicy{
			AllowMethods:  []string{"GET", "PUTT"},
			AllowHeaders:  []string{"header1", "header2"},
			ExposeHeaders: []string{"header3"},
			MaxAge:        &duration.Duration{Seconds: 2},
		}, valid: false},
		{name: "bad header", in: &routingv2.CorsPolicy{
			AllowMethods:  []string{"GET", "POST"},
			AllowHeaders:  []string{"header1", "header2"},
			ExposeHeaders: []string{"HEADER3"},
			MaxAge:        &duration.Duration{Seconds: 2},
		}, valid: false},
		{name: "bad max age", in: &routingv2.CorsPolicy{
			AllowMethods:  []string{"GET", "POST"},
			AllowHeaders:  []string{"header1", "header2"},
			ExposeHeaders: []string{"header3"},
			MaxAge:        &duration.Duration{Seconds: 2, Nanos: 42},
		}, valid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateCORSPolicy(tc.in); (got == nil) != tc.valid {
				t.Errorf("got valid=%v, want valid=%v: %v",
					got == nil, tc.valid, got)
			}
		})
	}
}

func TestValidateHTTPStatus(t *testing.T) {
	testCases := []struct {
		in    int32
		valid bool
	}{
		{-100, false},
		{0, true},
		{200, true},
		{600, true},
		{601, false},
	}

	for _, tc := range testCases {
		if got := validateHTTPStatus(tc.in); (got == nil) != tc.valid {
			t.Errorf("validateHTTPStatus(%d) => got valid=%v, want valid=%v",
				tc.in, got, tc.valid)
		}
	}
}

func TestValidateHTTPFaultInjectionAbort(t *testing.T) {
	testCases := []struct {
		name  string
		in    *routingv2.HTTPFaultInjection_Abort
		valid bool
	}{
		{name: "nil", in: nil, valid: true},
		{name: "valid", in: &routingv2.HTTPFaultInjection_Abort{
			Percent: 20,
			ErrorType: &routingv2.HTTPFaultInjection_Abort_HttpStatus{
				HttpStatus: 200,
			},
		}, valid: true},
		{name: "valid default", in: &routingv2.HTTPFaultInjection_Abort{
			ErrorType: &routingv2.HTTPFaultInjection_Abort_HttpStatus{
				HttpStatus: 200,
			},
		}, valid: true},
		{name: "invalid percent", in: &routingv2.HTTPFaultInjection_Abort{
			Percent: -1,
			ErrorType: &routingv2.HTTPFaultInjection_Abort_HttpStatus{
				HttpStatus: 200,
			},
		}, valid: false},
		{name: "invalid http status", in: &routingv2.HTTPFaultInjection_Abort{
			Percent: 20,
			ErrorType: &routingv2.HTTPFaultInjection_Abort_HttpStatus{
				HttpStatus: 9000,
			},
		}, valid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateHTTPFaultInjectionAbort(tc.in); (got == nil) != tc.valid {
				t.Errorf("got valid=%v, want valid=%v: %v",
					got == nil, tc.valid, got)
			}
		})
	}
}

func TestValidateHTTPFaultInjectionDelay(t *testing.T) {
	testCases := []struct {
		name  string
		in    *routingv2.HTTPFaultInjection_Delay
		valid bool
	}{
		{name: "nil", in: nil, valid: true},
		{name: "valid fixed", in: &routingv2.HTTPFaultInjection_Delay{
			Percent: 20,
			HttpDelayType: &routingv2.HTTPFaultInjection_Delay_FixedDelay{
				FixedDelay: &duration.Duration{Seconds: 3},
			},
		}, valid: true},
		{name: "valid default", in: &routingv2.HTTPFaultInjection_Delay{
			HttpDelayType: &routingv2.HTTPFaultInjection_Delay_FixedDelay{
				FixedDelay: &duration.Duration{Seconds: 3},
			},
		}, valid: true},
		{name: "invalid percent", in: &routingv2.HTTPFaultInjection_Delay{
			Percent: 101,
			HttpDelayType: &routingv2.HTTPFaultInjection_Delay_FixedDelay{
				FixedDelay: &duration.Duration{Seconds: 3},
			},
		}, valid: false},
		{name: "invalid delay", in: &routingv2.HTTPFaultInjection_Delay{
			Percent: 20,
			HttpDelayType: &routingv2.HTTPFaultInjection_Delay_FixedDelay{
				FixedDelay: &duration.Duration{Seconds: 3, Nanos: 42},
			},
		}, valid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateHTTPFaultInjectionDelay(tc.in); (got == nil) != tc.valid {
				t.Errorf("got valid=%v, want valid=%v: %v",
					got == nil, tc.valid, got)
			}
		})
	}
}

func TestValidateHTTPRetry(t *testing.T) {
	testCases := []struct {
		name  string
		in    *routingv2.HTTPRetry
		valid bool
	}{
		{name: "valid", in: &routingv2.HTTPRetry{
			Attempts:      10,
			PerTryTimeout: &duration.Duration{Seconds: 2},
		}, valid: true},
		{name: "valid default", in: &routingv2.HTTPRetry{
			Attempts: 10,
		}, valid: true},
		{name: "bad attempts", in: &routingv2.HTTPRetry{
			Attempts:      -1,
			PerTryTimeout: &duration.Duration{Seconds: 2},
		}, valid: false},
		{name: "invalid timeout", in: &routingv2.HTTPRetry{
			Attempts:      10,
			PerTryTimeout: &duration.Duration{Seconds: 2, Nanos: 1},
		}, valid: false},
		{name: "timeout too small", in: &routingv2.HTTPRetry{
			Attempts:      10,
			PerTryTimeout: &duration.Duration{Nanos: 999},
		}, valid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateHTTPRetry(tc.in); (got == nil) != tc.valid {
				t.Errorf("got valid=%v, want valid=%v: %v",
					got == nil, tc.valid, got)
			}
		})
	}
}

func TestValidateHTTPRewrite(t *testing.T) {
	testCases := []struct {
		name  string
		in    *routingv2.HTTPRewrite
		valid bool
	}{
		{name: "uri and authority", in: &routingv2.HTTPRewrite{
			Uri:       "/path/to/resource",
			Authority: "foobar.org",
		}, valid: true},
		{name: "uri", in: &routingv2.HTTPRewrite{
			Uri: "/path/to/resource",
		}, valid: true},
		{name: "authority", in: &routingv2.HTTPRewrite{
			Authority: "foobar.org",
		}, valid: true},
		{name: "no uri or authority", in: &routingv2.HTTPRewrite{}, valid: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateHTTPRewrite(tc.in); (got == nil) != tc.valid {
				t.Errorf("got valid=%v, want valid=%v: %v",
					got == nil, tc.valid, got)
			}
		})
	}
}

func TestValidateDestinationRule(t *testing.T) {
	cases := []struct {
		name  string
		in    proto.Message
		valid bool
	}{
		{name: "simple destination rule", in: &routingv2.DestinationRule{
			Name: "reviews",
			Subsets: []*routingv2.Subset{
				{Name: "v1", Labels: map[string]string{"version": "v1"}},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: true},

		{name: "missing destination name", in: &routingv2.DestinationRule{
			Name: "",
			Subsets: []*routingv2.Subset{
				{Name: "v1", Labels: map[string]string{"version": "v1"}},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: false},

		{name: "missing subset name", in: &routingv2.DestinationRule{
			Name: "reviews",
			Subsets: []*routingv2.Subset{
				{Name: "", Labels: map[string]string{"version": "v1"}},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: false},

		{name: "valid traffic policy, top level", in: &routingv2.DestinationRule{
			Name: "reviews",
			TrafficPolicy: &routingv2.TrafficPolicy{
				LoadBalancer: &routingv2.LoadBalancerSettings{
					LbPolicy: &routingv2.LoadBalancerSettings_Simple{
						Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
					},
				},
				ConnectionPool: &routingv2.ConnectionPoolSettings{
					Tcp:  &routingv2.ConnectionPoolSettings_TCPSettings{MaxConnections: 7},
					Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http2MaxRequests: 11},
				},
				OutlierDetection: &routingv2.OutlierDetection{
					Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
				},
			},
			Subsets: []*routingv2.Subset{
				{Name: "v1", Labels: map[string]string{"version": "v1"}},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: true},

		{name: "invalid traffic policy, top level", in: &routingv2.DestinationRule{
			Name: "reviews",
			TrafficPolicy: &routingv2.TrafficPolicy{
				LoadBalancer: &routingv2.LoadBalancerSettings{
					LbPolicy: &routingv2.LoadBalancerSettings_Simple{
						Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
					},
				},
				ConnectionPool: &routingv2.ConnectionPoolSettings{},
				OutlierDetection: &routingv2.OutlierDetection{
					Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
				},
			},
			Subsets: []*routingv2.Subset{
				{Name: "v1", Labels: map[string]string{"version": "v1"}},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: false},

		{name: "valid traffic policy, subset level", in: &routingv2.DestinationRule{
			Name: "reviews",
			Subsets: []*routingv2.Subset{
				{Name: "v1", Labels: map[string]string{"version": "v1"},
					TrafficPolicy: &routingv2.TrafficPolicy{
						LoadBalancer: &routingv2.LoadBalancerSettings{
							LbPolicy: &routingv2.LoadBalancerSettings_Simple{
								Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
							},
						},
						ConnectionPool: &routingv2.ConnectionPoolSettings{
							Tcp:  &routingv2.ConnectionPoolSettings_TCPSettings{MaxConnections: 7},
							Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http2MaxRequests: 11},
						},
						OutlierDetection: &routingv2.OutlierDetection{
							Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
						},
					},
				},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: true},

		{name: "invalid traffic policy, subset level", in: &routingv2.DestinationRule{
			Name: "reviews",
			Subsets: []*routingv2.Subset{
				{Name: "v1", Labels: map[string]string{"version": "v1"},
					TrafficPolicy: &routingv2.TrafficPolicy{
						LoadBalancer: &routingv2.LoadBalancerSettings{
							LbPolicy: &routingv2.LoadBalancerSettings_Simple{
								Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
							},
						},
						ConnectionPool: &routingv2.ConnectionPoolSettings{},
						OutlierDetection: &routingv2.OutlierDetection{
							Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
						},
					},
				},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: false},

		{name: "valid traffic policy, both levels", in: &routingv2.DestinationRule{
			Name: "reviews",
			TrafficPolicy: &routingv2.TrafficPolicy{
				LoadBalancer: &routingv2.LoadBalancerSettings{
					LbPolicy: &routingv2.LoadBalancerSettings_Simple{
						Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
					},
				},
				ConnectionPool: &routingv2.ConnectionPoolSettings{
					Tcp:  &routingv2.ConnectionPoolSettings_TCPSettings{MaxConnections: 7},
					Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http2MaxRequests: 11},
				},
				OutlierDetection: &routingv2.OutlierDetection{
					Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
				},
			},
			Subsets: []*routingv2.Subset{
				{Name: "v1", Labels: map[string]string{"version": "v1"},
					TrafficPolicy: &routingv2.TrafficPolicy{
						LoadBalancer: &routingv2.LoadBalancerSettings{
							LbPolicy: &routingv2.LoadBalancerSettings_Simple{
								Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
							},
						},
						ConnectionPool: &routingv2.ConnectionPoolSettings{
							Tcp:  &routingv2.ConnectionPoolSettings_TCPSettings{MaxConnections: 7},
							Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http2MaxRequests: 11},
						},
						OutlierDetection: &routingv2.OutlierDetection{
							Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
						},
					},
				},
				{Name: "v2", Labels: map[string]string{"version": "v2"}},
			},
		}, valid: true},
	}
	for _, c := range cases {
		if got := ValidateDestinationRule(c.in); (got == nil) != c.valid {
			t.Errorf("ValidateDestinationRule failed on %v: got valid=%v but wanted valid=%v: %v",
				c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateTrafficPolicy(t *testing.T) {
	cases := []struct {
		name  string
		in    routingv2.TrafficPolicy
		valid bool
	}{
		{name: "valid traffic policy", in: routingv2.TrafficPolicy{
			LoadBalancer: &routingv2.LoadBalancerSettings{
				LbPolicy: &routingv2.LoadBalancerSettings_Simple{
					Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
				},
			},
			ConnectionPool: &routingv2.ConnectionPoolSettings{
				Tcp:  &routingv2.ConnectionPoolSettings_TCPSettings{MaxConnections: 7},
				Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http2MaxRequests: 11},
			},
			OutlierDetection: &routingv2.OutlierDetection{
				Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
			},
		},
			valid: true},

		{name: "invalid traffic policy, bad connection pool", in: routingv2.TrafficPolicy{
			LoadBalancer: &routingv2.LoadBalancerSettings{
				LbPolicy: &routingv2.LoadBalancerSettings_Simple{
					Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
				},
			},
			ConnectionPool: &routingv2.ConnectionPoolSettings{},
			OutlierDetection: &routingv2.OutlierDetection{
				Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: 5},
			},
		},
			valid: false},

		{name: "invalid traffic policy, bad outlier detection", in: routingv2.TrafficPolicy{
			LoadBalancer: &routingv2.LoadBalancerSettings{
				LbPolicy: &routingv2.LoadBalancerSettings_Simple{
					Simple: routingv2.LoadBalancerSettings_ROUND_ROBIN,
				},
			},
			ConnectionPool: &routingv2.ConnectionPoolSettings{
				Tcp:  &routingv2.ConnectionPoolSettings_TCPSettings{MaxConnections: 7},
				Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http2MaxRequests: 11},
			},
			OutlierDetection: &routingv2.OutlierDetection{},
		},
			valid: false},
	}
	for _, c := range cases {
		if got := validateTrafficPolicy(&c.in); (got == nil) != c.valid {
			t.Errorf("ValidateTrafficPolicy failed on %v: got valid=%v but wanted valid=%v: %v",
				c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateConnectionPool(t *testing.T) {
	cases := []struct {
		name  string
		in    routingv2.ConnectionPoolSettings
		valid bool
	}{
		{name: "valid connection pool, tcp and http", in: routingv2.ConnectionPoolSettings{
			Tcp: &routingv2.ConnectionPoolSettings_TCPSettings{
				MaxConnections: 7,
				ConnectTimeout: &duration.Duration{Seconds: 2},
			},
			Http: &routingv2.ConnectionPoolSettings_HTTPSettings{
				Http1MaxPendingRequests:  2,
				Http2MaxRequests:         11,
				MaxRequestsPerConnection: 5,
				MaxRetries:               4,
			},
		},
			valid: true},

		{name: "valid connection pool, tcp only", in: routingv2.ConnectionPoolSettings{
			Tcp: &routingv2.ConnectionPoolSettings_TCPSettings{
				MaxConnections: 7,
				ConnectTimeout: &duration.Duration{Seconds: 2},
			},
		},
			valid: true},

		{name: "valid connection pool, http only", in: routingv2.ConnectionPoolSettings{
			Http: &routingv2.ConnectionPoolSettings_HTTPSettings{
				Http1MaxPendingRequests:  2,
				Http2MaxRequests:         11,
				MaxRequestsPerConnection: 5,
				MaxRetries:               4,
			},
		},
			valid: true},

		{name: "invalid connection pool, empty", in: routingv2.ConnectionPoolSettings{}, valid: false},

		{name: "invalid connection pool, bad max connections", in: routingv2.ConnectionPoolSettings{
			Tcp: &routingv2.ConnectionPoolSettings_TCPSettings{MaxConnections: -1}},
			valid: false},

		{name: "invalid connection pool, bad connect timeout", in: routingv2.ConnectionPoolSettings{
			Tcp: &routingv2.ConnectionPoolSettings_TCPSettings{
				ConnectTimeout: &duration.Duration{Seconds: 2, Nanos: 5}}},
			valid: false},

		{name: "invalid connection pool, bad max pending requests", in: routingv2.ConnectionPoolSettings{
			Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http1MaxPendingRequests: -1}},
			valid: false},

		{name: "invalid connection pool, bad max requests", in: routingv2.ConnectionPoolSettings{
			Http: &routingv2.ConnectionPoolSettings_HTTPSettings{Http2MaxRequests: -1}},
			valid: false},

		{name: "invalid connection pool, bad max requests per connection", in: routingv2.ConnectionPoolSettings{
			Http: &routingv2.ConnectionPoolSettings_HTTPSettings{MaxRequestsPerConnection: -1}},
			valid: false},

		{name: "invalid connection pool, bad max retries", in: routingv2.ConnectionPoolSettings{
			Http: &routingv2.ConnectionPoolSettings_HTTPSettings{MaxRetries: -1}},
			valid: false},
	}

	for _, c := range cases {
		if got := validateConnectionPool(&c.in); (got == nil) != c.valid {
			t.Errorf("ValidateConnectionSettings failed on %v: got valid=%v but wanted valid=%v: %v",
				c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateOutlierDetection(t *testing.T) {
	cases := []struct {
		name  string
		in    routingv2.OutlierDetection
		valid bool
	}{
		{name: "valid outlier detection", in: routingv2.OutlierDetection{
			Http: &routingv2.OutlierDetection_HTTPSettings{
				ConsecutiveErrors:  5,
				Interval:           &duration.Duration{Seconds: 2},
				BaseEjectionTime:   &duration.Duration{Seconds: 2},
				MaxEjectionPercent: 50,
			},
		}, valid: true},

		{name: "invalid outlier detection, bad consecutive errors", in: routingv2.OutlierDetection{
			Http: &routingv2.OutlierDetection_HTTPSettings{ConsecutiveErrors: -1}},
			valid: false},

		{name: "invalid outlier detection, bad interval", in: routingv2.OutlierDetection{
			Http: &routingv2.OutlierDetection_HTTPSettings{Interval: &duration.Duration{Seconds: 2, Nanos: 5}}},
			valid: false},

		{name: "invalid outlier detection, bad base ejection time", in: routingv2.OutlierDetection{
			Http: &routingv2.OutlierDetection_HTTPSettings{BaseEjectionTime: &duration.Duration{Seconds: 2, Nanos: 5}}},
			valid: false},

		{name: "invalid outlier detection, bad max ejection percent", in: routingv2.OutlierDetection{
			Http: &routingv2.OutlierDetection_HTTPSettings{MaxEjectionPercent: 105}},
			valid: false},
	}

	for _, c := range cases {
		if got := validateOutlierDetection(&c.in); (got == nil) != c.valid {
			t.Errorf("ValidateOutlierDetection failed on %v: got valid=%v but wanted valid=%v: %v",
				c.name, got == nil, c.valid, got)
		}
	}
}

func TestValidateExternalServices(t *testing.T) {
	cases := []struct {
		name  string
		in    routingv2.ExternalService
		valid bool
	}{
		{name: "discovery type DNS", in: routingv2.ExternalService{
			Hosts: []string{"*.google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "lon.google.com", Ports: map[string]uint32{"http-valid1": 8080}},
				{Address: "in.google.com", Ports: map[string]uint32{"http-valid2": 9080}},
			},
			Discovery: routingv2.ExternalService_DNS,
		},
			valid: true},

		{name: "discovery type DNS, IP in endpoints", in: routingv2.ExternalService{
			Hosts: []string{"*.google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "1.1.1.1", Ports: map[string]uint32{"http-valid1": 8080}},
				{Address: "in.google.com", Ports: map[string]uint32{"http-valid2": 9080}},
			},
			Discovery: routingv2.ExternalService_DNS,
		},
			valid: true},

		{name: "empty hosts", in: routingv2.ExternalService{
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "in.google.com", Ports: map[string]uint32{"http-valid2": 9080}},
			},
			Discovery: routingv2.ExternalService_DNS,
		},
			valid: false},

		{name: "bad hosts", in: routingv2.ExternalService{
			Hosts: []string{"-"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "in.google.com", Ports: map[string]uint32{"http-valid2": 9080}},
			},
			Discovery: routingv2.ExternalService_DNS,
		},
			valid: false},

		{name: "undefined endpoint port", in: routingv2.ExternalService{
			Hosts: []string{"google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 80, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "lon.google.com", Ports: map[string]uint32{"http-valid1": 8080}},
				{Address: "in.google.com", Ports: map[string]uint32{"http-dne": 9080}},
			},
			Discovery: routingv2.ExternalService_DNS,
		},
			valid: false},

		{name: "discovery type DNS, non-FQDN endpoint", in: routingv2.ExternalService{
			Hosts: []string{"*.google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "*.lon.google.com", Ports: map[string]uint32{"http-valid1": 8080}},
				{Address: "in.google.com", Ports: map[string]uint32{"http-dne": 9080}},
			},
			Discovery: routingv2.ExternalService_DNS,
		},
			valid: false},

		{name: "discovery type DNS, non-FQDN host", in: routingv2.ExternalService{
			Hosts: []string{"*.google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},

			Discovery: routingv2.ExternalService_DNS,
		},
			valid: false},

		{name: "discovery type DNS, no endpoints", in: routingv2.ExternalService{
			Hosts: []string{"google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},

			Discovery: routingv2.ExternalService_DNS,
		},
			valid: true},

		{name: "discovery type none", in: routingv2.ExternalService{
			Hosts: []string{"google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Discovery: routingv2.ExternalService_NONE,
		},
			valid: true},

		{name: "discovery type none, endpoints provided", in: routingv2.ExternalService{
			Hosts: []string{"google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "lon.google.com", Ports: map[string]uint32{"http-valid1": 8080}},
			},
			Discovery: routingv2.ExternalService_NONE,
		},
			valid: false},

		{name: "discovery type DNS, non-FQDN host", in: routingv2.ExternalService{
			Hosts: []string{"*.google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},

			Discovery: routingv2.ExternalService_DNS,
		},
			valid: false},

		{name: "discovery type static", in: routingv2.ExternalService{
			Hosts: []string{"172.1.2.16/16"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "1.1.1.1", Ports: map[string]uint32{"http-valid1": 8080}},
				{Address: "2.2.2.2", Ports: map[string]uint32{"http-valid2": 9080}},
			},
			Discovery: routingv2.ExternalService_STATIC,
		},
			valid: true},

		{name: "discovery type static, FQDN in endpoints", in: routingv2.ExternalService{
			Hosts: []string{"172.1.2.16/16"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "google.com", Ports: map[string]uint32{"http-valid1": 8080}},
				{Address: "2.2.2.2", Ports: map[string]uint32{"http-valid2": 9080}},
			},
			Discovery: routingv2.ExternalService_STATIC,
		},
			valid: false},

		{name: "discovery type static, missing endpoints", in: routingv2.ExternalService{
			Hosts: []string{"172.1.2.16/16"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Discovery: routingv2.ExternalService_STATIC,
		},
			valid: false},

		{name: "discovery type static, bad endpoint port name", in: routingv2.ExternalService{
			Hosts: []string{"172.1.2.16/16"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-valid1"},
				{Number: 8080, Protocol: "http", Name: "http-valid2"},
			},
			Endpoints: []*routingv2.ExternalService_Endpoint{
				{Address: "1.1.1.1", Ports: map[string]uint32{"http-valid1": 8080}},
				{Address: "2.2.2.2", Ports: map[string]uint32{"http-dne": 9080}},
			},
			Discovery: routingv2.ExternalService_STATIC,
		},
			valid: false},

		{name: "discovery type none, conflicting port names", in: routingv2.ExternalService{
			Hosts: []string{"google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-conflict"},
				{Number: 8080, Protocol: "http", Name: "http-conflict"},
			},
			Discovery: routingv2.ExternalService_NONE,
		},
			valid: false},

		{name: "discovery type none, conflicting port numbers", in: routingv2.ExternalService{
			Hosts: []string{"google.com"},
			Ports: []*routingv2.Port{
				{Number: 80, Protocol: "http", Name: "http-conflict1"},
				{Number: 80, Protocol: "http", Name: "http-conflict2"},
			},
			Discovery: routingv2.ExternalService_NONE,
		},
			valid: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ValidateExternalService(&c.in); (got == nil) != c.valid {
				t.Errorf("ValidateExternalService got valid=%v but wanted valid=%v: %v",
					got == nil, c.valid, got)
			}
		})
	}
}
