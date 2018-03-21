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

package config

import (
	"errors"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	brokerconfig "istio.io/api/broker/dev"
)

type testStore struct {
	mock  *mockStore
	store BrokerConfigStore
}

type mockStore struct {
	Store
	entries []Entry
	err     error
}

func (m mockStore) List(typ, namespace string) ([]Entry, error) {
	return m.entries, m.err
}

func newTestStore() *testStore {
	ms := &mockStore{}
	return &testStore{
		mock:  ms,
		store: MakeBrokerConfigStore(ms),
	}
}

func TestServiceClasses(t *testing.T) {
	r := newTestStore()

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

	cases := []struct {
		name        string
		mockError   error
		mockEntries []Entry
		want        map[string]*brokerconfig.ServiceClass
	}{
		{
			name:      "list success test",
			mockError: nil,
			mockEntries: []Entry{
				{Meta: Meta{Name: "productpage-service-class"}, Spec: sc},
			},
			want: map[string]*brokerconfig.ServiceClass{
				"//productpage-service-class": sc,
			},
		},
		{
			name:      "list with error test",
			mockError: errors.New("timeout"),
			mockEntries: []Entry{
				{Meta: Meta{Name: "productpage-service-class"}, Spec: sc},
			},
			want: map[string]*brokerconfig.ServiceClass{},
		},
	}
	for _, c := range cases {
		r.mock.entries = c.mockEntries
		r.mock.err = c.mockError
		if got := r.store.ServiceClasses(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v failed: \ngot %+vwant %+v", c.name, spew.Sdump(got), spew.Sdump(c.want))
		}
	}
}

func TestServicePlans(t *testing.T) {
	r := newTestStore()

	sp := &brokerconfig.ServicePlan{
		Plan: &brokerconfig.CatalogPlan{
			Name:        "istio-monthly",
			Id:          "58646b26-867a-4954-a1b9-233dac07815b",
			Description: "monthly subscription",
		},
		Services: []string{
			"productpage-service-class",
		},
	}

	cases := []struct {
		name        string
		mockError   error
		mockEntries []Entry
		want        map[string]*brokerconfig.ServicePlan
	}{
		{
			name:      "list success test",
			mockError: nil,
			mockEntries: []Entry{
				{Meta: Meta{Name: "monthly-service-plan"}, Spec: sp},
			},
			want: map[string]*brokerconfig.ServicePlan{
				"//monthly-service-plan": sp,
			},
		},
		{
			name:      "list with error test",
			mockError: errors.New("timeout"),
			mockEntries: []Entry{
				{Meta: Meta{Name: "monthly-service-plan"}, Spec: sp},
			},
			want: map[string]*brokerconfig.ServicePlan{},
		},
	}
	for _, c := range cases {
		r.mock.entries = c.mockEntries
		r.mock.err = c.mockError
		if got := r.store.ServicePlans(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v failed: \ngot %+vwant %+v", c.name, spew.Sdump(got), spew.Sdump(c.want))
		}
	}
}

func TestServicePlansByService(t *testing.T) {
	r := newTestStore()

	sp := &brokerconfig.ServicePlan{
		Plan: &brokerconfig.CatalogPlan{
			Name:        "istio-monthly",
			Id:          "58646b26-867a-4954-a1b9-233dac07815b",
			Description: "monthly subscription",
		},
		Services: []string{
			"productpage-service-class",
		},
	}

	cases := []struct {
		name        string
		mockError   error
		mockEntries []Entry
		input       string
		want        map[string]*brokerconfig.ServicePlan
	}{
		{
			name:      "list success test",
			mockError: nil,
			mockEntries: []Entry{
				{Meta: Meta{Name: "monthly-service-plan"}, Spec: sp},
			},
			input: "productpage-service-class",
			want: map[string]*brokerconfig.ServicePlan{
				"//monthly-service-plan": sp,
			},
		},
		{
			name:      "list with error test",
			mockError: errors.New("timeout"),
			mockEntries: []Entry{
				{Meta: Meta{Name: "monthly-service-plan"}, Spec: sp},
			},
			input: "productpage-service-class",
			want:  map[string]*brokerconfig.ServicePlan{},
		},
		{
			name:      "cannot find",
			mockError: errors.New("timeout"),
			mockEntries: []Entry{
				{Meta: Meta{Name: "monthly-service-plan"}, Spec: sp},
			},
			input: "other-service-class",
			want:  map[string]*brokerconfig.ServicePlan{},
		},
	}
	for _, c := range cases {
		r.mock.entries = c.mockEntries
		r.mock.err = c.mockError
		if got := r.store.ServicePlansByService(c.input); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v failed: \ngot %+vwant %+v", c.name, spew.Sdump(got), spew.Sdump(c.want))
		}
	}
}
