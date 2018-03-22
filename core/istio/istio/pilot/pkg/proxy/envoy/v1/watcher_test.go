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

package v1

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"reflect"
	"sync"
	"testing"
	"time"

	_ "github.com/golang/glog"
	"github.com/howeyc/fsnotify"
	"github.com/stretchr/testify/assert"

	"istio.io/istio/pilot/pkg/model"
)

type TestAgent struct {
	schedule func(interface{})
}

func (ta TestAgent) ScheduleConfigUpdate(config interface{}) {
	ta.schedule(config)
}

func (ta TestAgent) Run(ctx context.Context) {
	<-ctx.Done()
}

func TestRunReload(t *testing.T) {
	called := make(chan bool)
	agent := TestAgent{
		schedule: func(_ interface{}) {
			called <- true
		},
	}
	config := model.DefaultProxyConfig()
	node := model.Proxy{
		Type: model.Ingress,
		ID:   "random",
	}
	watcher := NewWatcher(config, agent, node, []CertSource{{Directory: "random"}}, nil)
	ctx, cancel := context.WithCancel(context.Background())

	// watcher starts agent and schedules a config update
	go watcher.Run(ctx)

	select {
	case <-called:
		// expected
		cancel()
	case <-time.After(time.Second):
		t.Errorf("The callback is not called within time limit " + time.Now().String())
		cancel()
	}
}

type pilotStubHandler struct {
	sync.Mutex
	States []pilotStubState
}

type pilotStubState struct {
	StatusCode int
	Response   string
}

func (p *pilotStubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Lock()
	w.WriteHeader(p.States[0].StatusCode)
	_, _ = w.Write([]byte(p.States[0].Response))
	p.States = p.States[1:]
	p.Unlock()
}

func Test_watcher_retrieveAZ(t *testing.T) {
	tests := []struct {
		name        string
		az          string
		retries     int
		nodeType    model.NodeType
		wantReload  bool
		wantAZ      string
		pilotStates []pilotStubState
	}{
		{
			name:       "retrieves an AZ and calls for a reload",
			wantReload: true,
			wantAZ:     "az1",
			nodeType:   model.Ingress,
			retries:    5,
			pilotStates: []pilotStubState{
				{StatusCode: 200, Response: "az1"},
			},
		},
		{
			name:       "retries if it receives an error",
			wantReload: true,
			wantAZ:     "az1",
			nodeType:   model.Ingress,
			retries:    5,
			pilotStates: []pilotStubState{
				{StatusCode: 301, Response: ""},
				{StatusCode: 200, Response: "az1"},
			},
		},
		{
			name:       "retries if it receives non 200 status from pilot",
			wantReload: true,
			wantAZ:     "az1",
			nodeType:   model.Ingress,
			retries:    5,
			pilotStates: []pilotStubState{
				{StatusCode: 500, Response: ""},
				{StatusCode: 200, Response: "az1"},
			},
		},
		{
			name:       "do nothing if az is set",
			az:         "az1",
			wantAZ:     "az1",
			nodeType:   model.Ingress,
			retries:    5,
			wantReload: false,
		},
		{
			name:       "do nothing if node type is pilot",
			nodeType:   "pilot",
			wantReload: false,
		},
		{
			name:       "do nothing if node type is mixer",
			nodeType:   "mixer",
			wantReload: false,
		},
		{
			name:     "give up after retry count is reached",
			nodeType: model.Ingress,
			retries:  2,
			pilotStates: []pilotStubState{
				{StatusCode: 500, Response: ""},
				{StatusCode: 500, Response: ""},
				{StatusCode: 500, Response: ""},
				{StatusCode: 200, Response: "az1"},
			},
			wantReload: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := make(chan bool)
			agent := TestAgent{
				schedule: func(_ interface{}) {
					called <- true
				},
			}
			node := model.Proxy{
				Type:      tt.nodeType,
				ID:        "id",
				Domain:    "domain",
				IPAddress: "ip",
			}
			config := model.DefaultProxyConfig()
			config.AvailabilityZone = tt.az
			pilotStub := httptest.NewServer(
				&pilotStubHandler{States: tt.pilotStates},
			)
			stubURL, _ := url.Parse(pilotStub.URL)
			config.DiscoveryAddress = stubURL.Host
			w := NewWatcher(config, agent, node, nil, nil)
			ctx, cancel := context.WithCancel(context.Background())

			go w.(*watcher).retrieveAZ(ctx, 0, tt.retries)

			select {
			case <-called:
				if !tt.wantReload {
					t.Errorf("Unexpected reload called")
				}
				assert.Equal(t, tt.wantAZ, w.(*watcher).config.AvailabilityZone)
				cancel()
			case <-time.After(time.Second):
				if tt.wantReload {
					t.Errorf("The callback is not called within time limit " + time.Now().String())
				}
				cancel()
			}

		})
	}
}

func TestWatchCerts_Multiple(t *testing.T) {

	lock := sync.Mutex{}
	called := 0

	callback := func() {
		lock.Lock()
		defer lock.Unlock()
		called++
	}

	maxDelay := 500 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	wch := make(chan *fsnotify.FileEvent, 10)

	go watchFileEvents(ctx, wch, maxDelay, callback)

	// fire off multiple events
	wch <- &fsnotify.FileEvent{Name: "f1"}
	wch <- &fsnotify.FileEvent{Name: "f2"}
	wch <- &fsnotify.FileEvent{Name: "f3"}

	// sleep for less than maxDelay
	time.Sleep(maxDelay / 2)

	// Expect no events to be delivered within maxDelay.
	lock.Lock()
	if called != 0 {
		t.Fatalf("Called %d times, want 0", called)
	}
	lock.Unlock()

	// wait for quiet period
	time.Sleep(maxDelay)

	// Expect exactly 1 event to be delivered.
	lock.Lock()
	defer lock.Unlock()
	if called != 1 {
		t.Fatalf("Called %d times, want 1", called)
	}

	cancel()
}

func TestWatchCerts(t *testing.T) {
	name, err := ioutil.TempDir("testdata", "certs")
	if err != nil {
		t.Errorf("failed to create a temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(name); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	called := make(chan bool)
	callbackFunc := func() {
		called <- true
	}

	ctx, cancel := context.WithCancel(context.Background())

	go watchCerts(ctx, []string{name}, watchFileEvents, 50*time.Millisecond, callbackFunc)

	// sleep one second to make sure the watcher is set up before change is made
	time.Sleep(time.Second)

	// make a change to the watched dir
	if _, err := ioutil.TempFile(name, "test.file"); err != nil {
		t.Errorf("failed to create a temp file in testdata/certs: %v", err)
	}

	select {
	case <-called:
		// expected
		cancel()
	case <-time.After(time.Second):
		t.Errorf("The callback is not called within time limit " + time.Now().String())
		cancel()
	}

	// should terminate immediately
	go watchCerts(ctx, nil, watchFileEvents, 50*time.Millisecond, callbackFunc)
}

func TestGenerateCertHash(t *testing.T) {
	name, err := ioutil.TempDir("testdata", "certs")
	if err != nil {
		t.Errorf("failed to create a temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(name); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	h := sha256.New()
	authFiles := []string{model.CertChainFilename, model.KeyFilename, model.RootCertFilename}
	for _, file := range authFiles {
		content := []byte(file)
		if err := ioutil.WriteFile(path.Join(name, file), content, 0644); err != nil {
			t.Errorf("failed to write file %s (error %v)", file, err)
		}
		if _, err := h.Write(content); err != nil {
			t.Errorf("failed to write hash (error %v)", err)
		}
	}
	expectedHash := h.Sum(nil)

	h2 := sha256.New()
	generateCertHash(h2, name, append(authFiles, "missing-file"))
	actualHash := h2.Sum(nil)
	if !bytes.Equal(actualHash, expectedHash) {
		t.Errorf("Actual hash value (%v) is different than the expected hash value (%v)", actualHash, expectedHash)
	}

	generateCertHash(h2, "", nil)
	emptyHash := h2.Sum(nil)
	if !bytes.Equal(emptyHash, expectedHash) {
		t.Error("hash should not be affected by empty directory")
	}
}

func TestEnvoyArgs(t *testing.T) {
	config := model.DefaultProxyConfig()
	config.ServiceCluster = "my-cluster"
	config.AvailabilityZone = "my-zone"
	config.Concurrency = 8

	test := envoy{config: config, node: "my-node", extraArgs: []string{"-l", "trace"}}
	testProxy := NewProxy(config, "my-node", "trace")
	if !reflect.DeepEqual(testProxy, test) {
		t.Errorf("unexpected struct got\n%v\nwant\n%v", testProxy, test)
	}

	got := test.args("test.json", 5)
	want := []string{
		"-c", "test.json",
		"--restart-epoch", "5",
		"--drain-time-s", "2",
		"--parent-shutdown-time-s", "3",
		"--service-cluster", "my-cluster",
		"--service-node", "my-node",
		"--max-obj-name-len", fmt.Sprint(MaxClusterNameLength), // TODO: use MeshConfig.StatNameLength instead
		"-l", "trace",
		"--concurrency", "8",
		"--service-zone", "my-zone",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("envoyArgs() => got %v, want %v", got, want)
	}
}

func TestEnvoyRun(t *testing.T) {
	config := model.DefaultProxyConfig()
	dir := os.Getenv("ISTIO_BIN")
	var err error
	if len(dir) == 0 {
		t.Fatalf("envoy binary dir empty")
	}
	config.BinaryPath = path.Join(dir, "envoy")

	config.ConfigPath = "tmp"

	envoyConfig := BuildConfig(config, nil)
	envoyProxy := envoy{config: config, node: "my-node", extraArgs: []string{"--mode", "validate"}}
	abortCh := make(chan error, 1)

	if err = envoyProxy.Run(nil, 0, abortCh); err == nil {
		t.Error("expected error on nil config")
	}

	if err = envoyProxy.Run(envoyConfig, 0, abortCh); err != nil {
		t.Error(err)
	}

	envoyProxy.Cleanup(0)

	badConfig := config
	badConfig.ConfigPath = ""
	envoyProxy.config = badConfig

	if err = envoyProxy.Run(envoyConfig, 0, abortCh); err == nil {
		t.Errorf("expected error on bad config path")
	}
}
