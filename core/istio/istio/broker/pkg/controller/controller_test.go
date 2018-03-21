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

package controller

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	brokerconfig "istio.io/api/broker/dev"
	"istio.io/istio/broker/pkg/model/config"
	"istio.io/istio/broker/pkg/model/osb"
)

type testStore struct {
	mockStore  *mockStore
	controller *Controller
}

type mockStore struct {
	config.BrokerConfigStore
	services map[string]*brokerconfig.ServiceClass
	plans    map[string]*brokerconfig.ServicePlan
}

func (m mockStore) ServiceClasses() map[string]*brokerconfig.ServiceClass {
	return m.services
}

func (m mockStore) ServicePlans() map[string]*brokerconfig.ServicePlan {
	return m.plans
}

func (m mockStore) ServicePlansByService(service string) map[string]*brokerconfig.ServicePlan {
	return m.plans
}

func newTestStore(t *testing.T) *testStore {
	ms := &mockStore{}
	return &testStore{
		ms,
		&Controller{ms},
	}
}

func TestCatalog(t *testing.T) {
	r := newTestStore(t)

	sc := &brokerconfig.ServiceClass{
		Deployment: &brokerconfig.Deployment{
			Instance: "productpage",
		},
		Entry: &brokerconfig.CatalogEntry{
			Name:        "istio-bookinfo-productpage",
			Id:          "4395a443-f49a-41b0-8d14-d17294cf612f",
			Description: "A book info service",
		},
	}

	sp := &brokerconfig.ServicePlan{
		Services: []string{
			"service-class/default/productpage-service-class",
		},
		Plan: &brokerconfig.CatalogPlan{
			Name:        "istio-yearly",
			Id:          "cdd76b03-a28b-4638-b4e2-19ee44b36db7",
			Description: "yearly subscription",
		},
	}

	cases := []struct {
		name         string
		mockServices map[string]*brokerconfig.ServiceClass
		mockPlans    map[string]*brokerconfig.ServicePlan
		want         *osb.Catalog
	}{
		{
			name: "success test",
			mockServices: map[string]*brokerconfig.ServiceClass{
				"service-class/default/productpage-service-class": sc,
			},
			mockPlans: map[string]*brokerconfig.ServicePlan{
				"service-plan/default/istio-yearly": sp,
			},
			want: &osb.Catalog{
				Services: []osb.Service{
					{
						Name:        "istio-bookinfo-productpage",
						ID:          "4395a443-f49a-41b0-8d14-d17294cf612f",
						Description: "A book info service",
						Plans: []osb.ServicePlan{
							{
								Name:        "istio-yearly",
								ID:          "cdd76b03-a28b-4638-b4e2-19ee44b36db7",
								Description: "yearly subscription"}}}}},
		},
	}
	for _, c := range cases {
		r.mockStore.services = c.mockServices
		r.mockStore.plans = c.mockPlans
		if got := r.controller.catalog(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v failed: \ngot %+vwant %+v", c.name, spew.Sdump(got), spew.Sdump(c.want))
		}
	}
}
