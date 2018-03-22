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

package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"istio.io/istio/pkg/log"
	"istio.io/istio/security/pkg/pki/ca"
	"istio.io/istio/security/pkg/pki/util"
	"istio.io/istio/security/pkg/registry"
	pb "istio.io/istio/security/proto"
)

const certExpirationBuffer = time.Minute

// Server implements pb.IstioCAService and provides the service on the
// specified port.
type Server struct {
	authenticators []authenticator
	authorizer     authorizer
	serverCertTTL  time.Duration
	ca             ca.CertificateAuthority
	certificate    *tls.Certificate
	hostname       string
	port           int
}

// HandleCSR handles an incoming certificate signing request (CSR). It does
// proper validation (e.g. authentication) and upon validated, signs the CSR
// and returns the resulting certificate. If not approved, reason for refusal
// to sign is returned as part of the response object.
func (s *Server) HandleCSR(ctx context.Context, request *pb.CsrRequest) (*pb.CsrResponse, error) {
	caller := s.authenticate(ctx)
	if caller == nil {
		log.Warn("request authentication failure")
		return nil, status.Error(codes.Unauthenticated, "request authenticate failure")
	}

	csr, err := util.ParsePemEncodedCSR(request.CsrPem)
	if err != nil {
		log.Warnf("CSR parsing error (error %v)", err)
		return nil, status.Errorf(codes.InvalidArgument, "CSR parsing error (%v)", err)
	}

	_, err = util.ExtractIDs(csr.Extensions)
	if err != nil {
		log.Warnf("CSR identity extraction error (%v)", err)
		return nil, status.Errorf(codes.InvalidArgument, "CSR identity extraction error (%v)", err)
	}

	cert, err := s.ca.Sign(request.CsrPem, time.Duration(request.RequestedTtlMinutes)*time.Minute, request.ForCA)
	if err != nil {
		log.Errorf("CSR signing error (%v)", err)
		return nil, status.Errorf(codes.Internal, "CSR signing error (%v)", err)
	}

	response := &pb.CsrResponse{
		IsApproved:      true,
		SignedCertChain: cert,
	}
	log.Info("CSR successfully signed.")

	return response, nil
}

// Run starts a GRPC server on the specified port.
func (s *Server) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("cannot listen on port %d (error: %v)", s.port, err)
	}

	serverOption := s.createTLSServerOption()

	grpcServer := grpc.NewServer(serverOption)
	pb.RegisterIstioCAServiceServer(grpcServer, s)

	// grpcServer.Serve() is a blocking call, so run it in a goroutine.
	go func() {
		log.Infof("Starting GRPC server on port %d", s.port)

		err := grpcServer.Serve(listener)

		// grpcServer.Serve() always returns a non-nil error.
		log.Warnf("GRPC server returns an error: %v", err)
	}()

	return nil
}

// New creates a new instance of `IstioCAServiceServer`.
func New(ca ca.CertificateAuthority, ttl time.Duration, hostname string, port int) *Server {
	// Notice that the order of authenticators matters, since at runtime
	// authenticators are actived sequentially and the first successful attempt
	// is used as the authentication result.
	authenticators := []authenticator{&clientCertAuthenticator{}}
	aud := fmt.Sprintf("grpc://%s:%d", hostname, port)
	if jwtAuthenticator, err := newIDTokenAuthenticator(aud); err != nil {
		log.Errorf("failed to create JWT authenticator (error %v)", err)
	} else {
		authenticators = append(authenticators, jwtAuthenticator)
	}

	return &Server{
		authenticators: authenticators,
		authorizer:     &registryAuthorizor{registry.GetIdentityRegistry()},
		serverCertTTL:  ttl,
		ca:             ca,
		hostname:       hostname,
		port:           port,
	}
}

func (s *Server) createTLSServerOption() grpc.ServerOption {
	cp := x509.NewCertPool()
	cp.AppendCertsFromPEM(s.ca.GetRootCertificate())

	config := &tls.Config{
		ClientCAs:  cp,
		ClientAuth: tls.VerifyClientCertIfGiven,
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			if s.certificate == nil || shouldRefresh(s.certificate) {
				// Apply new certificate if there isn't one yet, or the one has become invalid.
				newCert, err := s.applyServerCertificate()
				if err != nil {
					return nil, fmt.Errorf("failed to apply TLS server certificate (%v)", err)
				}
				s.certificate = newCert
			}
			return s.certificate, nil
		},
	}
	return grpc.Creds(credentials.NewTLS(config))
}

func (s *Server) applyServerCertificate() (*tls.Certificate, error) {
	opts := util.CertOptions{
		Host:       s.hostname,
		RSAKeySize: 2048,
	}

	csrPEM, privPEM, err := util.GenCSR(opts)
	if err != nil {
		return nil, err
	}

	certPEM, err := s.ca.Sign(csrPEM, s.serverCertTTL, false)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certPEM, privPEM)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (s *Server) authenticate(ctx context.Context) *caller {
	// TODO: apply different authenticators in specific order / according to configuration.
	for _, authn := range s.authenticators {
		if u, _ := authn.authenticate(ctx); u != nil {
			return u
		}
	}
	return nil
}

// shouldRefresh indicates whether the given certificate should be refreshed.
func shouldRefresh(cert *tls.Certificate) bool {
	// Check whether there is a valid leaf certificate.
	leaf := cert.Leaf
	if leaf == nil {
		return true
	}

	// Check whether the leaf certificate is about to expire.
	return leaf.NotAfter.Add(-certExpirationBuffer).Before(time.Now())
}
