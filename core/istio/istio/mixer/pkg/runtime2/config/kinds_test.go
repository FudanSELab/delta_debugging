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

package config

import (
	"reflect"
	"testing"

	"github.com/gogo/protobuf/proto"

	cpb "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/pkg/template"
)

func TestKindMap(t *testing.T) {
	ti := map[string]*template.Info{
		"t1": {
			CtrCfg: &cpb.Instance{},
		},
	}
	ai := map[string]*adapter.Info{
		"a1": {
			DefaultConfig: &cpb.Handler{},
		},
	}

	km := KindMap(ai, ti)

	want := map[string]proto.Message{
		"t1":                  &cpb.Instance{},
		"a1":                  &cpb.Handler{},
		RulesKind:             &cpb.Rule{},
		AttributeManifestKind: &cpb.AttributeManifest{},
	}

	if !reflect.DeepEqual(km, want) {
		t.Fatalf("Got %v\nwant %v", km, want)
	}
}
