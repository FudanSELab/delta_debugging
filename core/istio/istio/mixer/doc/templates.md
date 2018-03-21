# Istio Mixer: Template Developer’s Guide

Templates are a foundational building block of the Mixer architecture. This document
describes the template format and how to extend Mixer with custom templates.

- [Background](#background)
- [Template format](#template-format)
- [Supported field names](#supported-field-names)
- [Supported field types](#supported-field-types)
- [Adding a template to Mixer](#adding-a-template-to-mixer)
- [Template evolution compatibility](#template-evolution-compatibility)

## Background

Mixer is an attribute processing machine. It ingests attributes coming from the proxy and responds by
triggering calls into a suite of adapters. These adapters in turn communicate with infrastructure
backends that offer a variety of capabilities such as logging, monitoring, quotas, ACL checking, and more.
The operator that configures Istio controls what Mixer
does with incoming attributes and which adapters are called as a result.

![Attribute Processing Machine](./img/mixer%20architecture.svg) 

Mixer categorizes adapters based on the type of data they consume. For example, there are metric adapters, logging
adapters, quota adapters, access control adapters, and more. The type of data Mixer delivers to individual adapters
depends on each adapter’s category. All adapters of a given category receive the same type of data. For example,
metric adapters all receive metrics at runtime.

Mixer determines the type of data that each category of adapter processes at runtime using *templates*. A template determines
the data adapters receive, as well as the instances operators must create in order to use an adapter.
The [Adapter Developer's Guide](./adapters.md#template-overview) explains how templates are automatically transformed into Go
structs and interfaces that can be used by adapter developers, and into config definitions that can be used by operators. 

Mixer includes a number of [canonical templates](https://github.com/istio/istio/tree/master/mixer/template) which cover
most of the anticipated workloads that Mixer is expected to be used with. However, the set of supported templates can
readily be extended in order to support emerging usage scenarios. Note that it’s preferable to use existing templates
when possible as it tends to deliver a better end-to-end story for the ecosystem by making configuration state more portable 
between adapters.

This document describes the simple rules used to create templates for Mixer.
The [Adapter Developer’s Guide](https://github.com/istio/istio/blob/master/mixer/doc/adapters.md) describes how to use templates to
create adapters.

## Template format

Templates are expressed using the protobuf syntax. They are generally fairly simple data structures. An an example, here is the `listentry` template:

```proto
syntax = "proto3";

package listEntry;

import "mixer/adapter/model/v1beta1/extensions.proto";

option (istio.mixer.adapter.model.v1beta1.template_variety) = TEMPLATE_VARIETY_CHECK;

// ListEntry is used to verify the presence/absence of a string
// within a list.
//
// When writing the configuration, the value for the fields associated with this template can either be a
// literal or an [expression](https://istio.io/docs/reference/config/mixer/expression-language.html). Please note that if the datatype of a field is not istio.mixer.adapter.model.v1beta1.Value,
// then the expression's [inferred type](https://istio.io/docs/reference/config/mixer/expression-language.html#type-checking) must match the datatype of the field.
//
// Example config:
// 
// apiVersion: "config.istio.io/v1alpha2"
// kind: listentry
// metadata:
//   name: appversion
//   namespace: istio-system
// spec:
//   value: source.labels["version"]
message Template {
    // Specifies the entry to verify in the list.
    string value = 1;
}
```

The interesting parts of this definition are:

- **Package Name**. The package name determines the name by which the template will be known, both by
adapter authors and operators creating instances of that template. 
The name should be written in camelCase such that generated Go artifacts
are more readable.

- **Variety**. The variety of a template determines at what point in the Mixer's processing pipeline the
adapters that implement the template will be called. This can be CHECK, REPORT, QUOTA, or ATTRIBUTE_GENERATOR.

- **Message Name**. The name of the template message should always be `Template`.

- **Fields**. The fields represent the general shape of the data that will be delivered to the adapters at
runtime. The operator will need to populate these fields via configuration.

- **Comments** The comment on the `Template` message is used to generate documentation that is used by both adapter developers
as well as operators. Therefore, the comment should contain an example for the operator to use when writing configuration as well
as a good description expressing the intent of the template.

## OutputTemplate format
Message `OutputTemplate` is only applicable if the variety of the template is `ATTRIBUTE_GENERATOR`.
While message `Template` defines the type of data that gets passed to the adapter, the `OutputTemplate` defines type of data
that is returned from an attribute generating adapter. Attribute generating adapters are called with the input 'Template'
instance before other adapters are invoked and are responsible for instantiating and
returning an `instance` of `OutputTemplate`. The operator uses these output values to create more attributes which
can then be used to create instances of other non `ATTRIBUTE_GENERATOR` template variety. `OutputTemplate`s are
generally fairly simple data structures. An an example, here is the `kubernetes` template that outputs various
kubernetes environment specific attributes:

```proto
// OutputTemplate refers to the output from the adapter. It is used inside the attribute_binding section of the config
// to assign values to the generated attributes using the `$out.<field name of the OutputTemplate>` syntax.
message OutputTemplate {
    // Refers to source pod ip address. attribute_bindings can refer to this field using $out.source_pod_ip
    istio.mixer.adapter.model.v1beta1.IPAddress source_pod_ip = 1;

    // Refers to source pod name. attribute_bindings can refer to this field using $out.source_pod_name
    string source_pod_name = 2;

    // Refers to source pod labels. attribute_bindings can refer to this field using $out.source_labels
    map<string, string> source_labels = 3;

    // Refers to source pod namespace. attribute_bindings can refer to this field using $out.source_namespace
    string source_namespace = 4;

    // Refers to source service. attribute_bindings can refer to this field using $out.source_service
    string source_service = 5;

    // Refers to source pod service account name. attribute_bindings can refer to this field using $out.source_service_account_name
    string source_service_account_name = 6;

    // Refers to source pod host ip address. attribute_bindings can refer to this field using $out.source_host_ip
    istio.mixer.adapter.model.v1beta1.IPAddress source_host_ip = 7;


    // Refers to destination pod ip address. attribute_bindings can refer to this field using $out.destination_pod_ip
    istio.mixer.adapter.model.v1beta1.IPAddress destination_pod_ip = 8;

    // Refers to destination pod name. attribute_bindings can refer to this field using $out.destination_pod_name
    string destination_pod_name = 9;

    // Refers to destination pod labels. attribute_bindings can refer to this field using $out.destination_labels
    map<string, string> destination_labels = 10;

    // Refers to destination pod namespace. attribute_bindings can refer to this field using $out.destination_namespace
    string destination_namespace = 11;

    // Refers to destination service. attribute_bindings can refer to this field using $out.destination_service
    string destination_service = 12;

    // Refers to destination pod service account name. attribute_bindings can refer to this field using $out.destination_service_account_name
    string destination_service_account_name = 13;

    // Refers to destination pod host ip address. attribute_bindings can refer to this field using $out.destination_host_ip
    istio.mixer.adapter.model.v1beta1.IPAddress destination_host_ip = 14;

    ...
}
```

## Supported field names

Template fields can have any valid protobuf field name, except for the reserved name `name`. Field
name should follow the normal protobuf naming convention of snake_case.

## Supported field types

Templates currently only support a subset of the full protobuf type system. Following types can be used for fields in a
template. Below table also shows the corresponding Golang type delivered to the adapters:

Template field type | Golang type
--- | ---
`string` | `string`
`int64` | `int64`
`double` | `float64`
`bool` | `bool`
`istio.mixer.adapter.model.v1beta1.TimeStamp` | `time.Time`
`istio.mixer.adapter.model.v1beta1.Duration` | `time.Duration`
`istio.mixer.adapter.model.v1beta1.IPAddress` | `net.IP`
`istio.mixer.adapter.model.v1beta1.DNSName` | `adapter.DNSName`
`istio.mixer.adapter.model.v1beta1.Value` | `interface{}`
`map<string, string>` | `map[string]string`
`map<string, int64>` | `map[string]int64`
`map<string, double>` | `map[string]float64`
`map<string, bool>` | `map[string]bool`
`map<string, istio.mixer.adapter.model.v1beta1.TimeStamp>` | `map[string]time.Time`
`map<string, istio.mixer.adapter.model.v1beta1.Duration>` | `map[string]time.Duration`
`map<string, istio.mixer.adapter.model.v1beta1.IPAddress>` | `map[string]net.IP`
`map<string, istio.mixer.adapter.model.v1beta1.DNSName>` | `map[string]adapter.DNSName`
`map<string, istio.mixer.adapter.model.v1beta1.Value>` | `map[string]interface{}`

There is currently no support for nested messages, enums, `oneof`, and `repeated`.

The type `istio.mixer.adapter.model.v1beta1.Value` has a special meaning. Use of this type
tells Mixer that the associated value can be any of the supported attribute
types which are defined by the [ValueType](https://github.com/istio/api/blob/master/mixer/v1/config/descriptor/value_type.proto)
enum. The specific type that will be used at runtime depends on the configuration the operator writes.
Adapters are told what these types are at [configuration time](./adapters.md##adapter-lifecycle) so they can prepare
themselves accordingly.

There is currently no support for nested messages, enums, `oneof`, and `repeated`.

The type `istio.mixer.adapter.model.v1beta1.Value` has a special meaning. Use of this type
tells Mixer that the associated value can be any of the supported attribute
types which are defined by the [ValueType](https://github.com/istio/api/blob/master/mixer/v1/config/descriptor/value_type.proto)
enum. The specific type that will be used at runtime depends on the configuration the operator writes.
Adapters are told what these types are at [configuration time](./adapters.md##adapter-lifecycle) so they can prepare
themselves accordingly.

## Adding a template to Mixer

Templates are statically linked into Mixer. To add or modify a template, it is therefore necessary to produce a new
Mixer binary.

** TBD **

## Template evolution compatibility

Here's what can be done to a template and remain backward compatible with existing adapters and operator configurations that use the
template:

- Adding a field

The following changes cannot generally be made while maintaining compatibility:

- Renaming the template
- Changing the template's variety
- Removing a field
- Renaming a field
- Changing the type of a field
