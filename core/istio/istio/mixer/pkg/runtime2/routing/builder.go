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

// Package routing implements a routing table for resolving incoming requests to handlers. The table data model
// is structured for efficient use by the runtime code during actual dispatch. At a high-level, the structure
// of table is as follows:
//
// Table:               map[variety]varietyTable
// varietyTable:        map[namespace]NamespaceTable
// NamespaceTable:      list(Destination)
// Destination:         unique(handler&template) + list(InstanceGroup)
// InstanceGroup:       condition + list(InstanceBuilders) + list(OutputMappers)
//
// The call into table.GetDestinations performs a lookup on the first map by the variety (i.e. quota, check,
// report, apa etc.), followed by a lookup on the second map for the namespace, and a NamespaceTable struct
// is returned.
//
// The returned NamespaceTable holds all the handlers that should be dispatched to, along with conditions and
// builders for the instances. These include handlers that were defined for the namespace of the request, as
// well as the handlers from the default namespace. If there were no explicit rules in the request's namespace,
// then only the handlers from the default namespace is applied. Similarly, if the request is for the default
// namespace, then only the handlers from the default namespace is applied.
//
// Beneath the namespace layer, the same handler can appear multiple times in this list for each template that
// is supported by the handler. This helps caller to ensure that each dispatch to the handler will use a unique
// template.
//
// The client code is expected to work as follows:
// - Call GetDestinations(variety, namespace) to get a NamespaceTable.
// - Go through the list of entries in the NamespaceTable.
// - For each entry begin a dispatch session to the associated handler.
// - Go through the InstanceGroup
// - For each InstanceGroup, check the condition and see if the inputs/outputs apply.
// - If applies, then call InstanceBuilders to create instances
// - Depending on the variety, either aggregate all instances in the group, and send them all at once, or
//   dispatch for every instance individually to the adapter.
//
package routing

import (
	"fmt"
	"strings"

	descriptor "istio.io/api/mixer/v1/config/descriptor"
	tpb "istio.io/api/mixer/v1/template"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/pkg/expr"
	"istio.io/istio/mixer/pkg/il/compiled"
	"istio.io/istio/mixer/pkg/runtime2/config"
	"istio.io/istio/mixer/pkg/runtime2/handler"
	"istio.io/istio/mixer/pkg/template"
	"istio.io/istio/pkg/log"
)

// builder keeps the ephemeral state while the routing table is built.
type builder struct {
	// table that is being built.
	table                  *Table
	handlers               *handler.Table
	expb                   *compiled.ExpressionBuilder
	defaultConfigNamespace string

	// id counter for assigning ids to various items in the hierarchy. These reference into the debug
	// information.
	nextIDCounter uint32

	// Ephemeral data that can also be used as debugging info.

	// match condition sets by the input set id.
	matchesByID map[uint32]string

	// instanceName set of builders by the input set.
	instanceNamesByID map[uint32][]string

	// InstanceBuilderFns by instance name.
	builders map[string]template.InstanceBuilderFn

	// OutputMapperFns by instance name.
	mappers map[string]template.OutputMapperFn

	// compiled.Expressions by canonicalized rule match clauses
	expressions map[string]compiled.Expression
}

// BuildTable builds and returns a routing table. If debugInfo is set, the returned table will have debugging information
// attached, which will show up in String() call.
func BuildTable(
	handlers *handler.Table,
	config *config.Snapshot,
	expb *compiled.ExpressionBuilder,
	defaultConfigNamespace string,
	debugInfo bool) *Table {

	b := &builder{

		table: &Table{
			id:      config.ID,
			entries: make(map[tpb.TemplateVariety]*varietyTable, 4),
		},

		handlers: handlers,
		expb:     expb,
		defaultConfigNamespace: defaultConfigNamespace,
		nextIDCounter:          1,

		matchesByID:       make(map[uint32]string, len(config.Rules)),
		instanceNamesByID: make(map[uint32][]string, len(config.Instances)),

		builders:    make(map[string]template.InstanceBuilderFn, len(config.Instances)),
		mappers:     make(map[string]template.OutputMapperFn, len(config.Instances)),
		expressions: make(map[string]compiled.Expression, len(config.Rules)),
	}

	b.build(config)

	if debugInfo {
		b.table.debugInfo = &tableDebugInfo{
			matchesByID:       b.matchesByID,
			instanceNamesByID: b.instanceNamesByID,
		}
	}

	return b.table
}

func (b *builder) nextID() uint32 {
	id := b.nextIDCounter
	b.nextIDCounter++
	return id
}

func (b *builder) build(config *config.Snapshot) {

	for _, rule := range config.Rules {

		// Create a compiled expression for the rule condition first.
		condition, err := b.getConditionExpression(rule)
		if err != nil {
			log.Warnf("Unable to compile match condition expression: '%v', rule='%s', expression='%s'",
				err, rule.Name, rule.Match)
			config.Counters.MatchErrors.Inc()
			// Skip the rule
			continue
		}

		// For each action, find unique instances to use, and add entries to the map.
		for i, action := range rule.Actions {

			// Find the matching handler.
			handlerName := action.Handler.Name
			entry, found := b.handlers.Get(handlerName)
			if !found {
				// This can happen if we cannot initialize a handler, even if the config itself self-consistent.
				log.Warnf("Unable to find a handler for action. rule[action]='%s[%d]', handler='%s'",
					rule.Name, i, handlerName)

				config.Counters.UnsatisfiedActionHandlers.Inc()
				// Skip the rule
				continue
			}

			for _, instance := range action.Instances {
				// get the instance mapper and builder for this instance. Mapper is used by APA instances
				// to map the instance result back to attributes.
				builder, mapper, err := b.getBuilderAndMapper(config.Attributes, instance)
				if err != nil {
					log.Warnf("Unable to create builder/mapper for instance: instance='%s', err='%v'", instance.Name, err)
					continue
				}

				b.add(rule.Namespace, instance.Template, entry.Adapter, entry.Handler, condition, builder, mapper,
					entry.Name, instance.Name, rule.Match, rule.ResourceType)
			}
		}
	}

	// Capture the default namespace rule set and flatten all default namespace rule into other namespace tables for
	// faster processing.
	for _, vTable := range b.table.entries {
		defaultSet, found := vTable.entries[b.defaultConfigNamespace]
		if !found {
			log.Warnf("No destination sets found for the default namespace '%s'.", b.defaultConfigNamespace)
			defaultSet = emptyDestinations
		}
		// Set the default rule set for the variety.
		vTable.defaultSet = defaultSet

		if defaultSet.Count() != 0 {
			// Prefix all namespace destinations with the destinations from the default namespace.
			for namespace, set := range vTable.entries {
				if namespace == b.defaultConfigNamespace {
					// Skip the default namespace itself
					continue
				}

				set.entries = append(defaultSet.entries, set.entries...)
			}
		}
	}
}

// get or create a builder and a mapper for the given instance. The mapper is created only if the template
// is an attribute generator.
func (b *builder) getBuilderAndMapper(
	finder expr.AttributeDescriptorFinder,
	instance *config.Instance) (template.InstanceBuilderFn, template.OutputMapperFn, error) {
	var err error

	t := instance.Template

	builder := b.builders[instance.Name]
	if builder == nil {
		if builder, err = t.CreateInstanceBuilder(instance.Name, instance.Params, b.expb); err != nil {
			return nil, nil, err
		}
		b.builders[instance.Name] = builder
	}

	var mapper template.OutputMapperFn
	if t.Variety == tpb.TEMPLATE_VARIETY_ATTRIBUTE_GENERATOR {
		mapper = b.mappers[instance.Name]
		if mapper == nil {
			var expressions map[string]compiled.Expression
			if expressions, err = t.CreateOutputExpressions(instance.Params, finder, b.expb); err != nil {
				return nil, nil, err
			}
			mapper = template.NewOutputMapperFn(expressions)
		}

		b.mappers[instance.Name] = mapper
	}

	return builder, mapper, nil
}

// get or create a compiled.Expression for the rule's match clause, if necessary.
func (b *builder) getConditionExpression(rule *config.Rule) (compiled.Expression, error) {
	text := strings.TrimSpace(rule.Match)

	if text == "" {
		return nil, nil
	}

	// Minor optimization for a simple case.
	if text == "true" {
		return nil, nil
	}

	expression := b.expressions[text]
	if expression == nil {
		var err error
		var t descriptor.ValueType
		if expression, t, err = b.expb.Compile(text); err != nil {
			return nil, err
		}
		if t != descriptor.BOOL {
			return nil, fmt.Errorf("expression does not return a boolean: '%s'", text)
		}

		b.expressions[text] = expression
	}

	return expression, nil
}

func (b *builder) add(
	namespace string,
	t *template.Info,
	a *adapter.Info,
	handler adapter.Handler,
	condition compiled.Expression,
	builder template.InstanceBuilderFn,
	mapper template.OutputMapperFn,
	handlerName string,
	instanceName string,
	matchText string,
	resourceType config.ResourceType) {

	// Find or create the variety entry.
	byVariety, found := b.table.entries[t.Variety]
	if !found {
		byVariety = &varietyTable{
			entries: make(map[string]*NamespaceTable),
		}
		b.table.entries[t.Variety] = byVariety
	}

	// Find or create the namespace entry.
	byNamespace, found := byVariety.entries[namespace]
	if !found {
		byNamespace = &NamespaceTable{
			entries: []*Destination{},
		}
		byVariety.entries[namespace] = byNamespace
	}

	// Find or create the handler&template entry.
	var byHandler *Destination
	for _, d := range byNamespace.Entries() {
		if d.Handler == handler && d.Template.Name == t.Name {
			byHandler = d
			break
		}
	}

	if byHandler == nil {
		byHandler = &Destination{
			id:             b.nextID(),
			Handler:        handler,
			FriendlyName:   fmt.Sprintf("%s:%s(%s)", t.Name, handlerName, a.Name),
			HandlerName:    handlerName,
			AdapterName:    a.Name,
			Template:       t,
			InstanceGroups: []*InstanceGroup{},
			Counters:       newDestinationCounters(t.Name, handlerName, a.Name),
		}
		byNamespace.entries = append(byNamespace.entries, byHandler)
	}

	// TODO(Issue #2690): We should dedupe instances that are being dispatched to a particular handler.

	// Find or create the input set.
	var instanceGroup *InstanceGroup
	for _, set := range byHandler.InstanceGroups {
		// Try to find an input set to place the entry by comparing the compiled expression and resource type.
		// This doesn't flatten across all actions, but only for actions coming from the same rule. We can
		// flatten based on the expression text as well.
		if set.Condition == condition && set.ResourceType == resourceType {
			instanceGroup = set
			break
		}
	}

	if instanceGroup == nil {
		instanceGroup = &InstanceGroup{
			id:           b.nextID(),
			Condition:    condition,
			ResourceType: resourceType,
			Builders:     []template.InstanceBuilderFn{},
			Mappers:      []template.OutputMapperFn{},
		}
		byHandler.InstanceGroups = append(byHandler.InstanceGroups, instanceGroup)

		if matchText != "" {
			b.matchesByID[instanceGroup.id] = matchText
		}

		// Create a slot in the debug info for storing the instance names for this input-set.
		instanceNames, found := b.instanceNamesByID[instanceGroup.id]
		if !found {
			instanceNames = make([]string, 0, 1)
		}
		b.instanceNamesByID[instanceGroup.id] = instanceNames
	}

	// Append the builder & mapper.
	instanceGroup.Builders = append(instanceGroup.Builders, builder)

	if mapper != nil {
		instanceGroup.Mappers = append(instanceGroup.Mappers, mapper)
	}

	// Recalculate the maximum number of instances that can be created.
	byHandler.recalculateMaxInstances()

	// record the instance name for this id.
	instanceNames := b.instanceNamesByID[instanceGroup.id]
	instanceNames = append(instanceNames, instanceName)
	b.instanceNamesByID[instanceGroup.id] = instanceNames
}
