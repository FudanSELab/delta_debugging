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

package workloadapi

import (
	"fmt"
	"log"

	rpc "github.com/gogo/googleapis/google/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	wlh "istio.io/istio/security/cmd/node_agent_k8s/workload/handler"
	pb "istio.io/istio/security/proto"
)

// server implements WorkloadService from workload_service.proto.
type server struct{}

// RegisterGrpc register grpc
func RegisterGrpc(s *grpc.Server) {
	pb.RegisterWorkloadServiceServer(s, &server{})
}

// Check do the check
func (s *server) Check(ctx context.Context, request *pb.CheckRequest) (*pb.CheckResponse, error) {

	log.Printf("[%v]: %v Check called", s, request)
	// Get the caller's credentials from the context.
	creds, e := wlh.CallerFromContext(ctx)
	if !e {
		resp := fmt.Sprint("Not able to get credentials")
		status := &rpc.Status{Code: int32(rpc.PERMISSION_DENIED), Message: resp}
		return &pb.CheckResponse{Status: status}, nil
	}

	log.Printf("Credentials are %v", creds)

	resp := fmt.Sprintf("all good to workload with service account %v", creds.ServiceAccount)
	status := &rpc.Status{Code: int32(rpc.OK), Message: resp}
	return &pb.CheckResponse{Status: status}, nil
}
