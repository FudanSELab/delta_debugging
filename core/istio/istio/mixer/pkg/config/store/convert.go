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

package store

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"istio.io/istio/pkg/log"
)

const (
	ruleKind      = "rule"
	selectorField = "selector"
	matchField    = "match"
)

// warnDeprecationAndFix warns users about deprecated fields.
// It maps the field into new name.
func warnDeprecationAndFix(key Key, spec map[string]interface{}) map[string]interface{} {
	if key.Kind != ruleKind {
		return spec
	}
	sel := spec[selectorField]
	if sel == nil {
		return spec
	}
	log.Warnf("Deprecated field 'selector' used in %s. Use 'match' instead.", key)
	spec[matchField] = sel
	delete(spec, selectorField)
	return spec
}

// cloneMessage looks up the kind in the map, and creates a clone of it.
func cloneMessage(kind string, kinds map[string]proto.Message) (proto.Message, error) {
	msg, ok := kinds[kind]
	if !ok {
		return nil, fmt.Errorf("unrecognized kind %s", kind)
	}
	return proto.Clone(msg), nil
}

// convert converts unstructured spec into the target proto.
func convert(key Key, spec map[string]interface{}, target proto.Message) error {
	jsonData, err := json.Marshal(warnDeprecationAndFix(key, spec))
	if err != nil {
		return err
	}
	if err = jsonpb.Unmarshal(bytes.NewReader(jsonData), target); err != nil {
		log.Warnf("%s unable to unmarshal: %s, %s", key, err.Error(), string(jsonData))
	}

	return err
}
