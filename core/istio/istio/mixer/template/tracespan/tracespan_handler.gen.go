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

// THIS FILE IS AUTOMATICALLY GENERATED.

package tracespan

import (
	"context"
	"time"

	"istio.io/istio/mixer/pkg/adapter"
)

// Fully qualified name of the template
const TemplateName = "tracespan"

// Instance is constructed by Mixer for the 'tracespan' template.
//
// TraceSpan represents an individual span within a distributed trace.
//
// When writing the configuration, the value for the fields associated with this template can either be a
// literal or an [expression](https://istio.io/docs/reference/config/mixer/expression-language.html). Please note that if the datatype of a field is not istio.mixer.v1.template.Value,
// then the expression's [inferred type](https://istio.io/docs/reference/config/mixer/expression-language.html#type-checking) must match the datatype of the field.
//
// Example config:
// ```
// apiVersion: "config.istio.io/v1alpha2"
// kind: tracespan
// metadata:
//   name: default
//   namespace: istio-system
// spec:
//   traceId: request.headers["x-b3-traceid"]
//   spanId: request.headers["x-b3-spanid"] | ""
//   parentSpanId: request.headers["x-b3-parentspanid"] | ""
//   spanName: request.path | "/"
//   startTime: request.time
//   endTime: response.time
//   spanTags:
//     http.method: request.method | ""
//     http.status_code: response.code | 200
//     http.url: request.path | ""
//     request.size: request.size | 0
//     response.size: response.size | 0
//     source.ip: source.ip | ip("0.0.0.0")
//     source.service: source.service | ""
//     source.user: source.user | ""
//     source.version: source.labels["version"] | ""
// ```
//
// See also: [Distributed Tracing](https://istio.io/docs/tasks/telemetry/distributed-tracing.html)
// for information on tracing within Istio.
type Instance struct {
	// Name of the instance as specified in configuration.
	Name string

	// Trace ID is the unique identifier for a trace. All spans from the same
	// trace share the same Trace ID.
	//
	// Required.
	TraceId string

	// Span ID is the unique identifier for a span within a trace. It is assigned
	// when the span is created.
	//
	// Optional.
	SpanId string

	// Parent Span ID is the unique identifier for a parent span of this span
	// instance. If this is a root span, then this field MUST be empty.
	//
	// Optional.
	ParentSpanId string

	// Span name is a description of the span's operation.
	//
	// For example, the name can be a qualified method name or a file name
	// and a line number where the operation is called. A best practice is to use
	// the same display name within an application and at the same call point.
	// This makes it easier to correlate spans in different traces.
	//
	// Required.
	SpanName string

	// The start time of the span.
	//
	// Required.
	StartTime time.Time

	// The end time of the span.
	//
	// Required.
	EndTime time.Time

	// Span tags are a set of <key, value> pairs that provide metadata for the
	// entire span. The values can be specified in the form of expressions.
	//
	// Optional.
	SpanTags map[string]interface{}
}

// HandlerBuilder must be implemented by adapters if they want to
// process data associated with the 'tracespan' template.
//
// Mixer uses this interface to call into the adapter at configuration time to configure
// it with adapter-specific configuration as well as all template-specific type information.
type HandlerBuilder interface {
	adapter.HandlerBuilder

	// SetTraceSpanTypes is invoked by Mixer to pass the template-specific Type information for instances that an adapter
	// may receive at runtime. The type information describes the shape of the instance.
	SetTraceSpanTypes(map[string]*Type /*Instance name -> Type*/)
}

// Handler must be implemented by adapter code if it wants to
// process data associated with the 'tracespan' template.
//
// Mixer uses this interface to call into the adapter at request time in order to dispatch
// created instances to the adapter. Adapters take the incoming instances and do what they
// need to achieve their primary function.
//
// The name of each instance can be used as a key into the Type map supplied to the adapter
// at configuration time via the method 'SetTraceSpanTypes'.
// These Type associated with an instance describes the shape of the instance
type Handler interface {
	adapter.Handler

	// HandleTraceSpan is called by Mixer at request time to deliver instances to
	// to an adapter.
	HandleTraceSpan(context.Context, []*Instance) error
}