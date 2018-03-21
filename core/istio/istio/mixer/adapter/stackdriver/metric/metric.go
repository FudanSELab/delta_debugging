// Copyright 2017 the Istio Authors.
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

package metric

import (
	"context"
	"fmt"
	"sync"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes"
	gax "github.com/googleapis/gax-go"
	xcontext "golang.org/x/net/context"
	labelpb "google.golang.org/genproto/googleapis/api/label"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	descriptor "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/adapter/stackdriver/config"
	"istio.io/istio/mixer/adapter/stackdriver/helper"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/metric"
)

// TODO: implement adapter validation
// TODO: change batching to be size aware: right now we batch and send data to stackdriver based on only a ticker.
// Ideally we'd also size our buffer and send data whenever we hit the size limit or config.push_interval time has passed
// since the last push.
// TODO: today we start a ticker per aspect instance, each keeping an independent data set it pushes to SD. This needs to
// be promoted up to the builder, which will hold a buffer that all aspects write in to, with a single ticker/loop responsible
// for pushing the data from all aspect instances.

type (

	// createClientFunc abstracts over the creation of the stackdriver client to enable network-less testing.
	createClientFunc func(*config.Params) (*monitoring.MetricClient, error)

	// pushFunc abstracts over client.CreateTimeSeries for testing
	pushFunc func(ctx xcontext.Context, req *monitoringpb.CreateTimeSeriesRequest, opts ...gax.CallOption) error

	builder struct {
		createClient createClientFunc
		metrics      map[string]*metric.Type
		cfg          *config.Params
	}

	info struct {
		ttype string
		vtype descriptor.ValueType
		minfo *config.Params_MetricInfo
	}

	handler struct {
		l   adapter.Logger
		now func() time.Time // used to control time in tests

		projectID  string
		metricInfo map[string]info
		client     bufferedClient
		// We hold a ref for cleanup during Close()
		ticker *time.Ticker
	}
)

const (
	// From https://github.com/GoogleCloudPlatform/golang-samples/blob/master/monitoring/custommetric/custommetric.go
	customMetricPrefix = "custom.googleapis.com/"
)

var (
	// TODO: evaluate how we actually want to do this mapping - this first stab w/ everything as String probably
	// isn't what we really want.
	// The better path forward is probably to constrain the input types and err on bad combos.
	labelMap = map[descriptor.ValueType]labelpb.LabelDescriptor_ValueType{
		descriptor.STRING:        labelpb.LabelDescriptor_STRING,
		descriptor.INT64:         labelpb.LabelDescriptor_INT64,
		descriptor.DOUBLE:        labelpb.LabelDescriptor_INT64,
		descriptor.BOOL:          labelpb.LabelDescriptor_BOOL,
		descriptor.TIMESTAMP:     labelpb.LabelDescriptor_INT64,
		descriptor.IP_ADDRESS:    labelpb.LabelDescriptor_STRING,
		descriptor.EMAIL_ADDRESS: labelpb.LabelDescriptor_STRING,
		descriptor.URI:           labelpb.LabelDescriptor_STRING,
		descriptor.DNS_NAME:      labelpb.LabelDescriptor_STRING,
		descriptor.DURATION:      labelpb.LabelDescriptor_INT64,
		descriptor.STRING_MAP:    labelpb.LabelDescriptor_STRING,
	}

	_ metric.HandlerBuilder = &builder{}
	_ metric.Handler        = &handler{}
)

// NewBuilder returns a builder implementing the metric.HandlerBuilder interface.
func NewBuilder() metric.HandlerBuilder {
	return &builder{createClient: createClient}
}

func createClient(cfg *config.Params) (*monitoring.MetricClient, error) {
	return monitoring.NewMetricClient(context.Background(), helper.ToOpts(cfg)...)
}

func (b *builder) SetMetricTypes(metrics map[string]*metric.Type) {
	b.metrics = metrics
}

func (b *builder) SetAdapterConfig(cfg adapter.Config) {
	b.cfg = cfg.(*config.Params)
}
func (b *builder) Validate() *adapter.ConfigErrors {
	return nil
}

// NewMetricsAspect provides an implementation for adapter.MetricsBuilder.
func (b *builder) Build(ctx context.Context, env adapter.Env) (adapter.Handler, error) {
	cfg := b.cfg
	types := make(map[string]info, len(b.metrics))
	for name, t := range b.metrics {
		i, found := cfg.MetricInfo[name]
		if !found {
			env.Logger().Warningf("No stackdriver info found for metric %s, skipping it", name)
			continue
		}
		// TODO: do we want to make sure that the definition conforms to stackdrvier requirements? Really that needs to happen during config validation
		types[name] = info{
			ttype: metricType(name),
			vtype: t.Value,
			minfo: i,
		}
	}

	// Per the documentation on config.proto, if push_interval is zero we'll default to a 1 minute push interval
	if cfg.PushInterval == time.Duration(0) {
		cfg.PushInterval = 1 * time.Minute
	}

	ticker := time.NewTicker(cfg.PushInterval)

	var err error
	var client *monitoring.MetricClient
	if client, err = b.createClient(cfg); err != nil {
		return nil, err
	}
	buffered := &buffered{
		pushMetrics: client.CreateTimeSeries,
		closeMe:     client,
		project:     cfg.ProjectId,
		m:           sync.Mutex{},
		l:           env.Logger(),
	}
	// We hold on to the ref to the ticker so we can stop it later
	buffered.start(env, ticker)
	h := &handler{
		l:          env.Logger(),
		now:        time.Now,
		projectID:  cfg.ProjectId,
		client:     buffered,
		metricInfo: types,
		ticker:     ticker,
	}
	return h, nil
}

func (h *handler) HandleMetric(_ context.Context, vals []*metric.Instance) error {
	h.l.Infof("stackdriver.Record called with %d vals", len(vals))

	// TODO: len(vals) is constant for config lifetime, consider pooling
	data := make([]*monitoringpb.TimeSeries, 0, len(vals))
	for _, val := range vals {
		minfo, found := h.metricInfo[val.Name]
		if !found {
			// We weren't configured with stackdriver data about this metric, so we don't know how to publish it.
			if h.l.VerbosityLevel(4) {
				h.l.Warningf("Skipping metric %s due to not being configured with stackdriver info about it.", val.Name)
			}
			continue
		}

		// TODO: support timestamps in templates. When we do, we can add these back
		//start, _ := ptypes.TimestampProto(val.StartTime)
		//end, _ := ptypes.TimestampProto(val.EndTime)
		start, _ := ptypes.TimestampProto(h.now())
		end, _ := ptypes.TimestampProto(h.now())

		ts := &monitoringpb.TimeSeries{
			Metric: &metricpb.Metric{
				Type:   minfo.ttype,
				Labels: helper.ToStringMap(val.Dimensions),
			},
			MetricKind: minfo.minfo.Kind,
			ValueType:  minfo.minfo.Value,
			// Since we're sending a `CreateTimeSeries` request we can only populate a single point, see
			// the documentation on the `points` field: https://cloud.google.com/monitoring/api/ref_v3/rest/v3/TimeSeries
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: start,
					EndTime:   end,
				},
				Value: toTypedVal(val.Value, minfo)},
			},
		}

		// The logging SDK has logic built in that does this for us: if a resource is not provided it fills in the global
		// resource as a default. Since we don't have equivalent behavior for monitoring, we do it ourselves.
		if val.MonitoredResourceType != "" {
			ts.Resource = &monitoredres.MonitoredResource{
				Type:   val.MonitoredResourceType,
				Labels: helper.ToStringMap(val.MonitoredResourceDimensions),
			}
		} else {
			ts.Resource = &monitoredres.MonitoredResource{
				Type: "global",
				Labels: map[string]string{
					"project_id": h.projectID,
				},
			}
		}

		data = append(data, ts)
	}
	h.client.Record(data)
	return nil
}

func (h *handler) Close() error {
	h.ticker.Stop()
	return h.client.Close()
}

func toTypedVal(val interface{}, i info) *monitoringpb.TypedValue {
	if i.minfo.Value == metricpb.MetricDescriptor_DISTRIBUTION {
		v, err := toDist(val, i)
		if err != nil {
			return &monitoringpb.TypedValue{&monitoringpb.TypedValue_DistributionValue{}}
		}
		return v
	}

	switch labelMap[i.vtype] {
	case labelpb.LabelDescriptor_BOOL:
		return &monitoringpb.TypedValue{&monitoringpb.TypedValue_BoolValue{BoolValue: val.(bool)}}
	case labelpb.LabelDescriptor_INT64:
		if t, ok := val.(time.Time); ok {
			val = t.Nanosecond() / int(time.Microsecond)
		} else if d, ok := val.(time.Duration); ok {
			val = d.Nanoseconds() / int64(time.Microsecond)
		}
		return &monitoringpb.TypedValue{&monitoringpb.TypedValue_Int64Value{Int64Value: val.(int64)}}
	default:
		return &monitoringpb.TypedValue{&monitoringpb.TypedValue_StringValue{StringValue: fmt.Sprintf("%v", val)}}
	}
}

func metricType(name string) string {
	// TODO: figure out what, if anything, we need to do to sanitize these.
	return customMetricPrefix + name
}
