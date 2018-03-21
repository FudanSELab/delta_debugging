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

package model_test

import (
	"fmt"
	"reflect"
	"testing"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/proxy/envoy/v1/mock"
)

func TestServiceNode(t *testing.T) {
	nodes := []struct {
		in  model.Proxy
		out string
	}{
		{
			in:  mock.HelloProxyV0,
			out: "sidecar~10.1.1.0~v0.default~default.svc.cluster.local",
		},
		{
			in: model.Proxy{
				Type:   model.Ingress,
				ID:     "random",
				Domain: "local",
			},
			out: "ingress~~random~local",
		},
	}

	for _, node := range nodes {
		out := node.in.ServiceNode()
		if out != node.out {
			t.Errorf("%#v.ServiceNode() => Got %s, want %s", node.in, out, node.out)
		}
		in, err := model.ParseServiceNode(node.out)
		if err != nil {
			t.Errorf("ParseServiceNode(%q) => Got error %v", node.out, err)
		}
		if !reflect.DeepEqual(in, node.in) {
			t.Errorf("ParseServiceNode(%q) => Got %#v, want %#v", node.out, in, node.in)
		}
	}
}

func TestParsePort(t *testing.T) {
	if port := model.ParsePort("localhost:3000"); port != 3000 {
		t.Errorf("ParsePort(localhost:3000) => Got %d, want 3000", port)
	}
	if port := model.ParsePort("localhost"); port != 0 {
		t.Errorf("ParsePort(localhost) => Got %d, want 0", port)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := model.DefaultProxyConfig()
	if err := model.ValidateProxyConfig(&config); err != nil {
		t.Errorf("validation of default proxy config failed with %v", err)
	}
}

func TestDefaultMeshConfig(t *testing.T) {
	mesh := model.DefaultMeshConfig()
	if err := model.ValidateMeshConfig(&mesh); err != nil {
		t.Errorf("validation of default mesh config failed with %v", err)
	}
}

func TestApplyMeshConfigDefaults(t *testing.T) {
	configPath := "/test/config/patch"
	yaml := fmt.Sprintf(`
defaultConfig:
  configPath: %s
`, configPath)

	want := model.DefaultMeshConfig()
	want.DefaultConfig.ConfigPath = configPath

	got, err := model.ApplyMeshConfigDefaults(yaml)
	if err != nil {
		t.Fatalf("ApplyMeshConfigDefaults() failed: %v", err)
	}
	if !reflect.DeepEqual(got, &want) {
		t.Fatalf("Wrong default values:\n got %#v \nwant %#v", got, &want)
	}
}
