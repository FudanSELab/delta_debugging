// Copyright 2017 Istio Authors.
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

package noop // import "istio.io/istio/mixer/adapter/noop"

// NOTE: This adapter will eventually be auto-generated so that it automatically supports all templates
//       known to Mixer. For now, it's manually curated.

import (
	"context"
	"time"

	rpc "github.com/gogo/googleapis/google/rpc"
	"github.com/gogo/protobuf/types"

	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/authorization"
	"istio.io/istio/mixer/template/checknothing"
	"istio.io/istio/mixer/template/listentry"
	"istio.io/istio/mixer/template/logentry"
	"istio.io/istio/mixer/template/metric"
	"istio.io/istio/mixer/template/quota"
	"istio.io/istio/mixer/template/reportnothing"
	"istio.io/istio/mixer/template/tracespan"
)

type handler struct{}

var checkResult = adapter.CheckResult{
	Status:        rpc.Status{Code: int32(rpc.OK)},
	ValidDuration: 1000000000 * time.Second,
	ValidUseCount: 1000000000,
}

func (*handler) HandleAuthorization(context.Context, *authorization.Instance) (adapter.CheckResult, error) {
	return checkResult, nil
}

func (*handler) HandleCheckNothing(context.Context, *checknothing.Instance) (adapter.CheckResult, error) {
	return checkResult, nil
}

func (*handler) HandleListEntry(context.Context, *listentry.Instance) (adapter.CheckResult, error) {
	return checkResult, nil
}

func (*handler) HandleLogEntry(context.Context, []*logentry.Instance) error {
	return nil
}

func (*handler) HandleMetric(context.Context, []*metric.Instance) error {
	return nil
}

func (*handler) HandleQuota(ctx context.Context, _ *quota.Instance, args adapter.QuotaArgs) (adapter.QuotaResult, error) {
	return adapter.QuotaResult{
			ValidDuration: 1000000000 * time.Second,
			Amount:        args.QuotaAmount,
		},
		nil
}

func (*handler) HandleReportNothing(context.Context, []*reportnothing.Instance) error {
	return nil
}

func (*handler) HandleTraceSpan(context.Context, []*tracespan.Instance) error {
	return nil
}

func (*handler) Close() error { return nil }

////////////////// Config //////////////////////////

// GetInfo returns the Info associated with this adapter implementation.
func GetInfo() adapter.Info {
	return adapter.Info{
		Name:        "noop",
		Impl:        "istio.io/istio/mixer/adapter/noop",
		Description: "Does nothing (useful for testing)",
		SupportedTemplates: []string{
			authorization.TemplateName,
			checknothing.TemplateName,
			reportnothing.TemplateName,
			listentry.TemplateName,
			logentry.TemplateName,
			metric.TemplateName,
			quota.TemplateName,
			tracespan.TemplateName,
		},
		DefaultConfig: &types.Empty{},

		NewBuilder: func() adapter.HandlerBuilder { return &builder{} },
	}
}

type builder struct{}

func (*builder) SetCheckNothingTypes(map[string]*checknothing.Type)   {}
func (*builder) SetAuthorizationTypes(map[string]*authorization.Type) {}
func (*builder) SetReportNothingTypes(map[string]*reportnothing.Type) {}
func (*builder) SetListEntryTypes(map[string]*listentry.Type)         {}
func (*builder) SetLogEntryTypes(map[string]*logentry.Type)           {}
func (*builder) SetMetricTypes(map[string]*metric.Type)               {}
func (*builder) SetQuotaTypes(map[string]*quota.Type)                 {}
func (*builder) SetTraceSpanTypes(map[string]*tracespan.Type)         {}
func (*builder) SetAdapterConfig(adapter.Config)                      {}
func (*builder) Validate() (ce *adapter.ConfigErrors)                 { return }

func (b *builder) Build(context context.Context, env adapter.Env) (adapter.Handler, error) {
	return &handler{}, nil
}
