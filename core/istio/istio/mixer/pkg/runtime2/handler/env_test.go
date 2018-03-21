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

package handler

import (
	"sync"
	"testing"
	"time"

	"istio.io/istio/mixer/pkg/pool"
	"istio.io/istio/pkg/log"
)

func TestEnv(t *testing.T) {
	for i := 0; i < 2; i++ {
		gp := pool.NewGoroutinePool(128, i == 0)
		gp.AddWorkers(32)

		// set up the ambient logger so newEnv picks it up
		o := log.DefaultOptions()
		_ = log.Configure(o)

		env := newEnv(0, "Foo", gp)
		log := env.Logger()
		log.Infof("Test%s", "ing")
		log.Warningf("Test%s", "ing")
		err := log.Errorf("Test%s", "ing")
		if err == nil {
			t.Error("Expected an error but got nil")
		}

		if log.VerbosityLevel(999) {
			t.Error("Expected false for verbosity level 999 check")
		}
		if !log.VerbosityLevel(0) {
			t.Error("Expected true for verbosity level 0 check")
		}
		if !log.VerbosityLevel(1) {
			t.Error("Expected true for verbosity level 1 check")
		}
		if !log.VerbosityLevel(2) {
			t.Error("Expected true for verbosity level 2 check")
		}
		if log.VerbosityLevel(-1) {
			t.Error("Expected false for verbosity level -1 check")
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		env.ScheduleWork(func() {
			wg.Done()
		})

		wg.Add(1)
		env.ScheduleDaemon(func() {
			wg.Done()
		})

		env.ScheduleWork(func() {
			panic("bye!")
		})

		env.ScheduleDaemon(func() {
			panic("bye!")
		})

		wg.Wait()

		// hack to give time for the panic to 'take hold' if it doesn't get recovered properly
		time.Sleep(200 * time.Millisecond)

		_ = gp.Close()
	}
}
