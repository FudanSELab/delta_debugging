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

import "istio.io/istio/pilot/pkg/model"

type serviceAccounts struct {
}

// NewServiceAccounts instantiates the Eureka service account interface
func NewServiceAccounts() model.ServiceAccounts {
	return &serviceAccounts{}
}

// GetIstioServiceAccounts implements model.ServiceAccounts operation TODO
func (sa *serviceAccounts) GetIstioServiceAccounts(hostname string, ports []string) []string {
	//for external services, there is no istio auth, no service accounts, etc. It is just a
	// service, with service instances, and dns.
	return []string{}
}
