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

package external

import (
	"fmt"
	"sort"
	"testing"

	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/config/memory"
	"istio.io/istio/pilot/pkg/model"
)

func createExternalServices(externalSvcs []*networking.ExternalService, store model.IstioConfigStore, t *testing.T) {
	for _, svc := range externalSvcs {
		config := model.Config{
			ConfigMeta: model.ConfigMeta{
				Type:      model.ExternalService.Type,
				Name:      svc.Hosts[0],
				Namespace: "default",
				Domain:    "cluster.local",
			},
			Spec: svc,
		}

		_, err := store.Create(config)
		if err != nil {
			t.Errorf("error occurred crearting ExternalService config: %v", err)
		}
	}
}

func initConfigStore() model.IstioConfigStore {
	configDescriptor := model.ConfigDescriptor{
		model.ExternalService,
	}
	store := memory.Make(configDescriptor)
	configController := memory.NewController(store)
	return model.MakeIstioStore(configController)
}

func TestServiceDiscoveryServices(t *testing.T) {
	store := initConfigStore()
	sd := NewServiceDiscovery(store)
	expectedServices := []*model.Service{
		makeService("*.google.com", map[string]int{"http-port": 80, "http-alt-port": 8080}, model.DNSLB),
		makeService("172.217.0.0/16", map[string]int{"tcp-444": 444}, model.DNSLB),
	}

	createExternalServices([]*networking.ExternalService{httpDNS, tcpDNS}, store, t)

	services, err := sd.Services()
	if err != nil {
		t.Errorf("Services() encountered unexpected error: %v", err)
	}
	sortServices(services)
	sortServices(expectedServices)
	if err := compare(t, services, expectedServices); err != nil {
		t.Error(err)
	}
}

func TestServiceDiscoveryGetService(t *testing.T) {
	host := "*.google.com"
	hostDNE := "does.not.exist.local"

	store := initConfigStore()
	sd := NewServiceDiscovery(store)

	createExternalServices([]*networking.ExternalService{httpDNS, tcpDNS}, store, t)

	service, err := sd.GetService(hostDNE)
	if err != nil {
		t.Errorf("GetService() encountered unexpected error: %v", err)
	}
	if service != nil {
		t.Errorf("GetService(%q) => should not exist, got %s", hostDNE, service.Hostname)
	}

	service, err = sd.GetService(host)
	if err != nil {
		t.Errorf("GetService(%q) encountered unexpected error: %v", host, err)
	}
	if service == nil {
		t.Errorf("GetService(%q) => should exist", host)
	}
	if service.Hostname != host {
		t.Errorf("GetService(%q) => %q, want %q", host, service.Hostname, host)
	}
}

func TestServiceDiscoveryGetProxyServiceInstances(t *testing.T) {
	store := initConfigStore()
	sd := NewServiceDiscovery(store)

	createExternalServices([]*networking.ExternalService{httpStatic, tcpStatic}, store, t)

	instances, err := sd.GetProxyServiceInstances(model.Proxy{IPAddress: "2.2.2.2"})
	if err != nil {
		t.Errorf("GetProxyServiceInstances() encountered unexpected error: %v", err)
	}

	if len(instances) != 0 {
		t.Error("ExternalService registry should never return something for GetProxyServiceInstances()")
	}
}

func TestServiceDiscoveryInstances(t *testing.T) {
	store := initConfigStore()
	sd := NewServiceDiscovery(store)

	createExternalServices([]*networking.ExternalService{httpDNS, tcpDNS}, store, t)

	expectedInstances := []*model.ServiceInstance{
		makeInstance(httpDNS, "us.google.com", 7080, httpDNS.Ports[0], nil),
		makeInstance(httpDNS, "us.google.com", 18080, httpDNS.Ports[1], nil),
		makeInstance(httpDNS, "uk.google.com", 1080, httpDNS.Ports[0], nil),
		makeInstance(httpDNS, "uk.google.com", 8080, httpDNS.Ports[1], nil),
		makeInstance(httpDNS, "de.google.com", 80, httpDNS.Ports[0], map[string]string{"foo": "bar"}),
		makeInstance(httpDNS, "de.google.com", 8080, httpDNS.Ports[1], map[string]string{"foo": "bar"}),
	}

	instances, err := sd.Instances("*.google.com", nil, nil)
	if err != nil {
		t.Errorf("Instances() encountered unexpected error: %v", err)
	}
	sortServiceInstances(instances)
	sortServiceInstances(expectedInstances)
	if err := compare(t, instances, expectedInstances); err != nil {
		t.Error(err)
	}
}

func sortServices(services []*model.Service) {
	sort.Slice(services, func(i, j int) bool { return services[i].Hostname < services[j].Hostname })
	for _, service := range services {
		sortPorts(service.Ports)
	}
}

func sortServiceInstances(instances []*model.ServiceInstance) {
	labelsToSlice := func(labels model.Labels) []string {
		out := make([]string, 0, len(labels))
		for k, v := range labels {
			out = append(out, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(out)
		return out
	}

	sort.Slice(instances, func(i, j int) bool {
		if instances[i].Service.Hostname == instances[j].Service.Hostname {
			if instances[i].Endpoint.Port == instances[j].Endpoint.Port {
				if instances[i].Endpoint.Address == instances[j].Endpoint.Address {
					if len(instances[i].Labels) == len(instances[j].Labels) {
						iLabels := labelsToSlice(instances[i].Labels)
						jLabels := labelsToSlice(instances[j].Labels)
						for k := range iLabels {
							if iLabels[k] < jLabels[k] {
								return true
							}
						}
					}
					return len(instances[i].Labels) < len(instances[j].Labels)
				}
				return instances[i].Endpoint.Address < instances[j].Endpoint.Address
			}
			return instances[i].Endpoint.Port < instances[j].Endpoint.Port
		}
		return instances[i].Service.Hostname < instances[j].Service.Hostname
	})
}

func sortPorts(ports []*model.Port) {
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Port == ports[j].Port {
			if ports[i].Name == ports[j].Name {
				return ports[i].Protocol < ports[j].Protocol
			}
			return ports[i].Name < ports[j].Name
		}
		return ports[i].Port < ports[j].Port
	})
}
