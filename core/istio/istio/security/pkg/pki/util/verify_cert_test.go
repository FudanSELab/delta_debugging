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

package util

import (
	"crypto/x509"
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"
)

func loadPEMFile(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to load the pem file = %v, err = %v", path, err)
	}
	return string(b)
}

var (
	key             = loadPEMFile("../testdata/key-10y.pem")
	keyMismatch     = loadPEMFile("../testdata/key-mismatch.pem")
	keyBad          = loadPEMFile("../testdata/key-verify-fail.pem")
	certChainBad    = loadPEMFile("../testdata/cert-verify-fail.pem")
	certChainNoRoot = loadPEMFile("../testdata/cert-noroot.pem")
	certChain       = loadPEMFile("../testdata/cert-chain-10y.pem")
	rootCertBad     = loadPEMFile("../testdata/root-verify-fail.pem")
	rootCert        = loadPEMFile("../testdata/root-cert-10y.pem")
	notBefore       = &VerifyFields{
		NotBefore: time.Unix(0, 0),
	}

	ttl = &VerifyFields{
		TTL: time.Duration(0),
	}

	extKeyUsage = &VerifyFields{
		TTL: time.Duration(1),
	}

	keyUsage = &VerifyFields{
		ExtKeyUsage: []x509.ExtKeyUsage{1, 2},
		KeyUsage:    2,
	}

	isCA = &VerifyFields{
		ExtKeyUsage: []x509.ExtKeyUsage{1, 2},
		KeyUsage:    5,
		IsCA:        true,
	}

	org = &VerifyFields{
		ExtKeyUsage: []x509.ExtKeyUsage{1, 2},
		KeyUsage:    5,
		Org:         "bad",
	}

	success = &VerifyFields{
		ExtKeyUsage: []x509.ExtKeyUsage{1, 2},
		KeyUsage:    5,
	}
)

func TestVerifyCert(t *testing.T) {
	testCases := map[string]struct {
		privPem        []byte
		certChainPem   []byte
		rootCertPem    []byte
		host           string
		expectedFields *VerifyFields
		expectedErr    string
	}{
		"Root cert bad": {
			privPem:        nil,
			certChainPem:   nil,
			rootCertPem:    []byte(rootCertBad),
			host:           "",
			expectedFields: nil,
			expectedErr:    "failed to parse root certificate",
		},
		"Cert chain bad": {
			privPem:        nil,
			certChainPem:   []byte(certChainBad),
			rootCertPem:    []byte(rootCert),
			host:           "",
			expectedFields: nil,
			expectedErr:    "failed to parse certificate chain",
		},
		"Failed to verify cert chain": {
			privPem:        nil,
			certChainPem:   []byte(certChainNoRoot),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe",
			expectedFields: nil,
			expectedErr:    "failed to verify certificate: x509:",
		},
		"Failed to verify key": {
			privPem:        []byte(keyBad),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe",
			expectedFields: nil,
			expectedErr:    "invalid PEM-encoded key",
		},
		"Failed to match key/cert": {
			privPem:        []byte(keyMismatch),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe",
			expectedFields: nil,
			expectedErr:    "the generated private key and cert doesn't match",
		},
		"Wrong SAN": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe",
			expectedFields: nil,
			expectedErr:    "the certificate doesn't have the expected SAN for: spiffe",
		},
		"Timestamp error": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe://cluster.local/ns/default/sa/default",
			expectedFields: notBefore,
			expectedErr:    "unexpected value for 'NotBefore' field",
		},
		"TTL error": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe://cluster.local/ns/default/sa/default",
			expectedFields: extKeyUsage,
			expectedErr:    "unexpected value for 'NotAfter' - 'NotBefore'",
		},
		"extKeyUsage error": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe://cluster.local/ns/default/sa/default",
			expectedFields: ttl,
			expectedErr:    "unexpected value for 'ExtKeyUsage' field",
		},
		"KeyUsage Error": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe://cluster.local/ns/default/sa/default",
			expectedFields: keyUsage,
			expectedErr:    "unexpected value for 'KeyUsage' field",
		},
		"IsCA error": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe://cluster.local/ns/default/sa/default",
			expectedFields: isCA,
			expectedErr:    "unexpected value for 'IsCA' field",
		},
		"Org error": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe://cluster.local/ns/default/sa/default",
			expectedFields: org,
			expectedErr:    "unexpected value for 'Organization' field",
		},
		"Succeeded": {
			privPem:        []byte(key),
			certChainPem:   []byte(certChain),
			rootCertPem:    []byte(rootCert),
			host:           "spiffe://cluster.local/ns/default/sa/default",
			expectedFields: success,
			expectedErr:    "",
		},
	}
	for id, tc := range testCases {
		err := VerifyCertificate(
			tc.privPem, tc.certChainPem, tc.rootCertPem, tc.host, tc.expectedFields)
		if err != nil {
			if len(tc.expectedErr) == 0 {
				t.Errorf("%s: Unexpected error: %v", id, err)
			} else if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("%s: Unexpected error: %v VS (expected) %s", id, err, tc.expectedErr)
			}
		} else if len(tc.expectedErr) != 0 {
			t.Errorf("%s: Expected error %s but succeeded", id, tc.expectedErr)
		}
	}
}
