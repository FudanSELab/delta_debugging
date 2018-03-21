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
	"strings"
	"testing"
)

const (
	rootCertFile        = "../testdata/multilevelpki/root-cert.pem"
	rootKeyFile         = "../testdata/multilevelpki/root-key.pem"
	intCertFile         = "../testdata/multilevelpki/int-cert.pem"
	intKeyFile          = "../testdata/multilevelpki/int-key.pem"
	intCertChainFile    = "../testdata/multilevelpki/int-cert-chain.pem"
	int2CertFile        = "../testdata/multilevelpki/int2-cert.pem"
	int2KeyFile         = "../testdata/multilevelpki/int2-key.pem"
	int2CertChainFile   = "../testdata/multilevelpki/int2-cert-chain.pem"
	badCertFile         = "../testdata/cert-parse-fail.pem"
	badKeyFile          = "../testdata/key-parse-fail.pem"
	anotherKeyFile      = "../testdata/key.pem"
	anotherRootCertFile = "../testdata/cert.pem"
)

func TestKeyCertBundleWithRootCertFromFile(t *testing.T) {
	testCases := map[string]struct {
		rootCertFile string
		expectedErr  string
	}{
		"File not found": {
			rootCertFile: "bad.pem",
			expectedErr:  "open bad.pem: no such file or directory",
		},
		"With root cert": {
			rootCertFile: rootCertFile,
			expectedErr:  "",
		},
	}
	for id, tc := range testCases {
		bundle, err := NewKeyCertBundleWithRootCertFromFile(tc.rootCertFile)
		if err != nil {
			if len(tc.expectedErr) == 0 {
				t.Errorf("%s: Unexpected error: %v", id, err)
			} else if strings.Compare(err.Error(), tc.expectedErr) != 0 {
				t.Errorf("%s: Unexpected error: %v VS (expected) %s", id, err, tc.expectedErr)
			}
		} else if len(tc.expectedErr) != 0 {
			t.Errorf("%s: Expected error %s but succeeded", id, tc.expectedErr)
		} else if bundle == nil {
			t.Errorf("%s: the bundle should not be empty", id)
		} else {
			cert, key, chain, root := bundle.GetAllPem()
			if len(cert) != 0 {
				t.Errorf("%s: certBytes should be empty", id)
			}
			if len(key) != 0 {
				t.Errorf("%s: privateKeyBytes should be empty", id)
			}
			if len(chain) != 0 {
				t.Errorf("%s: certChainBytes should be empty", id)
			}
			if len(root) == 0 {
				t.Errorf("%s: rootCertBytes should not be empty", id)
			}

			x509Cert, privKey, chain, root := bundle.GetAll()
			if x509Cert != nil {
				t.Errorf("%s: cert should be nil", id)
			}
			if privKey != nil {
				t.Errorf("%s: private key should be nil", id)
			}
			if len(chain) != 0 {
				t.Errorf("%s: certChainBytes should be empty", id)
			}
			if len(root) == 0 {
				t.Errorf("%s: rootCertBytes should not be empty", id)
			}
		}
	}
}

// The test of NewVerifiedKeyCertBundleFromPem, VerifyAndSetAll can be covered by this test.
func TestNewVerifiedKeyCertBundleFromFile(t *testing.T) {
	testCases := map[string]struct {
		caCertFile    string
		caKeyFile     string
		certChainFile string
		rootCertFile  string
		expectedErr   string
	}{
		"Success - 1 level CA": {
			caCertFile:    rootCertFile,
			caKeyFile:     rootKeyFile,
			certChainFile: "",
			rootCertFile:  rootCertFile,
			expectedErr:   "",
		},
		"Success - 2 level CA": {
			caCertFile:    intCertFile,
			caKeyFile:     intKeyFile,
			certChainFile: intCertChainFile,
			rootCertFile:  rootCertFile,
			expectedErr:   "",
		},
		"Success - 3 level CA": {
			caCertFile:    int2CertFile,
			caKeyFile:     int2KeyFile,
			certChainFile: int2CertChainFile,
			rootCertFile:  rootCertFile,
			expectedErr:   "",
		},
		"Success - 2 level CA without cert chain file": {
			caCertFile:    intCertFile,
			caKeyFile:     intKeyFile,
			certChainFile: "",
			rootCertFile:  rootCertFile,
			expectedErr:   "",
		},
		"Failure - invalid cert chain file": {
			caCertFile:    intCertFile,
			caKeyFile:     intKeyFile,
			certChainFile: "bad.pem",
			rootCertFile:  rootCertFile,
			expectedErr:   "open bad.pem: no such file or directory",
		},
		"Failure - no root cert file": {
			caCertFile:    intCertFile,
			caKeyFile:     intKeyFile,
			certChainFile: "",
			rootCertFile:  "bad.pem",
			expectedErr:   "open bad.pem: no such file or directory",
		},
		"Failure - cert and key do not match": {
			caCertFile:    int2CertFile,
			caKeyFile:     anotherKeyFile,
			certChainFile: int2CertChainFile,
			rootCertFile:  rootCertFile,
			expectedErr:   "the cert does not match the key",
		},
		"Failure - 3 level CA without cert chain file": {
			caCertFile:    int2CertFile,
			caKeyFile:     int2KeyFile,
			certChainFile: "",
			rootCertFile:  rootCertFile,
			expectedErr:   "cannot verify the cert with the provided root chain and cert pool",
		},
		"Failure - cert not verifiable from root cert": {
			caCertFile:    intCertFile,
			caKeyFile:     intKeyFile,
			certChainFile: intCertChainFile,
			rootCertFile:  anotherRootCertFile,
			expectedErr:   "cannot verify the cert with the provided root chain and cert pool",
		},
		"Failure - invalid cert": {
			caCertFile:    badCertFile,
			caKeyFile:     intKeyFile,
			certChainFile: "",
			rootCertFile:  rootCertFile,
			expectedErr:   "failed to parse cert PEM: invalid PEM encoded certificate",
		},
		"Failure - not existing private key": {
			caCertFile:    intCertFile,
			caKeyFile:     "bad.pem",
			certChainFile: "",
			rootCertFile:  rootCertFile,
			expectedErr:   "open bad.pem: no such file or directory",
		},
		"Failure - invalid private key": {
			caCertFile:    intCertFile,
			caKeyFile:     badKeyFile,
			certChainFile: "",
			rootCertFile:  rootCertFile,
			expectedErr:   "failed to parse private key PEM: invalid PEM-encoded key",
		},
		"Failure - file does not exist": {
			caCertFile:    "random/path/does/not/exist",
			caKeyFile:     intKeyFile,
			certChainFile: "",
			rootCertFile:  rootCertFile,
			expectedErr:   "open random/path/does/not/exist: no such file or directory",
		},
	}
	for id, tc := range testCases {
		_, err := NewVerifiedKeyCertBundleFromFile(
			tc.caCertFile, tc.caKeyFile, tc.certChainFile, tc.rootCertFile)
		if err != nil {
			if len(tc.expectedErr) == 0 {
				t.Errorf("%s: Unexpected error: %v", id, err)
			} else if strings.Compare(err.Error(), tc.expectedErr) != 0 {
				t.Errorf("%s: Unexpected error: %v VS (expected) %s", id, err, tc.expectedErr)
			}
		} else if len(tc.expectedErr) != 0 {
			t.Errorf("%s: Expected error %s but succeeded", id, tc.expectedErr)
		}
	}
}
