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

package cloudfoundry

import (
	"reflect"
	"time"

	copilotapi "code.cloudfoundry.org/copilot/api"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	"golang.org/x/net/context"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
)

type serviceHandler func(*model.Service, model.Event)
type instanceHandler func(*model.ServiceInstance, model.Event)

// Ticker acts like time.Ticker but is mockable for testing
type Ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type realTicker struct {
	*time.Ticker
}

func (r realTicker) Chan() <-chan time.Time {
	return r.C
}

// NewTicker returns a Ticker used to instantiate a Controller
func NewTicker(d time.Duration) Ticker {
	return realTicker{time.NewTicker(d)}
}

// Controller communicates with Cloud Foundry and monitors for changes
type Controller struct {
	Client           copilotapi.IstioCopilotClient
	Ticker           Ticker
	serviceHandlers  []serviceHandler
	instanceHandlers []instanceHandler
}

// AppendServiceHandler implements a service catalog operation
func (c *Controller) AppendServiceHandler(f func(*model.Service, model.Event)) error {
	c.serviceHandlers = append(c.serviceHandlers, f)
	return nil
}

// AppendInstanceHandler implements a service catalog operation
func (c *Controller) AppendInstanceHandler(f func(*model.ServiceInstance, model.Event)) error {
	c.instanceHandlers = append(c.instanceHandlers, f)
	return nil
}

// Run will loop, calling handlers in response to changes, until a signal is received
func (c *Controller) Run(stop <-chan struct{}) {
	cache := &copilotapi.RoutesResponse{}
	for {
		select {
		case <-c.Ticker.Chan():
			backendSets, err := c.Client.Routes(context.Background(), &copilotapi.RoutesRequest{})
			if err != nil {
				log.Warnf("periodic copilot routes poll failed: %s", err)
				continue
			}

			if !reflect.DeepEqual(backendSets, cache) {
				cache = backendSets
				// Clear service discovery cache
				for _, h := range c.serviceHandlers {
					go h(&model.Service{}, model.EventAdd)
				}
				for _, h := range c.instanceHandlers {
					go h(&model.ServiceInstance{}, model.EventAdd)
				}
			}
		case <-stop:
			c.Ticker.Stop()
			return
		}
	}
}
