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

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"google.golang.org/grpc"

	"sync"

	istio_mixer_v1 "istio.io/api/mixer/v1"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/pkg/attribute"
	"istio.io/istio/mixer/pkg/config/storetest"
	"istio.io/istio/mixer/pkg/server"
	"istio.io/istio/mixer/pkg/template"
	template2 "istio.io/istio/mixer/template"
)

// Utility to help write Mixer-adapter integration tests.

type (
	// Scenario fully defines an adapter integration test
	Scenario struct {
		// Configs is a list of CRDs that Mixer will read.
		Configs []string

		// ParallelCalls is a list of test calls to be made to Mixer
		// in parallel.
		ParallelCalls []Call

		// Setup is a callback function that will be called at the beginning of the test. It is
		// meant to be used for things like starting a local backend server. Setup function returns a
		// context (interface{}) which is passed back into the Teardown and the GetState functions.
		// pass nil if no setup needed
		Setup SetupFn

		// Teardown is a callback function that will be called at the end of the test. It is
		// meant to be used for things like stopping a local backend server that might have been started during Setup.
		// pass nil if no teardown is needed
		Teardown TeardownFn

		// GetState lets the test provide any (interface{}) adapter specific data to be part of baseline.
		// Example: for prometheus adapter, the actual metric reported to the local backend can be embedded into the
		// expected json baseline.
		// pass nil if no adapter specific state is part of baseline.
		GetState GetStateFn

		// Templates supported by Mixer.
		// If `tmpls` is not specified, the default templates inside istio.io/istio/mixer/template.SupportedTmplInfo
		// are made available to the Mixer.
		tmpls map[string]template.Info

		// Want is the expected serialized json for the Result struct.
		// Result.AdapterState is what the callback function `getState`, passed to `RunTest`, returns.
		//
		// New test can start of with an empty "{}" string and then
		// get the baseline from the failure logs upon execution.
		Want string
	}
	// Call represents the input to make a call to Mixer
	Call struct {
		// CallKind can be either CHECK or REPORT
		CallKind CallKind
		// Attrs to call the Mixer with.
		Attrs map[string]interface{}
		// Quotas info to call the Mixer with.
		Quotas map[string]istio_mixer_v1.CheckRequest_QuotaParams
	}
	// CallKind represents the call to make; check or report.
	CallKind int32

	// Result represents the test baseline
	Result struct {
		// AdapterState represents adapter specific baseline data. AdapterState is what the callback function
		// `getState`, passed to `RunTest`, returns.
		AdapterState interface{} `json:"AdapterState"`
		// Returns represents the return data from calls to Mixer
		Returns []Return `json:"Returns"`
	}
	// Return represents the return data from a call to Mixer
	Return struct {
		// Check is the response from a check call to Mixer
		Check adapter.CheckResult `json:"Check"`
		// Quota is the response from a check call to Mixer
		Quota map[string]adapter.QuotaResult `json:"Quota"`
		// Error is the error from call to Mixer
		Error error `json:"Error"`
	}
)

const (
	// CHECK for  Mixer Check
	CHECK CallKind = iota
	// REPORT for  Mixer Report
	REPORT
)

type (
	// SetupFn functions will be called at the beginning of the test
	SetupFn func() (ctx interface{}, err error)
	// TeardownFn functions will be called at the end of the test
	TeardownFn func(ctx interface{})
	// GetStateFn returns the adapter specific state upon test execution. The return value becomes part of
	// expected Result.AdapterState.
	GetStateFn func(ctx interface{}) (interface{}, error)
)

// RunTest performs a Mixer adapter integration test using in-memory Mixer and config store.
// NOTE: DO NOT invoke this using `t.Run(string, func)` because that would execute func in a separate go routine.
// Separate go routines would cause the test to fail randomly because fixed ports cannot be assigned and cleaned up
// deterministically on each iteration.
//
// * adapterInfo provides the InfoFn for the adapter under test.
// * Scenario provide the adapter/handler/rule configs along with the call parameters (check or report, and attributes)
//   Optionally, it also takes the test specific SetupFn, TeardownFn, GetStateFn and list of supported templates.
func RunTest(
	t *testing.T,
	adapterInfo adapter.InfoFn,
	scenario Scenario,
) {

	// Let the test do some initial setup.
	var ctx interface{}
	var err error
	if scenario.Setup != nil {
		ctx, err = scenario.Setup()
		// Teardown the initial setup
		if scenario.Teardown != nil {
			defer scenario.Teardown(ctx)
		}
		if err != nil {
			t.Fatalf("initial setup failed: %v", err)
		}
	}

	if len(scenario.tmpls) == 0 {
		scenario.tmpls = template2.SupportedTmplInfo
	}

	// Start Mixer
	var args *server.Args
	var env *server.Server
	if args, err = getServerArgs(scenario.tmpls, []adapter.InfoFn{adapterInfo}, scenario.Configs); err != nil {
		t.Fatalf("fail to create mixer args: %v", err)
	}

	// Setting zero will make Mixer pick any available port.
	args.APIPort = 0
	args.MonitoringPort = 0

	if env, err = server.New(args); err != nil {
		t.Fatalf("fail to new mixer: %v", err)
	}
	env.Run()
	defer closeHelper(env)

	// Connect the client to Mixer
	conn, err := grpc.Dial(env.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Unable to connect to gRPC server: %v", err)
	}
	client := istio_mixer_v1.NewMixerClient(conn)
	defer closeHelper(conn)

	// Invoke calls async
	var wg sync.WaitGroup
	wg.Add(len(scenario.ParallelCalls))

	got := Result{Returns: make([]Return, len(scenario.ParallelCalls))}
	for i, call := range scenario.ParallelCalls {
		go execute(call, args.ConfigIdentityAttribute, args.ConfigIdentityAttributeDomain, client, got.Returns, i, &wg)
	}
	// wait for calls to finish
	wg.Wait()

	// get adapter state. NOTE: We are doing marshal and then unmarshal it back into generic interface{}.
	// This is done to make getState output into generic json map or array; which is exactly what we get when un-marshalling
	// the baseline json. Without this, deep equality on un-marshalled baseline AdapterState would defer
	// from the rich object returned by getState function.
	if scenario.GetState != nil {
		adptState, _ := scenario.GetState(ctx)
		var adptStateBytes []byte
		if adptStateBytes, err = json.Marshal(adptState); err != nil {
			t.Fatalf("Unable to convert %v into json: %v", adptState, err)
		}
		if err = json.Unmarshal(adptStateBytes, &got.AdapterState); err != nil {
			t.Fatalf("Unable to unmarshal %s into interface{}: %v", string(adptStateBytes), err)
		}
	}

	var want Result
	if err = json.Unmarshal([]byte(scenario.Want), &want); err != nil {
		t.Fatalf("Unable to unmarshal %s into Result: %v", scenario.Want, err)
	}

	// compare
	if !reflect.DeepEqual(want, got) {
		gotJSON, err := json.MarshalIndent(got, "", " ")
		if err != nil {
			t.Fatalf("Unable to convert %v into json: %v", got, err)
		}
		wantJSON, err := json.MarshalIndent(want, "", " ")
		if err != nil {
			t.Fatalf("Unable to convert %v into json: %v", want, err)
		}
		t.Errorf("\ngot=>\n%s\nwant=>\n%s", gotJSON, wantJSON)
	}
}

func execute(c Call, idAttr string, idAttrDomain string, client istio_mixer_v1.MixerClient, returns []Return, i int, wg *sync.WaitGroup) {
	ret := Return{}
	switch c.CallKind {
	case CHECK:
		req := istio_mixer_v1.CheckRequest{
			Attributes: getAttrBag(c.Attrs,
				idAttr,
				idAttrDomain),
			Quotas: c.Quotas,
		}

		result, resultErr := client.Check(context.Background(), &req)
		result.Precondition.ReferencedAttributes = istio_mixer_v1.ReferencedAttributes{}
		ret.Error = resultErr
		if len(c.Quotas) > 0 {
			ret.Quota = make(map[string]adapter.QuotaResult)
			for k := range c.Quotas {
				ret.Quota[k] = adapter.QuotaResult{
					Amount: result.Quotas[k].GrantedAmount, ValidDuration: result.Quotas[k].ValidDuration,
				}
			}
		} else {
			ret.Check.ValidDuration = result.Precondition.ValidDuration
			ret.Check.ValidUseCount = result.Precondition.ValidUseCount
			ret.Check.Status = result.Precondition.Status
		}

	case REPORT:
		req := istio_mixer_v1.ReportRequest{
			Attributes: []istio_mixer_v1.CompressedAttributes{
				getAttrBag(c.Attrs,
					idAttr,
					idAttrDomain)},
		}
		_, responseErr := client.Report(context.Background(), &req)
		ret.Error = responseErr
	}
	returns[i] = ret
	wg.Done()
}

func closeHelper(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func getServerArgs(
	tmpls map[string]template.Info,
	adpts []adapter.InfoFn,
	cfgs []string) (*server.Args, error) {

	args := server.DefaultArgs()
	args.Templates = tmpls
	args.Adapters = adpts

	data := make([]string, 0)
	data = append(data, cfgs...)

	// always include the attribute vocabulary
	_, filename, _, _ := runtime.Caller(0)
	if f, err := filepath.Abs(path.Join(path.Dir(filename), "../../../testdata/config/attributes.yaml")); err != nil {
		return nil, fmt.Errorf("cannot load attributes.yaml: %v", err)
	} else if f, err := ioutil.ReadFile(f); err != nil {
		return nil, fmt.Errorf("cannot load attributes.yaml: %v", err)
	} else {
		data = append(data, string(f))
	}

	var err error
	args.ConfigStore, err = storetest.SetupStoreForTest(data...)
	return args, err
}

func getAttrBag(attrs map[string]interface{}, identityAttr, identityAttrDomain string) istio_mixer_v1.CompressedAttributes {
	requestBag := attribute.GetMutableBag(nil)
	requestBag.Set(identityAttr, identityAttrDomain)
	for k, v := range attrs {
		switch v.(type) {
		case map[string]interface{}:
			mapCast := make(map[string]string, len(v.(map[string]interface{})))

			for k1, v1 := range v.(map[string]interface{}) {
				mapCast[k1] = v1.(string)
			}
			requestBag.Set(k, mapCast)
		default:
			requestBag.Set(k, v)
		}
	}

	var attrProto istio_mixer_v1.CompressedAttributes
	requestBag.ToProto(&attrProto, nil, 0)
	return attrProto
}
