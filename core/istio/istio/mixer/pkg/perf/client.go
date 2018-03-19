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

package perf

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	istio_mixer_v1 "istio.io/api/mixer/v1"
)

// client encapsulates a Mixer client, for the purposes of perf testing.
type client struct {
	mixer istio_mixer_v1.MixerClient
	conn  *grpc.ClientConn
	setup *Setup
}

// initialize is the first method to be called. The client is expected to perform initialization by setting up
// any local state using setup, and connecting to the Mixer rpc server at the given address.
func (c *client) initialize(address string, setup *Setup) error {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	c.mixer = istio_mixer_v1.NewMixerClient(conn)
	c.conn = conn
	c.setup = setup

	return nil
}

// close indicates that the client should perform graceful close and cleanup of resources and get ready
// to exit.
func (c *client) close() (err error) {
	if c.conn != nil {
		err = c.conn.Close()

		c.conn = nil
		c.setup = nil
		c.mixer = nil
	}

	return err
}

func (c *client) run(iterations int) (err error) {
	requests := c.setup.Load.createRequestProtos(c.setup.Config)

	for i := 0; i < iterations; i++ {

		for _, r := range requests {
			switch r.(type) {
			case *istio_mixer_v1.ReportRequest:
				_, e := c.mixer.Report(context.Background(), r.(*istio_mixer_v1.ReportRequest))
				if e != nil && err == nil {
					err = e
				}

			case *istio_mixer_v1.CheckRequest:
				_, e := c.mixer.Check(context.Background(), r.(*istio_mixer_v1.CheckRequest))
				if e != nil && err == nil {
					err = e
				}

			default:
				return fmt.Errorf("unknown request type: %v", r)
			}
		}
	}

	return err
}
