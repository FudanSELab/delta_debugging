// Copyright 2016 Istio Authors
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

package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"google.golang.org/grpc"

	mixerpb "istio.io/api/mixer/v1"
	rpc "istio.io/gogo-genproto/googleapis/google/rpc"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/pkg/attribute"
	"istio.io/istio/mixer/pkg/pool"
	"istio.io/istio/mixer/pkg/runtime"
	"istio.io/istio/mixer/pkg/status"
	"istio.io/istio/pkg/log"
)

type preprocCallback func(ctx context.Context, requestBag attribute.Bag, responseBag *attribute.MutableBag) error
type checkCallback func(ctx context.Context, requestBag attribute.Bag) (*adapter.CheckResult, error)
type reportCallback func(ctx context.Context, requestBag attribute.Bag) error
type quotaCallback func(ctx context.Context, requestBag attribute.Bag,
	qma *runtime.QuotaMethodArgs) (*adapter.QuotaResult, error)

type testState struct {
	client     mixerpb.MixerClient
	connection *grpc.ClientConn
	gs         *grpc.Server
	gp         *pool.GoroutinePool
	s          *grpcServer

	check   checkCallback
	report  reportCallback
	quota   quotaCallback
	preproc preprocCallback
}

func (ts *testState) createGRPCServer() (string, error) {
	// get the network stuff setup
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 0))
	if err != nil {
		return "", err
	}

	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(32))
	grpcOptions = append(grpcOptions, grpc.MaxMsgSize(1024*1024))

	// get everything wired up
	ts.gs = grpc.NewServer(grpcOptions...)

	ts.gp = pool.NewGoroutinePool(128, false)
	ts.gp.AddWorkers(32)

	ms := NewGRPCServer(ts, ts.gp)
	ts.s = ms.(*grpcServer)
	mixerpb.RegisterMixerServer(ts.gs, ts.s)

	go func() {
		_ = ts.gs.Serve(listener)
	}()

	return listener.Addr().String(), nil
}

func (ts *testState) deleteGRPCServer() {
	ts.gs.GracefulStop()
	_ = ts.gp.Close()
}

func (ts *testState) createAPIClient(dial string) error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	var err error
	if ts.connection, err = grpc.Dial(dial, opts...); err != nil {
		return err
	}

	ts.client = mixerpb.NewMixerClient(ts.connection)
	return nil
}

func (ts *testState) deleteAPIClient() {
	_ = ts.connection.Close()
	ts.client = nil
	ts.connection = nil
}

func prepTestState() (*testState, error) {
	ts := &testState{}
	dial, err := ts.createGRPCServer()
	if err != nil {
		return nil, err
	}

	if err = ts.createAPIClient(dial); err != nil {
		ts.deleteGRPCServer()
		return nil, err
	}

	ts.preproc = func(ctx context.Context, requestBag attribute.Bag, responseBag *attribute.MutableBag) error {
		return nil
	}

	return ts, nil
}

func (ts *testState) cleanupTestState() {
	ts.deleteAPIClient()
	ts.deleteGRPCServer()
}

func (ts *testState) Check(ctx context.Context, bag attribute.Bag) (*adapter.CheckResult, error) {
	return ts.check(ctx, bag)
}

func (ts *testState) Report(ctx context.Context, bag attribute.Bag) error {
	return ts.report(ctx, bag)
}

func (ts *testState) Quota(ctx context.Context, bag attribute.Bag,
	qma *runtime.QuotaMethodArgs) (*adapter.QuotaResult, error) {

	return ts.quota(ctx, bag, qma)
}

func (ts *testState) Preprocess(ctx context.Context, req attribute.Bag, resp *attribute.MutableBag) error {
	return ts.preproc(ctx, req, resp)
}

func TestCheck(t *testing.T) {
	ts, err := prepTestState()
	if err != nil {
		t.Fatalf("Unable to prep test state: %v", err)
	}
	defer ts.cleanupTestState()

	ts.check = func(ctx context.Context, requestBag attribute.Bag) (*adapter.CheckResult, error) {
		return &adapter.CheckResult{
			Status: status.WithPermissionDenied("Not Implemented"),
		}, nil
	}

	ts.quota = func(ctx context.Context, requestBag attribute.Bag, qma *runtime.QuotaMethodArgs) (*adapter.QuotaResult, error) {
		return &adapter.QuotaResult{
			Amount: 42,
		}, nil
	}

	attr0 := mixerpb.CompressedAttributes{
		Words: []string{"A1", "A2", "A3"},
		Int64S: map[int32]int64{
			-1: 25,
			-2: 26,
			-3: 27,
		},
	}

	request := mixerpb.CheckRequest{Attributes: attr0}
	request.Quotas = make(map[string]mixerpb.CheckRequest_QuotaParams)
	request.Quotas["RequestCount"] = mixerpb.CheckRequest_QuotaParams{Amount: 42}

	response, err := ts.client.Check(context.Background(), &request)

	// if precondition fails quota always fails.
	if err != nil {
		t.Errorf("Got %v, expected success", err)
	} else if status.IsOK(response.Precondition.Status) {
		t.Error("Got precondition success, expected error")
	} else if !strings.Contains(response.Precondition.Status.Message, "Not Implemented") {
		t.Errorf("'%s' doesn't contain 'Not Implemented'", response.Precondition.Status.Message)
	} else if response.Quotas["RequestCount"].GrantedAmount != 0 {
		t.Errorf("Got %v granted amount, expecting 0", response.Quotas["RequestCount"].GrantedAmount)
	}

	ts.quota = func(ctx context.Context, requestBag attribute.Bag, qma *runtime.QuotaMethodArgs) (*adapter.QuotaResult, error) {
		return &adapter.QuotaResult{
			Status: status.WithPermissionDenied("Not Implemented"),
		}, nil
	}

	_, err = ts.client.Check(context.Background(), &request)

	if err != nil {
		t.Errorf("Got %v, expected success", err)
	}

	ts.preproc = func(ctx context.Context, requestBag attribute.Bag, responseBag *attribute.MutableBag) error {
		responseBag.Set("A1", "override")
		responseBag.Set("genAttrGen", "genAttrGenValue")
		return nil
	}

	ts.check = func(ctx context.Context, requestBag attribute.Bag) (*adapter.CheckResult, error) {
		if val, _ := requestBag.Get("A1"); val == "override" {
			return nil, errors.New("attribute overriding not allowed in Check")
		}
		if val, _ := requestBag.Get("genAttrGen"); val != "genAttrGenValue" {
			return nil, errors.New("generated attribute via preproc not part of check attributes")
		}
		return &adapter.CheckResult{}, nil
	}

	chkRes, err := ts.client.Check(context.Background(), &request)
	if err != nil || chkRes.Precondition.Status.Code != 0 {
		t.Errorf("Got error; expect success: %v, %v", chkRes, err)
	}
}

func TestCheckQuota(t *testing.T) {
	ts, err := prepTestState()
	if err != nil {
		t.Fatalf("Unable to prep test state: %v", err)
	}
	defer ts.cleanupTestState()

	ts.check = func(ctx context.Context, requestBag attribute.Bag) (*adapter.CheckResult, error) {
		return &adapter.CheckResult{
			Status: status.OK,
		}, nil
	}

	ts.quota = func(ctx context.Context, requestBag attribute.Bag, qma *runtime.QuotaMethodArgs) (*adapter.QuotaResult, error) {
		return &adapter.QuotaResult{
			Amount: 42,
		}, nil
	}

	attr0 := mixerpb.CompressedAttributes{
		Words: []string{"A1", "A2", "A3"},
		Int64S: map[int32]int64{
			-1: 25,
			-2: 26,
			-3: 27,
		},
	}

	request := mixerpb.CheckRequest{Attributes: attr0}
	request.Quotas = make(map[string]mixerpb.CheckRequest_QuotaParams)
	request.Quotas["RequestCount"] = mixerpb.CheckRequest_QuotaParams{Amount: 42}

	response, err := ts.client.Check(context.Background(), &request)

	if err != nil {
		t.Errorf("Got %v, expected success", err)
	} else if !status.IsOK(response.Precondition.Status) {
		t.Errorf("Got unexpected failure %+v", response.Precondition.Status)
	} else if response.Quotas["RequestCount"].GrantedAmount != 42 {
		t.Errorf("Got %v granted amount, expecting 42", response.Quotas["RequestCount"].GrantedAmount)
	}

	ts.quota = func(ctx context.Context, requestBag attribute.Bag, qma *runtime.QuotaMethodArgs) (*adapter.QuotaResult, error) {
		return &adapter.QuotaResult{
			Status: status.WithPermissionDenied("Not Implemented"),
		}, nil
	}

	_, err = ts.client.Check(context.Background(), &request)

	if err != nil {
		t.Errorf("Got %v, expected success", err)
	}
}

func TestReport(t *testing.T) {
	ts, err := prepTestState()
	if err != nil {
		t.Fatalf("Unable to prep test state: %v", err)
	}
	defer ts.cleanupTestState()

	ts.report = func(ctx context.Context, requestBag attribute.Bag) error {
		return nil
	}

	request := mixerpb.ReportRequest{Attributes: []mixerpb.CompressedAttributes{{}}}
	_, err = ts.client.Report(context.Background(), &request)
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	ts.report = func(ctx context.Context, requestBag attribute.Bag) error {
		return errors.New("not Implemented")
	}

	_, err = ts.client.Report(context.Background(), &request)
	if err == nil {
		t.Errorf("Got success, expected failure")
	}

	// test out delta encoding of attributes
	attr0 := mixerpb.CompressedAttributes{
		Words: []string{"A1", "A2", "A3"},
		Int64S: map[int32]int64{
			-1: 25,
			-2: 26,
			-3: 27,
		},
	}

	attr1 := mixerpb.CompressedAttributes{
		Words: []string{"A1", "A2", "A3"},
		Int64S: map[int32]int64{
			-2: 42,
		},
	}

	callCount := 0
	ts.report = func(ctx context.Context, requestBag attribute.Bag) error {
		v1, _ := requestBag.Get("A1")
		v2, _ := requestBag.Get("A2")
		v3, _ := requestBag.Get("A3")

		i1 := v1.(int64)
		i2 := v2.(int64)
		i3 := v3.(int64)

		if callCount == 0 {
			if i1 != 25 || i2 != 26 || i3 != 27 {
				t.Errorf("Got %d %d %d, expected 25 26 27", i1, i2, i3)
			}
		} else if callCount == 1 {
			if i1 != 25 || i2 != 42 || i3 != 27 {
				t.Errorf("Got %d %d %d, expected 25 42 27", i1, i2, i3)
			}

		} else {
			t.Errorf("Dispatched to Report method more than twice")
		}
		callCount++
		return nil
	}

	request = mixerpb.ReportRequest{Attributes: []mixerpb.CompressedAttributes{attr0, attr1}}
	_, _ = ts.client.Report(context.Background(), &request)

	if callCount == 0 {
		t.Errorf("Got %d, expected call count of 2", callCount)
	}

	ts.preproc = func(ctx context.Context, requestBag attribute.Bag, responseBag *attribute.MutableBag) error {
		responseBag.Set("A1", "override")
		responseBag.Set("genAttrGen", "genAttrGenValue")
		return nil
	}

	ts.report = func(ctx context.Context, requestBag attribute.Bag) error {
		if val, _ := requestBag.Get("A1"); val == "override" {
			return errors.New("attribute overriding NOT allowed in Report")
		}
		if val, _ := requestBag.Get("genAttrGen"); val != "genAttrGenValue" {
			return errors.New("generated attribute via preproc not part of report attributes")
		}
		return nil
	}

	if _, err = ts.client.Report(context.Background(), &request); err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
}

func TestUnknownStatus(t *testing.T) {
	ts, err := prepTestState()
	if err != nil {
		t.Fatalf("Unable to prep test state: %v", err)
	}
	defer ts.cleanupTestState()

	ts.check = func(ctx context.Context, requestBag attribute.Bag) (*adapter.CheckResult, error) {
		return &adapter.CheckResult{
			Status: rpc.Status{
				Code:    12345678,
				Message: "DEADBEEF!",
			},
		}, nil
	}
	request := mixerpb.CheckRequest{}
	resp, err := ts.client.Check(context.Background(), &request)
	if err != nil {
		t.Error("Got failure, expected success")
	} else if !strings.Contains(resp.Precondition.Status.Message, "DEADBEEF!") {
		t.Errorf("Got '%s', expected DEADBEEF!", resp.Precondition.Status.Message)
	}
}

func TestFailingPreproc(t *testing.T) {
	ts, err := prepTestState()
	if err != nil {
		t.Fatalf("Unable to prep test state: %v", err)
	}
	defer ts.cleanupTestState()

	ts.preproc = func(ctx context.Context, requestBag attribute.Bag, responseBag *attribute.MutableBag) error {
		return errors.New("123 preproc failed")
	}

	{
		request := mixerpb.CheckRequest{}
		resp, err := ts.client.Check(context.Background(), &request)
		if resp != nil {
			t.Error("Expecting no response, got one")
		}
		if err == nil {
			t.Error("Got success, expected failure")
		} else if !strings.Contains(err.Error(), "123 preproc failed") {
			t.Errorf("Got '%s', expected '123 preproc failed'", err.Error())
		}
	}

	{
		request := mixerpb.ReportRequest{Attributes: []mixerpb.CompressedAttributes{{}}}
		resp, err := ts.client.Report(context.Background(), &request)
		if resp != nil {
			t.Error("Expecting no response, got one")
		}
		if err == nil {
			t.Error("Got success, expected failure")
		} else if !strings.Contains(err.Error(), "123 preproc failed") {
			t.Errorf("Got '%s', expected '123 preproc failed'", err.Error())
		}
	}
}

func init() {
	// bump up the log level so log-only logic runs during the tests, for correctness and coverage.
	o := log.NewOptions()
	_ = o.SetOutputLevel(log.DebugLevel)
	_ = log.Configure(o)
}
