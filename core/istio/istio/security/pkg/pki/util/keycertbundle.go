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

// Provides utility methods to generate X.509 certificates with different
// options. This implementation is Largely inspired from
// https://golang.org/src/crypto/tls/generate_cert.go.

package util

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync"
)

// KeyCertBundle stores the cert, private key, cert chain and root cert for an entity. It is thread safe.
type KeyCertBundle interface {
	// GetAllPem returns all key/cert PEMs in KeyCertBundle together. Getting all values together avoids inconsistency.
	GetAllPem() (certBytes, privKeyBytes, certChainBytes, rootCertBytes []byte)

	// GetAll returns all key/cert in KeyCertBundle together. Getting all values together avoids inconsistency.
	GetAll() (cert *x509.Certificate, privKey *crypto.PrivateKey, certChainBytes, rootCertBytes []byte)

	// VerifyAndSetAll verifies the key/certs, and sets all key/certs in KeyCertBundle together.
	// Setting all values together avoids inconsistency.
	VerifyAndSetAll(certBytes, privKeyBytes, certChainBytes, rootCertBytes []byte) error
}

// KeyCertBundleImpl implements the KeyCertBundle interface.
// The cert and privKey should be a public/private key pair.
// The cert should be verifiable from the rootCert through the certChain.
// cert and priveKey are pointers to the cert/key parsed from certBytes/privKeyBytes.
type KeyCertBundleImpl struct {
	certBytes      []byte
	cert           *x509.Certificate
	privKeyBytes   []byte
	privKey        *crypto.PrivateKey
	certChainBytes []byte
	rootCertBytes  []byte
	// mutex protects the R/W to all keys and certs.
	mutex sync.RWMutex
}

// NewVerifiedKeyCertBundleFromPem returns a new KeyCertBundle, or error if if the provided certs failed the
// verification.
func NewVerifiedKeyCertBundleFromPem(certBytes, privKeyBytes, certChainBytes, rootCertBytes []byte) (
	*KeyCertBundleImpl, error) {
	bundle := &KeyCertBundleImpl{}
	if err := bundle.VerifyAndSetAll(certBytes, privKeyBytes, certChainBytes, rootCertBytes); err != nil {
		return nil, err
	}
	return bundle, nil
}

// NewVerifiedKeyCertBundleFromFile returns a new KeyCertBundle, or error if if the provided certs failed the
// verification.
func NewVerifiedKeyCertBundleFromFile(certFile, privKeyFile, certChainFile, rootCertFile string) (
	*KeyCertBundleImpl, error) {
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	privKeyBytes, err := ioutil.ReadFile(privKeyFile)
	if err != nil {
		return nil, err
	}
	certChainBytes := []byte{}
	if len(certChainFile) != 0 {
		if certChainBytes, err = ioutil.ReadFile(certChainFile); err != nil {
			return nil, err
		}
	}
	rootCertBytes, err := ioutil.ReadFile(rootCertFile)
	if err != nil {
		return nil, err
	}
	return NewVerifiedKeyCertBundleFromPem(certBytes, privKeyBytes, certChainBytes, rootCertBytes)
}

// NewKeyCertBundleWithRootCertFromFile returns a new KeyCertBundle with the root cert without verification.
func NewKeyCertBundleWithRootCertFromFile(rootCertFile string) (*KeyCertBundleImpl, error) {
	rootCertBytes, err := ioutil.ReadFile(rootCertFile)
	if err != nil {
		return nil, err
	}
	return &KeyCertBundleImpl{
		certBytes:      []byte{},
		cert:           nil,
		privKeyBytes:   []byte{},
		privKey:        nil,
		certChainBytes: []byte{},
		rootCertBytes:  rootCertBytes,
	}, nil
}

// GetAllPem returns all key/cert PEMs in KeyCertBundle together. Getting all values together avoids inconsistency.
func (b *KeyCertBundleImpl) GetAllPem() (certBytes, privKeyBytes, certChainBytes, rootCertBytes []byte) {
	b.mutex.RLock()
	certBytes = copyBytes(b.certBytes)
	privKeyBytes = copyBytes(b.privKeyBytes)
	certChainBytes = copyBytes(b.certChainBytes)
	rootCertBytes = copyBytes(b.rootCertBytes)
	b.mutex.RUnlock()
	return
}

// GetAll returns all key/cert in KeyCertBundle together. Getting all values together avoids inconsistency.
// NOTE: Callers should not modify the content of cert and privKey.
func (b *KeyCertBundleImpl) GetAll() (cert *x509.Certificate, privKey *crypto.PrivateKey, certChainBytes,
	rootCertBytes []byte) {
	b.mutex.RLock()
	cert = b.cert
	privKey = b.privKey
	certChainBytes = copyBytes(b.certChainBytes)
	rootCertBytes = copyBytes(b.rootCertBytes)
	b.mutex.RUnlock()
	return
}

// VerifyAndSetAll verifies the key/certs, and sets all key/certs in KeyCertBundle together.
// Setting all values together avoids inconsistency.
func (b *KeyCertBundleImpl) VerifyAndSetAll(certBytes, privKeyBytes, certChainBytes, rootCertBytes []byte) error {
	if err := verify(certBytes, privKeyBytes, certChainBytes, rootCertBytes); err != nil {
		return err
	}
	b.mutex.Lock()
	b.certBytes = copyBytes(certBytes)
	b.privKeyBytes = copyBytes(privKeyBytes)
	b.certChainBytes = copyBytes(certChainBytes)
	b.rootCertBytes = copyBytes(rootCertBytes)
	// cert and privKey are always reset to point to new addresses. This avoids modifying the pointed structs that
	// could be still used outside of the class.
	b.cert, _ = ParsePemEncodedCertificate(certBytes)
	privKey, _ := ParsePemEncodedKey(privKeyBytes)
	b.privKey = &privKey
	b.mutex.Unlock()
	return nil
}

// verify that the cert chain, root cert and key/cert match.
func verify(certBytes, privKeyBytes, certChainBytes, rootCertBytes []byte) error {
	// Verify the cert can be verified from the root cert through the cert chain.
	rcp := x509.NewCertPool()
	rcp.AppendCertsFromPEM(rootCertBytes)

	icp := x509.NewCertPool()
	icp.AppendCertsFromPEM(certChainBytes)

	opts := x509.VerifyOptions{
		Intermediates: icp,
		Roots:         rcp,
	}
	cert, err := ParsePemEncodedCertificate(certBytes)
	if err != nil {
		return fmt.Errorf("failed to parse cert PEM: %v", err)
	}
	chains, err := cert.Verify(opts)

	if len(chains) == 0 || err != nil {
		return fmt.Errorf(
			"cannot verify the cert with the provided root chain and cert pool")
	}

	// Verify that the key can be correctly parsed.
	if _, err = ParsePemEncodedKey(privKeyBytes); err != nil {
		return fmt.Errorf("failed to parse private key PEM: %v", err)
	}

	// Verify the cert and key match.
	if _, err := tls.X509KeyPair(certBytes, privKeyBytes); err != nil {
		return fmt.Errorf("the cert does not match the key")
	}

	return nil
}

func copyBytes(src []byte) []byte {
	bs := make([]byte, len(src))
	copy(bs, src)
	return bs
}
