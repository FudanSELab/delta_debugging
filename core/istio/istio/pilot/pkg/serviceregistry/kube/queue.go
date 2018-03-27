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

package kube

import (
	"sync"
	"time"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	"k8s.io/client-go/util/flowcontrol"

	"os"
	"strconv"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
)

const (
	// enableQueueThrottleEnv is an environment variable that can be set to re-enable the
	// throttling, in case problems are discovered. This is not a flag - it would cause too
	// many changes, and it is only intended as a short term fail-safe and A/B testing.
	// TODO: remove in 0.7 after more testing without the throttle.
	enableQueueThrottleEnv = "PILOT_THROTTLE"
)

// Queue of work tickets processed using a rate-limiting loop
type Queue interface {
	// Push a ticket
	Push(Task)
	// Run the loop until a signal on the channel
	Run(<-chan struct{})
}

// Handler specifies a function to apply on an object for a given event type
type Handler func(obj interface{}, event model.Event) error

// Task object for the event watchers; processes until handler succeeds
type Task struct {
	handler Handler
	obj     interface{}
	event   model.Event
}

// NewTask creates a task from a work item
func NewTask(handler Handler, obj interface{}, event model.Event) Task {
	return Task{handler: handler, obj: obj, event: event}
}

type queueImpl struct {
	delay   time.Duration
	queue   []Task
	lock    sync.Mutex
	closing bool
}

// NewQueue instantiates a queue with a processing function
func NewQueue(errorDelay time.Duration) Queue {
	return &queueImpl{
		delay:   errorDelay,
		queue:   make([]Task, 0),
		closing: false,
		lock:    sync.Mutex{},
	}
}

func (q *queueImpl) Push(item Task) {
	q.lock.Lock()
	if !q.closing {
		q.queue = append(q.queue, item)
	}
	q.lock.Unlock()
}

func (q *queueImpl) Run(stop <-chan struct{}) {
	go func() {
		<-stop
		q.lock.Lock()
		q.closing = true
		q.lock.Unlock()
	}()

	rate := os.Getenv(enableQueueThrottleEnv)
	rateLimit := 100
	if len(rate) > 0 {
		r, err := strconv.Atoi(rate)
		if err == nil {
			rateLimit = r
		}
	}
	// Throttle processing up to smoothed 10 qps with bursts up to 100 qps
	var rateLimiter flowcontrol.RateLimiter
	if rateLimit > 0 {
		rateLimiter = flowcontrol.NewTokenBucketRateLimiter(float32(rateLimit), 10*rateLimit)
	}

	var item Task
	for {
		if rateLimit > 0 {
			rateLimiter.Accept()
		}

		q.lock.Lock()
		if q.closing {
			q.lock.Unlock()
			return
		} else if len(q.queue) == 0 {
			q.lock.Unlock()
		} else {
			item, q.queue = q.queue[0], q.queue[1:]
			q.lock.Unlock()

			for {
				err := item.handler(item.obj, item.event)
				if err != nil {
					log.Infof("Work item failed (%v), repeating after delay %v", err, q.delay)
					time.Sleep(q.delay)
				} else {
					break
				}
			}
		}
	}
}

// ChainHandler applies handlers in a sequence
type ChainHandler struct {
	funcs []Handler
}

// Apply is the handler function
func (ch *ChainHandler) Apply(obj interface{}, event model.Event) error {
	for _, f := range ch.funcs {
		if err := f(obj, event); err != nil {
			return err
		}
	}
	return nil
}

// Append a handler as the last handler in the chain
func (ch *ChainHandler) Append(h Handler) {
	ch.funcs = append(ch.funcs, h)
}
