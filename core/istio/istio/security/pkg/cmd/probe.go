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

package cmd

import (
	"github.com/spf13/cobra"

	"istio.io/istio/mixer/cmd/shared"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/probe"
)

// NewProbeCmd creates the cobra.Command for the probe command
func NewProbeCmd() *cobra.Command {
	logOptions := log.DefaultOptions()
	probeOptions := &probe.Options{}
	cmd := &cobra.Command{
		Use:   "probe",
		Short: "Check the liveness or readiness of a locally-running server",
		Run: func(cmd *cobra.Command, args []string) {
			if !probeOptions.IsValid() {
				shared.Fatalf("probe-path or interval are not valid\n\n%s", cmd.UsageString())
			}
			if err := probe.NewFileClient(probeOptions).GetStatus(); err != nil {
				shared.Fatalf("Fail on inspecting path %s: %v", probeOptions.Path, err)
			}
			shared.Printf("OK")
		},
	}
	logOptions.AttachCobraFlags(cmd)
	cmd.PersistentFlags().StringVar(&probeOptions.Path, "probe-path", "", "Path of the file for checking the availability.")
	cmd.PersistentFlags().DurationVar(&probeOptions.UpdateInterval, "interval", 0, "Duration used for checking the target file's last modified time.")
	return cmd
}
