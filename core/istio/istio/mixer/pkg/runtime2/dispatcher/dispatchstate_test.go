// Copyright 2018 Istio Authors
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

package dispatcher

import (
	"errors"
	"reflect"
	"testing"

	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/pkg/attribute"
	"istio.io/istio/mixer/pkg/runtime2/routing"
)

func TestDispatchStatePool(t *testing.T) {
	session := &session{}
	dest := &routing.Destination{}

	pool := newDispatchStatePool()

	// Prime the pool
	states := make([]*dispatchState, 100)
	for i := 0; i < 100; i++ {
		s := pool.get(nil, nil)
		states[i] = s
	}
	for i := 0; i < 100; i++ {
		pool.put(states[i])
	}

	// test cleaning
	for i := 0; i < 100; i++ {
		s := pool.get(session, dest)
		s.instance = "instanc"
		states[i] = s
	}
	for i := 0; i < 100; i++ {
		pool.put(states[i])
	}

	expected := &dispatchState{}

	for i := 0; i < 100; i++ {
		s := pool.get(nil, nil)
		if !reflect.DeepEqual(s, expected) {
			t.Fatalf("session mismatch '%+v' != '%+v'", s, expected)
		}
	}
}

func TestDispatchState_Clear(t *testing.T) {
	state := &dispatchState{
		instance:    "instance",
		session:     &session{},
		quotaResult: adapter.QuotaResult{Amount: 64},
		checkResult: adapter.CheckResult{ValidUseCount: 32},
		err:         errors.New("err"),
		destination: &routing.Destination{},
		inputBag:    attribute.GetMutableBag(nil),
		outputBag:   attribute.GetMutableBag(nil),
		quotaArgs:   adapter.QuotaArgs{BestEffort: true},
		mapper: func(attrs attribute.Bag) (*attribute.MutableBag, error) {
			return nil, nil
		},
		instances: make([]interface{}, 10),
	}

	state.clear()

	expected := &dispatchState{
		instances: make([]interface{}, 0, 10),
	}

	if !reflect.DeepEqual(state, expected) {
		t.Fail()
	}
	if cap(state.instances) != 10 {
		t.Fail()
	}
	if len(state.instances) != 0 {
		t.Fail()
	}
}
