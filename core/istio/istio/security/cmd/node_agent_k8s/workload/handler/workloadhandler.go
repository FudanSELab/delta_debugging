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
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	pbmgmt "istio.io/istio/security/proto"
)

// WorkloadHandler support this given interface.
// nodeagentmgmt invokes:
// - Serve() as a go routine when a Workload is added.
// - Stop() when a Workload is deleted.
// - WaitDone() to wait for a response back from Workloadhandler
type WorkloadHandler interface {
	Serve()
	Stop()
	WaitDone()
}

// RegisterGrpcServer is used by WorkloadAPI to register itself as the grpc server.
// It is invoked by the workload handler when it is initializing the workload socket.
type RegisterGrpcServer func(s *grpc.Server)

// Options contains the configuration for the workload service.
type Options struct {
	// PathPrefix is the uds path prefix for each workload service.
	PathPrefix string
	// Sockfile is the the uds file name for each workload service.
	SockFile string
	// RegAPI is the callback to invoke to connect the gRPC Server.
	RegAPI RegisterGrpcServer
}

// handler implements the WorkloadHandler (one per workload).
type handler struct {
	creds    *CredInfo
	filePath string
	done     chan bool
	regAPI   RegisterGrpcServer
}

// NewCreds creates the CredInfo.
func NewCreds(wli *pbmgmt.WorkloadInfo) *CredInfo {
	return &CredInfo{
		UID:            wli.Attrs.Uid,
		Name:           wli.Attrs.Workload,
		Namespace:      wli.Attrs.Namespace,
		ServiceAccount: wli.Attrs.Serviceaccount,
	}
}

// NewHandler returns the new server with default setup.
func NewHandler(wli *pbmgmt.WorkloadInfo, options Options) WorkloadHandler {
	if options.RegAPI == nil {
		return nil
	}
	s := &handler{
		done:     make(chan bool, 1),
		creds:    NewCreds(wli),
		filePath: options.PathPrefix + "/" + wli.Attrs.Uid + options.SockFile,
		regAPI:   options.RegAPI,
	}
	return s
}

// Serve adherence to nodeagent workload management interface.
func (s *handler) Serve() {
	grpcServer := grpc.NewServer(grpc.Creds(s.GetCred()))
	//s.wlS.RegAPI(grpcServer)
	s.regAPI(grpcServer)

	var lis net.Listener
	var err error
	_, e := os.Stat(s.filePath)
	if e == nil {
		e := os.RemoveAll(s.filePath)
		if e != nil {
			log.Printf("Failed to rm %v (%v)", s.filePath, e)
			return
		}
	}

	lis, err = net.Listen("unix", s.filePath)
	if err != nil {
		log.Printf("failed to %v", err)
		return
	}

	go func(ln net.Listener, c chan bool) {
		<-c
		_ = ln.Close()
		log.Printf("Closed the listener.")
		c <- true
	}(lis, s.done)

	log.Printf("workload [%v] listen", s)
	_ = grpcServer.Serve(lis)
}

// Stop tells the server it should stop
func (s *handler) Stop() {
	s.done <- true
}

// WaitDone notifies the handler to stop.
func (s *handler) WaitDone() {
	<-s.done
}
