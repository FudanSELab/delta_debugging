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

package runtime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	adptTmpl "istio.io/api/mixer/v1/template"
	"istio.io/istio/mixer/pkg/attribute"
	"istio.io/istio/mixer/pkg/expr"
	"istio.io/istio/pkg/log"
)

// Rule represents a runtime view of cpb.Rule.
type Rule struct {
	// Match condition from the original rule.
	match string
	// Actions are stored in runtime format.
	actions map[adptTmpl.TemplateVariety][]*Action
	// Rule is a top level config object and it has a unique name.
	// It is used here for informational purposes.
	name string
	// rtype is gathered from labels.
	rtype ResourceType
}

func (r Rule) String() string {
	return fmt.Sprintf("[name:<%s>, match:<%s>, type:%s, actions: %v",
		r.name, r.match, r.rtype, r.actions)
}

// resolver is the runtime view of the configuration database.
type resolver struct {
	// evaluator evaluates selectors
	evaluator expr.Evaluator

	// identityAttribute defines which configuration scopes apply to a request.
	// default: target.service
	// The value of this attribute is expected to be a hostname of form "svc.$ns.suffix"
	identityAttribute string

	// defaultConfigNamespace defines the namespace that contains configuration defaults for istio.
	// This is distinct from the "default" namespace in K8s.
	// default: istio-default-config
	defaultConfigNamespace string

	// rules in the configuration database keyed by $namespace.
	rules map[string][]*Rule

	// refCount tracks the number requests currently using this
	// configuration. resolver state can be cleaned up when this count is 0.
	refCount int32

	// id of the resolver for debugging.
	id int
}

// newResolver returns a Resolver.
func newResolver(evaluator expr.Evaluator, identityAttribute string, defaultConfigNamespace string,
	rules map[string][]*Rule, id int) *resolver {
	return &resolver{
		evaluator:              evaluator,
		identityAttribute:      identityAttribute,
		defaultConfigNamespace: defaultConfigNamespace,
		rules: rules,
		id:    id,
	}
}

const (
	// DefaultConfigNamespace holds istio wide configuration.
	DefaultConfigNamespace = "istio-system"

	// DefaultIdentityAttribute is attribute that defines config scopes.
	DefaultIdentityAttribute = "destination.service"

	// ContextProtocolAttributeName is the attribute that defines the protocol context.
	ContextProtocolAttributeName = "context.protocol"

	// ContextProtocolTCP defines constant for tcp protocol.
	ContextProtocolTCP = "tcp"

	// expectedResolvedActionsCount is used to preallocate slice for actions.
	expectedResolvedActionsCount = 10
)

// Resolve resolves the in memory configuration to a set of actions based on request attributes.
// Resolution is performed in the following order
// 1. Check rules from the defaultConfigNamespace -- these rules always apply
// 2. Check rules from the target.service namespace
func (r *resolver) Resolve(attrs attribute.Bag, variety adptTmpl.TemplateVariety) (ra Actions, err error) {
	nselected := 0
	target := "unknown"
	var ns string

	start := time.Now()
	// increase refcount just before returning
	// only if there is no error.
	defer func() {
		if err == nil {
			r.incRefCount()
		}
	}()

	// monitoring info
	defer func() {
		lbls := prometheus.Labels{
			targetStr: target,
			errorStr:  strconv.FormatBool(err != nil),
		}
		resolveCounter.With(lbls).Inc()
		resolveDuration.With(lbls).Observe(time.Since(start).Seconds())
		resolveRules.With(lbls).Observe(float64(nselected))
		raLen := 0
		if ra != nil {
			raLen = len(ra.Get())
		}
		resolveActions.With(lbls).Observe(float64(raLen))
	}()

	if target, ns, err = destAndNamespace(attrs, r.identityAttribute); err != nil {
		return nil, err
	}

	// at most this can have 2 elements.
	rulesArr := make([][]*Rule, 0, 2)

	// add default namespace if present
	rulesArr = appendRules(rulesArr, r.rules, r.defaultConfigNamespace)

	// If the destination namespace is different than the default namespace
	// add those rules too
	if r.defaultConfigNamespace != ns {
		rulesArr = appendRules(rulesArr, r.rules, ns)
	} else {
		log.Debugf("Resolve: skipping duplicate namespace %s", ns)
	}

	var res []*Action
	res, nselected, err = r.filterActions(rulesArr, attrs, variety)
	if err != nil {
		return nil, err
	}

	// TODO add dedupe + group actions by handler/template

	ra = &actions{a: res, done: r.decRefCount}
	return ra, nil
}

func appendRules(rulesArr [][]*Rule, rules map[string][]*Rule, ns string) [][]*Rule {
	if r := rules[ns]; r != nil {
		rulesArr = append(rulesArr, r)
	} else {
		log.Debugf("Resolve: no namespace config for %s", ns)
	}
	return rulesArr
}

// destAndNamespace extracts namespace from identity attribute.
func destAndNamespace(attrs attribute.Bag, idAttr string) (dest string, ns string, err error) {
	attr, _ := attrs.Get(idAttr)
	if attr == nil {
		msg := fmt.Sprintf("%s identity not found in attributes%v", idAttr, attrs.Names())
		log.Warnf(msg)
		return "", "", errors.New(msg)
	}

	var ok bool
	if dest, ok = attr.(string); !ok {
		msg := fmt.Sprintf("%s identity must be string: %v", idAttr, attr)
		log.Warnf(msg)
		return "", "", errors.New(msg)
	}
	splits := strings.SplitN(dest, ".", 3) // we only care about service and namespace.
	if len(splits) > 1 {
		ns = splits[1]
	}
	return dest, ns, nil
}

//filterActions filters rules based on template variety and selectors.
func (r *resolver) filterActions(rulesArr [][]*Rule, attrs attribute.Bag,
	variety adptTmpl.TemplateVariety) ([]*Action, int, error) {
	res := make([]*Action, 0, expectedResolvedActionsCount)
	var selected bool
	nselected := 0
	var err error
	ctxProtocol, _ := attrs.Get(ContextProtocolAttributeName)
	tcp := ctxProtocol == ContextProtocolTCP

	for _, rules := range rulesArr {
		for _, rule := range rules {
			act := rule.actions[variety]
			if act == nil { // do not evaluate match if there is no variety specific action there.
				continue
			}
			// default rtype is HTTP + Check|Report|Preprocess
			if tcp != rule.rtype.IsTCP() {
				log.Debugf("filterActions: rule %s removed ctxProtocol=%s, type %s", rule.name, ctxProtocol, rule.rtype)
				continue
			}

			// do not evaluate empty predicates.
			if len(rule.match) != 0 {
				if selected, err = r.evaluator.EvalPredicate(rule.match, attrs); err != nil {
					return nil, 0, err
				}
				if !selected {
					continue
				}
			}
			log.Debugf("filterActions: rule %s selected %v", rule.name, rule.rtype)
			nselected++
			res = append(res, act...)
		}
	}
	return res, nselected, nil
}

func (r *resolver) incRefCount() {
	atomic.AddInt32(&r.refCount, 1)
}

func (r *resolver) decRefCount() {
	atomic.AddInt32(&r.refCount, -1)
}

// actions implements Actions interface.
type actions struct {
	a    []*Action
	done func()
}

func (a *actions) Get() []*Action {
	return a.a
}

func (a *actions) Done() {
	a.done()
}
