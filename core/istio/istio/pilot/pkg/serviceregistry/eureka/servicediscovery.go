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

package eureka

import (
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
)

// NewServiceDiscovery instantiates an implementation of service discovery for Eureka
func NewServiceDiscovery(client Client) model.ServiceDiscovery {
	return &serviceDiscovery{
		client: client,
	}
}

type serviceDiscovery struct {
	client Client
}

// Services implements a service catalog operation
func (sd *serviceDiscovery) Services() ([]*model.Service, error) {
	apps, err := sd.client.Applications()
	if err != nil {
		log.Warnf("could not list Eureka instances: %v", err)
		return nil, err
	}
	services := convertServices(apps, nil)

	out := make([]*model.Service, 0, len(services))
	for _, service := range services {
		out = append(out, service)
	}
	return out, nil
}

// GetService implements a service catalog operation
func (sd *serviceDiscovery) GetService(hostname string) (*model.Service, error) {
	apps, err := sd.client.Applications()
	if err != nil {
		log.Warnf("could not list Eureka instances: %v", err)
		return nil, err
	}

	services := convertServices(apps, map[string]bool{hostname: true})
	service := services[hostname]
	return service, nil
}

// Instances implements a service catalog operation
func (sd *serviceDiscovery) Instances(hostname string, ports []string,
	tagsList model.LabelsCollection) ([]*model.ServiceInstance, error) {

	apps, err := sd.client.Applications()
	if err != nil {
		log.Warnf("could not list Eureka instances: %v", err)
		return nil, err
	}
	portSet := make(map[string]bool)
	for _, port := range ports {
		portSet[port] = true
	}
	services := convertServices(apps, map[string]bool{hostname: true})

	out := make([]*model.ServiceInstance, 0)
	for _, instance := range convertServiceInstances(services, apps) {
		if !tagsList.HasSubsetOf(instance.Labels) {
			continue
		}

		if len(portSet) > 0 && !portSet[instance.Endpoint.ServicePort.Name] {
			continue
		}

		out = append(out, instance)
	}
	return out, nil
}

// GetProxyServiceInstances returns service instances co-located with a proxy
func (sd *serviceDiscovery) GetProxyServiceInstances(proxy model.Proxy) ([]*model.ServiceInstance, error) {
	apps, err := sd.client.Applications()
	if err != nil {
		log.Warnf("could not list Eureka instances: %v", err)
		return nil, err
	}
	services := convertServices(apps, nil)

	out := make([]*model.ServiceInstance, 0)
	for _, instance := range convertServiceInstances(services, apps) {
		if proxy.IPAddress == instance.Endpoint.Address {
			out = append(out, instance)
		}
	}
	return out, nil
}

// ManagementPorts retries set of health check ports by instance IP.
// This does not apply to Eureka service registry, as Eureka does not
// manage the service instances.
func (sd *serviceDiscovery) ManagementPorts(addr string) model.PortList {
	return nil
}
