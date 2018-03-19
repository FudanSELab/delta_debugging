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

package metric

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/genproto/googleapis/api/distribution"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	descriptor "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/adapter/stackdriver/config"
	"istio.io/istio/mixer/pkg/adapter/test"
	metrict "istio.io/istio/mixer/template/metric"
)

type fakebuf struct {
	buf []*monitoringpb.TimeSeries
}

func (f *fakebuf) Record(in []*monitoringpb.TimeSeries) {
	f.buf = append(f.buf, in...)
}

func (*fakebuf) Close() error { return nil }

var clientFunc = func(err error) createClientFunc {
	return func(cfg *config.Params) (*monitoring.MetricClient, error) {
		return nil, err
	}
}

func TestFactory_NewMetricsAspect(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.Params
		metricNames    []string
		missingMetrics []string // We check that the method logged these metric names because they're not mapped in cfg
		err            string   // If != "" we expect an error containing this string
	}{
		{"empty", &config.Params{}, []string{}, []string{}, ""},
		{"missing metric", &config.Params{}, []string{"request_count"}, []string{"request_count"}, ""},
		{
			"happy path",
			&config.Params{MetricInfo: map[string]*config.Params_MetricInfo{"request_count": {}}},
			[]string{"request_count"},
			[]string{},
			"",
		},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("[%d] %s", idx, tt.name), func(t *testing.T) {
			metrics := make(map[string]*metrict.Type)
			for _, name := range tt.metricNames {
				metrics[name] = &metrict.Type{}
			}
			env := test.NewEnv(t)
			b := &builder{createClient: clientFunc(nil)}
			b.SetMetricTypes(metrics)
			b.SetAdapterConfig(tt.cfg)
			_, err := b.Build(context.Background(), env)
			if err != nil || tt.err != "" {
				if tt.err == "" {
					t.Fatalf("factory{}.NewMetricsAspect(test.NewEnv(t), nil, nil) = '%s', wanted no err", err.Error())
				} else if !strings.Contains(err.Error(), tt.err) {
					t.Fatalf("Expected errors containing the string '%s', actual: '%s'", tt.err, err.Error())
				}
			}
			// If we expect missing metrics make sure they're present in the logs; otherwise make sure none were missing.
			if len(tt.missingMetrics) > 0 {
				for _, missing := range tt.missingMetrics {
					found := false
					for _, log := range env.GetLogs() {
						found = found || strings.Contains(log, missing)
					}
					if !found {
						t.Errorf("Wanted missing log %s, got logs: %v", missing, env.GetLogs())
					}
				}
			} else {
				for _, log := range env.GetLogs() {
					if strings.Contains(log, "No stackdriver info found for metric") {
						t.Errorf("Expected no missing metrics, found log entry: %s", log)
					}
				}
			}
		})
	}
}

func TestFactory_NewMetricsAspect_Errs(t *testing.T) {
	err := fmt.Errorf("expected")
	b := &builder{createClient: clientFunc(err)}
	b.SetAdapterConfig(&config.Params{})
	res, e := b.Build(context.Background(), test.NewEnv(t))
	if e != nil && !strings.Contains(e.Error(), err.Error()) {
		t.Fatalf("Expected error from factory.createClient to be propagated, got %v, %v", res, e)
	} else if e == nil {
		t.Fatalf("Got no error")
	}
}

func TestRecord(t *testing.T) {
	projectID := "pid"
	resource := &monitoredres.MonitoredResource{
		Type: "global",
		Labels: map[string]string{
			"project_id": projectID,
		},
	}
	m := &metricpb.Metric{
		Type:   "type",
		Labels: map[string]string{"str": "str", "int": "34"},
	}
	info := map[string]info{
		"gauge": {
			ttype: "type",
			minfo: &config.Params_MetricInfo{Kind: metricpb.MetricDescriptor_GAUGE, Value: metricpb.MetricDescriptor_INT64},
			vtype: descriptor.INT64,
		},
		"cumulative": {
			ttype: "type",
			minfo: &config.Params_MetricInfo{Kind: metricpb.MetricDescriptor_CUMULATIVE, Value: metricpb.MetricDescriptor_STRING},
			vtype: descriptor.STRING,
		},
		"delta": {
			ttype: "type",
			minfo: &config.Params_MetricInfo{Kind: metricpb.MetricDescriptor_DELTA, Value: metricpb.MetricDescriptor_BOOL},
			vtype: descriptor.BOOL,
		},
		"distribution-linear": {
			ttype: "type",
			minfo: &config.Params_MetricInfo{
				Kind:  metricpb.MetricDescriptor_CUMULATIVE,
				Value: metricpb.MetricDescriptor_DISTRIBUTION,
				Buckets: &config.Params_MetricInfo_BucketsDefinition{Definition: &config.Params_MetricInfo_BucketsDefinition_LinearBuckets{
					// under, 1-6, 6-11, over
					LinearBuckets: &config.Params_MetricInfo_BucketsDefinition_Linear{
						NumFiniteBuckets: 2,
						Offset:           1,
						Width:            5,
					}}},
			},
			vtype: descriptor.DOUBLE,
		},
		"distribution-exp": {
			ttype: "type",
			minfo: &config.Params_MetricInfo{
				Kind:  metricpb.MetricDescriptor_CUMULATIVE,
				Value: metricpb.MetricDescriptor_DISTRIBUTION,
				Buckets: &config.Params_MetricInfo_BucketsDefinition{Definition: &config.Params_MetricInfo_BucketsDefinition_ExponentialBuckets{
					// under, 10-100, 100-1000, over
					ExponentialBuckets: &config.Params_MetricInfo_BucketsDefinition_Exponential{
						NumFiniteBuckets: 2,
						Scale:            10,
						GrowthFactor:     10,
					}}},
			},
			vtype: descriptor.DOUBLE,
		},
		"distribution-explicit": {
			ttype: "type",
			minfo: &config.Params_MetricInfo{
				Kind:  metricpb.MetricDescriptor_CUMULATIVE,
				Value: metricpb.MetricDescriptor_DISTRIBUTION,
				Buckets: &config.Params_MetricInfo_BucketsDefinition{Definition: &config.Params_MetricInfo_BucketsDefinition_ExplicitBuckets{
					// under, 1-10, 10-100, over
					ExplicitBuckets: &config.Params_MetricInfo_BucketsDefinition_Explicit{
						Bounds: []float64{1, 10, 100},
					}}},
			},
			vtype: descriptor.DOUBLE,
		},
	}
	now := time.Now()
	pbnow, _ := ptypes.TimestampProto(now)

	tests := []struct {
		name     string
		vals     []*metrict.Instance
		expected []*monitoringpb.TimeSeries
	}{
		{"empty", []*metrict.Instance{}, []*monitoringpb.TimeSeries{}},
		{"missing", []*metrict.Instance{{Name: "not in the info map"}}, []*monitoringpb.TimeSeries{}},
		{"gauge", []*metrict.Instance{
			{
				Name:       "gauge",
				Value:      int64(7),
				Dimensions: map[string]interface{}{"str": "str", "int": int64(34)},
			},
		}, []*monitoringpb.TimeSeries{
			{
				Metric:     m,
				Resource:   resource,
				MetricKind: metricpb.MetricDescriptor_GAUGE,
				ValueType:  metricpb.MetricDescriptor_INT64,
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{StartTime: pbnow, EndTime: pbnow},
					Value:    &monitoringpb.TypedValue{&monitoringpb.TypedValue_Int64Value{Int64Value: int64(7)}},
				}},
			},
		}},
		{"cumulative", []*metrict.Instance{
			{
				Name:       "cumulative",
				Value:      "asldkfj",
				Dimensions: map[string]interface{}{"str": "str", "int": int64(34)},
			},
		}, []*monitoringpb.TimeSeries{
			{
				Metric:     m,
				Resource:   resource,
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_STRING,
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{StartTime: pbnow, EndTime: pbnow},
					Value:    &monitoringpb.TypedValue{&monitoringpb.TypedValue_StringValue{StringValue: "asldkfj"}},
				}},
			},
		}},
		{"delta", []*metrict.Instance{
			{
				Name:       "delta",
				Value:      true,
				Dimensions: map[string]interface{}{"str": "str", "int": int64(34)},
			},
		}, []*monitoringpb.TimeSeries{
			{
				Metric:     m,
				Resource:   resource,
				MetricKind: metricpb.MetricDescriptor_DELTA,
				ValueType:  metricpb.MetricDescriptor_BOOL,
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{StartTime: pbnow, EndTime: pbnow},
					Value:    &monitoringpb.TypedValue{&monitoringpb.TypedValue_BoolValue{BoolValue: true}},
				}},
			},
		}},
		{"distribution-linear", []*metrict.Instance{
			{
				Name:       "distribution-linear",
				Value:      float64(6),
				Dimensions: map[string]interface{}{"str": "str", "int": int64(34)},
			},
		}, []*monitoringpb.TimeSeries{
			{
				Metric:     m,
				Resource:   resource,
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{StartTime: pbnow, EndTime: pbnow},
					Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DistributionValue{
						DistributionValue: &distribution.Distribution{
							Count:         1,
							BucketOptions: linear(1, 5, 2),
							BucketCounts:  []int64{0, 0, 1, 0},
						}}},
				}},
			},
		}},
		{"distribution-exp", []*metrict.Instance{
			{
				Name:       "distribution-exp",
				Value:      float64(99),
				Dimensions: map[string]interface{}{"str": "str", "int": int64(34)},
			},
		}, []*monitoringpb.TimeSeries{
			{
				Metric:     m,
				Resource:   resource,
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{StartTime: pbnow, EndTime: pbnow},
					Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DistributionValue{
						DistributionValue: &distribution.Distribution{
							Count:         1,
							BucketOptions: exp(10, 10, 2),
							BucketCounts:  []int64{0, 1, 0, 0},
						}}},
				}},
			},
		}},
		{"distribution-explicit", []*metrict.Instance{
			{
				Name:       "distribution-explicit",
				Value:      float64(9),
				Dimensions: map[string]interface{}{"str": "str", "int": int64(34)},
			},
		}, []*monitoringpb.TimeSeries{
			{
				Metric:     m,
				Resource:   resource,
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
				Points: []*monitoringpb.Point{{
					Interval: &monitoringpb.TimeInterval{StartTime: pbnow, EndTime: pbnow},
					Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DistributionValue{
						DistributionValue: &distribution.Distribution{
							Count:         1,
							BucketOptions: explicit([]float64{1, 10, 100}),
							BucketCounts:  []int64{0, 1, 0, 0},
						}}},
				}},
			},
		}},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("[%d] %s", idx, tt.name), func(t *testing.T) {
			buf := &fakebuf{}
			s := &handler{metricInfo: info, projectID: projectID, client: buf, l: test.NewEnv(t).Logger(), now: func() time.Time { return now }}
			_ = s.HandleMetric(context.Background(), tt.vals)

			if len(buf.buf) != len(tt.expected) {
				t.Errorf("Want %d values to send, got %d", len(tt.expected), len(buf.buf))
			}
			for _, expected := range tt.expected {
				found := false
				for _, actual := range buf.buf {
					found = found || reflect.DeepEqual(expected, actual)
				}
				if !found {
					t.Errorf("Want timeseries %v, but not present: %v", expected, buf.buf)
				}
			}
		})
	}
}
