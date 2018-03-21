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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"istio.io/istio/security/pkg/pki/ca"
	mockca "istio.io/istio/security/pkg/pki/ca/mock"
	mockutil "istio.io/istio/security/pkg/pki/util/mock"
	pb "istio.io/istio/security/proto"
)

const csr = `
-----BEGIN CERTIFICATE REQUEST-----
MIIBoTCCAQoCAQAwEzERMA8GA1UEChMISnVqdSBvcmcwgZ8wDQYJKoZIhvcNAQEB
BQADgY0AMIGJAoGBANFf06eqiDx0+qD/xBAR5aMwwgaBOn6TPfSy96vOxLTsfkTg
ir/vb8UG+F5hO6yxF+z2BgzD8LwcbKnxahoPq/aWGLw3Umcqm4wxgWKHxvtYSQDG
w4zpmKOqgkagxbx32JXDlMpi6adUVHNvB838CiUys6IkVB0obGHnre8zmCLdAgMB
AAGgTjBMBgkqhkiG9w0BCQ4xPzA9MDsGA1UdEQQ0MDKGMHNwaWZmZTovL3Rlc3Qu
Y29tL25hbWVzcGFjZS9ucy9zZXJ2aWNlYWNjb3VudC9zYTANBgkqhkiG9w0BAQsF
AAOBgQCw9dL6xRQSjdYKt7exqlTJliuNEhw/xDVGlNUbDZnT0uL3zXI//Z8tsejn
8IFzrDtm0Z2j4BmBzNMvYBKL/4JPZ8DFywOyQqTYnGtHIkt41CNjGfqJRk8pIqVC
hKldzzeCKNgztEvsUKVqltFZ3ZYnkj/8/Cg8zUtTkOhHOjvuig==
-----END CERTIFICATE REQUEST-----`

const badSanCsr = `
MIICdzCCAV8CAQAwCzEJMAcGA1UEChMAMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAr8uTt9MSXAHugljyfxCS1BE3X0U5YQnN8Cgj1qn5cnu43LDdwA/x
Zgsd7ZfkuA+fpxBW2x4yR4LOSEwZAav6z45f9dxoZea0/wTPUHXam2tHIuhz1F1F
LlZX0EbZErBcjiPs6Y/FUaVROZZftOkq+sfNExTiXR7q5fAyYP/9L57OHOEx6RA3
kNEFBaa190j4ITvuS8fqsMT3lsRqLQ7fTCd5Ygw8rGZWOT6GSpLm1YJvSXhDUdxL
hYvoMoDgJ+SRpXWvG/YlzP6nMvJN45flTcIGXSMFvqaFGs5HhYxIviX8dE1Vso+/
1GV5MNPksTuGh/QqCjjcKvzZ6cMRuUeziQIDAQABoCcwJQYJKoZIhvcNAQkOMRgw
FjAUBgNVHREEDTALgglsb2NhbGhvc3QwDQYJKoZIhvcNAQELBQADggEBAEHIduLz
5oei9NHapvYsDDe6A+Q2nUm9uvWn/mMBujbstY9ZmLc73gWS0A8maXFFCjtMf7+n
u8naR7rmw0MjbVJPL2gbbqWjlNqvfm/upiYT2o8UtXyi0ZIQwfxL/iLqHZOVfm//
GGpTOohc7joR0EUnBa5piK3XXc4U5aCWMwlnmENMBAtNlRBuAzYJsMydv0Be72ga
gCojNs0xyJ77JA80HLY7iR4J6BRYsZQ/5UB/pYR55e4TGFDbI+C/6NBqLkzEfyX0
5KLq/6IJesVZLnKoxOt07OYriZS+U4b+Lx3++vWVnI8z2iOdGPUuJj7ys57zKJZ3
1sT/u25qExkefck=
-----END CERTIFICATE REQUEST-----`

type mockCA struct {
	cert      string
	root      string
	certChain string
	errMsg    string
}

func (ca *mockCA) Sign(csrPEM []byte, ttl time.Duration, forCA bool) ([]byte, error) {
	if ca.errMsg != "" {
		return nil, fmt.Errorf(ca.errMsg)
	}
	return []byte(ca.cert), nil
}

func (ca *mockCA) GetRootCertificate() []byte {
	return []byte(ca.root)
}

func (ca *mockCA) GetCertChain() []byte {
	return []byte(ca.certChain)
}

type mockAuthenticator struct {
	authSource authSource
	identities []string
	errMsg     string
}

func (authn *mockAuthenticator) authenticate(ctx context.Context) (*caller, error) {
	if len(authn.errMsg) > 0 {
		return nil, fmt.Errorf("%v", authn.errMsg)
	}

	return &caller{
		authSource: authn.authSource,
		identities: authn.identities,
	}, nil
}

type mockAuthorizer struct {
	errMsg string
}

func (authz *mockAuthorizer) authorize(requester *caller, requestedIds []string) error {
	if len(authz.errMsg) > 0 {
		return fmt.Errorf("%v", authz.errMsg)
	}
	return nil
}

func TestSign(t *testing.T) {
	testCases := map[string]struct {
		authenticators []authenticator
		authorizer     *mockAuthorizer
		ca             ca.CertificateAuthority
		csr            string
		cert           string
		certChain      string
		code           codes.Code
	}{
		"No authenticator": {
			authenticators: nil,
			code:           codes.Unauthenticated,
			authorizer:     &mockAuthorizer{},
			ca:             &mockca.FakeCA{SignErr: fmt.Errorf("cannot sign")},
		},
		"Unauthenticated request": {
			authenticators: []authenticator{&mockAuthenticator{
				errMsg: "Not authorized",
			}},
			code:       codes.Unauthenticated,
			authorizer: &mockAuthorizer{},
			ca:         &mockca.FakeCA{SignErr: fmt.Errorf("cannot sign")},
		},
		"Corrupted CSR": {
			authorizer:     &mockAuthorizer{},
			authenticators: []authenticator{&mockAuthenticator{}},
			ca:             &mockca.FakeCA{SignErr: fmt.Errorf("cannot sign")},
			csr:            "deadbeef",
			code:           codes.InvalidArgument,
		},
		"Invalid SAN CSR": {
			authorizer:     &mockAuthorizer{},
			authenticators: []authenticator{&mockAuthenticator{}},
			ca:             &mockca.FakeCA{SignErr: fmt.Errorf("cannot sign")},
			csr:            badSanCsr,
			code:           codes.InvalidArgument,
		},
		"Failed to sign": {
			authorizer:     &mockAuthorizer{},
			authenticators: []authenticator{&mockAuthenticator{}},
			ca:             &mockca.FakeCA{SignErr: fmt.Errorf("cannot sign")},
			csr:            csr,
			code:           codes.Internal,
		},
		"Successful signing": {
			authenticators: []authenticator{&mockAuthenticator{}},
			authorizer:     &mockAuthorizer{},
			ca: &mockca.FakeCA{
				SignedCert:    []byte("generated cert"),
				KeyCertBundle: &mockutil.FakeKeyCertBundle{CertChainBytes: []byte("cert chain")},
			},
			csr:       csr,
			cert:      "generated cert",
			certChain: "cert chain",
			code:      codes.OK,
		},
	}

	for id, c := range testCases {
		server := &Server{
			ca:             c.ca,
			hostname:       "hostname",
			port:           8080,
			authorizer:     c.authorizer,
			authenticators: c.authenticators,
		}
		request := &pb.CsrRequest{CsrPem: []byte(c.csr)}

		response, err := server.HandleCSR(context.Background(), request)
		s, _ := status.FromError(err)
		code := s.Code()
		if c.code != code {
			t.Errorf("Case %s: expecting code to be (%d) but got (%d: %s)", id, c.code, code, s.Message())
		} else if c.code == codes.OK {
			if !bytes.Equal(response.SignedCert, []byte(c.cert)) {
				t.Errorf("Case %s: expecting cert to be (%s) but got (%s)", id, c.cert, response.SignedCert)
			}
			if !bytes.Equal(response.CertChain, []byte(c.certChain)) {
				t.Errorf("Case %s: expecting cert chain to be (%s) but got (%s)", id, c.certChain, response.CertChain)
			}

		}
	}
}

func TestShouldRefresh(t *testing.T) {
	now := time.Now()
	testCases := map[string]struct {
		cert          *tls.Certificate
		shouldRefresh bool
	}{
		"No leaf cert": {
			cert:          &tls.Certificate{},
			shouldRefresh: true,
		},
		"Cert is expired": {
			cert: &tls.Certificate{
				Leaf: &x509.Certificate{NotAfter: now},
			},
			shouldRefresh: true,
		},
		"Cert is about to expire": {
			cert: &tls.Certificate{
				Leaf: &x509.Certificate{NotAfter: now.Add(5 * time.Second)},
			},
			shouldRefresh: true,
		},
		"Cert is valid": {
			cert: &tls.Certificate{
				Leaf: &x509.Certificate{NotAfter: now.Add(5 * time.Minute)},
			},
			shouldRefresh: false,
		},
	}

	for id, tc := range testCases {
		result := shouldRefresh(tc.cert)
		if tc.shouldRefresh != result {
			t.Errorf("%s: expected result is %t but got %t", id, tc.shouldRefresh, result)
		}
	}
}

func TestRun(t *testing.T) {
	testCases := map[string]struct {
		ca                          *mockca.FakeCA
		hostname                    string
		port                        int
		expectedErr                 string
		applyServerCertificateError string
		expectedAuthenticatorsLen   int
	}{
		"Invalid listening port number": {
			ca:          &mockca.FakeCA{SignedCert: []byte(csr)},
			port:        -1,
			expectedErr: "cannot listen on port -1 (error: listen tcp: address -1: invalid port)",
		},
		"CA sign error": {
			ca:                          &mockca.FakeCA{SignErr: errors.New("mock CA cannot sign")},
			hostname:                    "localhost",
			port:                        0,
			expectedErr:                 "",
			expectedAuthenticatorsLen:   2,
			applyServerCertificateError: "mock CA cannot sign",
		},
		"Bad signed cert": {
			ca:                        &mockca.FakeCA{SignedCert: []byte(csr)},
			hostname:                  "localhost",
			port:                      0,
			expectedErr:               "",
			expectedAuthenticatorsLen: 2,
			applyServerCertificateError: "tls: failed to find \"CERTIFICATE\" PEM block in certificate " +
				"input after skipping PEM blocks of the following types: [CERTIFICATE REQUEST]",
		},
	}

	for id, tc := range testCases {
		server := New(tc.ca, time.Hour, tc.hostname, tc.port)
		err := server.Run()
		if len(tc.expectedErr) > 0 {
			if err == nil {
				t.Errorf("%s: Succeeded. Error expected: %v", id, err)
			} else if err.Error() != tc.expectedErr {
				t.Errorf("%s: incorrect error message: %s VS %s",
					id, err.Error(), tc.expectedErr)
			}
			continue
		} else if err != nil {
			t.Fatalf("%s: Unexpected Error: %v", id, err)
		}

		if len(server.authenticators) != tc.expectedAuthenticatorsLen {
			t.Fatalf("%s: Unexpected Authenticators Length. Expected: %v Actual: %v",
				id, tc.expectedAuthenticatorsLen, len(server.authenticators))
		}

		_, err = server.applyServerCertificate()
		if len(tc.applyServerCertificateError) > 0 {
			if err == nil {
				t.Errorf("%s: Succeeded. Error expected: %v", id, err)
			} else if err.Error() != tc.applyServerCertificateError {
				t.Errorf("%s: incorrect error message: %s VS %s",
					id, err.Error(), tc.applyServerCertificateError)
			}
			continue
		} else if err != nil {
			t.Fatalf("%s: Unexpected Error: %v", id, err)
		}
	}
}
