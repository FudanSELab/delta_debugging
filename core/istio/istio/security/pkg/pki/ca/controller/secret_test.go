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

package controller

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	mockca "istio.io/istio/security/pkg/pki/ca/mock"
	"istio.io/istio/security/pkg/pki/util"
	mockutil "istio.io/istio/security/pkg/pki/util/mock"
)

const (
	defaultTTL              = time.Hour
	defaultGracePeriodRatio = 0.5
	defaultMinGracePeriod   = 10 * time.Minute
)

func TestSecretController(t *testing.T) {
	gvr := schema.GroupVersionResource{
		Resource: "secrets",
		Version:  "v1",
	}
	testCases := map[string]struct {
		existingSecret  *v1.Secret
		saToAdd         *v1.ServiceAccount
		saToDelete      *v1.ServiceAccount
		sasToUpdate     *updatedSas
		expectedActions []ktesting.Action
		injectFailure   bool
	}{
		"adding service account creates new secret": {
			saToAdd: createServiceAccount("test", "test-ns"),
			expectedActions: []ktesting.Action{
				ktesting.NewCreateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
		},
		"removing service account deletes existing secret": {
			saToDelete: createServiceAccount("deleted", "deleted-ns"),
			expectedActions: []ktesting.Action{
				ktesting.NewDeleteAction(gvr, "deleted-ns", "istio.deleted"),
			},
		},
		"updating service accounts does nothing if name and namespace are not changed": {
			sasToUpdate: &updatedSas{
				curSa: createServiceAccount("name", "ns"),
				oldSa: createServiceAccount("name", "ns"),
			},
			expectedActions: []ktesting.Action{},
		},
		"updating service accounts deletes old secret and creates a new one": {
			sasToUpdate: &updatedSas{
				curSa: createServiceAccount("new-name", "new-ns"),
				oldSa: createServiceAccount("old-name", "old-ns"),
			},
			expectedActions: []ktesting.Action{
				ktesting.NewDeleteAction(gvr, "old-ns", "istio.old-name"),
				ktesting.NewCreateAction(gvr, "new-ns", createSecret("new-name", "istio.new-name", "new-ns")),
			},
		},
		"adding new service account does not overwrite existing secret": {
			existingSecret:  createSecret("test", "istio.test", "test-ns"),
			saToAdd:         createServiceAccount("test", "test-ns"),
			expectedActions: []ktesting.Action{},
		},
		"adding service account retries when failed": {
			saToAdd: createServiceAccount("test", "test-ns"),
			expectedActions: []ktesting.Action{
				ktesting.NewCreateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
				ktesting.NewCreateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
				ktesting.NewCreateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			injectFailure: true,
		},
	}

	for k, tc := range testCases {
		client := fake.NewSimpleClientset()

		if tc.injectFailure {
			callCount := 0
			// PrependReactor to ensure action handled by our handler.
			client.Fake.PrependReactor("*", "*", func(a ktesting.Action) (bool, runtime.Object, error) {
				callCount++
				if callCount < secretCreationRetry {
					return true, nil, errors.New("failed to create secret deliberately")
				}
				return true, nil, nil
			})
		}

		controller, err := NewSecretController(createFakeCA(), defaultTTL, defaultGracePeriodRatio, defaultMinGracePeriod,
			client.CoreV1(), metav1.NamespaceAll)
		if err != nil {
			t.Errorf("failed to create secret controller: %v", err)
		}

		if tc.existingSecret != nil {
			err := controller.scrtStore.Add(tc.existingSecret)
			if err != nil {
				t.Errorf("Failed to add a secret (error %v)", err)
			}
		}

		if tc.saToAdd != nil {
			controller.saAdded(tc.saToAdd)
		}
		if tc.saToDelete != nil {
			controller.saDeleted(tc.saToDelete)
		}
		if tc.sasToUpdate != nil {
			controller.saUpdated(tc.sasToUpdate.oldSa, tc.sasToUpdate.curSa)
		}

		if err := checkActions(client.Actions(), tc.expectedActions); err != nil {
			t.Errorf("Case %q: %s", k, err.Error())
		}
	}
}

func TestDeletedIstioSecret(t *testing.T) {
	client := fake.NewSimpleClientset()
	controller, err := NewSecretController(createFakeCA(), defaultTTL, defaultGracePeriodRatio, defaultMinGracePeriod,
		client.CoreV1(), metav1.NamespaceAll)
	if err != nil {
		t.Errorf("failed to create secret controller: %v", err)
	}
	sa := createServiceAccount("test-sa", "test-ns")
	if _, err := client.CoreV1().ServiceAccounts("test-ns").Create(sa); err != nil {
		t.Error(err)
	}

	saGvr := schema.GroupVersionResource{
		Resource: "serviceaccounts",
		Version:  "v1",
	}
	scrtGvr := schema.GroupVersionResource{
		Resource: "secrets",
		Version:  "v1",
	}

	testCases := map[string]struct {
		secret          *v1.Secret
		expectedActions []ktesting.Action
	}{
		"Recover secret for existing service account": {
			secret: createSecret("test-sa", "istio.test-sa", "test-ns"),
			expectedActions: []ktesting.Action{
				ktesting.NewGetAction(saGvr, "test-ns", "test-sa"),
				ktesting.NewCreateAction(scrtGvr, "test-ns", createSecret("test-sa", "istio.test-sa", "test-ns")),
			},
		},
		"Do not recover secret for non-existing service account in the same namespace": {
			secret: createSecret("test-sa2", "istio.test-sa2", "test-ns"),
			expectedActions: []ktesting.Action{
				ktesting.NewGetAction(saGvr, "test-ns", "test-sa2"),
			},
		},
		"Do not recover secret for service account in different namespace": {
			secret: createSecret("test-sa", "istio.test-sa", "test-ns2"),
			expectedActions: []ktesting.Action{
				ktesting.NewGetAction(saGvr, "test-ns2", "test-sa"),
			},
		},
	}

	for k, tc := range testCases {
		client.ClearActions()
		controller.scrtDeleted(tc.secret)
		if err := checkActions(client.Actions(), tc.expectedActions); err != nil {
			t.Errorf("Failure in test case %s: %v", k, err)
		}
	}
}

func TestUpdateSecret(t *testing.T) {
	gvr := schema.GroupVersionResource{
		Resource: "secrets",
		Version:  "v1",
	}
	testCases := map[string]struct {
		expectedActions  []ktesting.Action
		ttl              time.Duration
		gracePeriodRatio float32
		minGracePeriod   time.Duration
		rootCert         []byte
	}{
		"Does not update non-expiring secret": {
			expectedActions:  []ktesting.Action{},
			ttl:              time.Hour,
			gracePeriodRatio: 0.5,
			minGracePeriod:   10 * time.Minute,
		},
		"Update secret in grace period": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              time.Hour,
			gracePeriodRatio: 1, // Always in grace period
			minGracePeriod:   10 * time.Minute,
		},
		"Update secret in min grace period": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              10 * time.Minute,
			gracePeriodRatio: 0.5,
			minGracePeriod:   time.Hour, // ttl is always in minGracePeriod
		},
		"Update expired secret": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              -time.Second,
			gracePeriodRatio: 0.5,
			minGracePeriod:   10 * time.Minute,
		},
		"Update secret with different root cert": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              time.Hour,
			gracePeriodRatio: 0.5,
			minGracePeriod:   10 * time.Minute,
			rootCert:         []byte("Outdated root cert"),
		},
	}

	for k, tc := range testCases {
		client := fake.NewSimpleClientset()
		controller, err := NewSecretController(createFakeCA(), time.Hour, tc.gracePeriodRatio, tc.minGracePeriod,
			client.CoreV1(), metav1.NamespaceAll)
		if err != nil {
			t.Errorf("failed to create secret controller: %v", err)
		}

		scrt := createSecret("test", "istio.test", "test-ns")
		if rc := tc.rootCert; rc != nil {
			scrt.Data[RootCertID] = rc
		}

		opts := util.CertOptions{
			IsSelfSigned: true,
			TTL:          tc.ttl,
			RSAKeySize:   512,
		}
		bs, _, err := util.GenCertKeyFromOptions(opts)
		if err != nil {
			t.Error(err)
		}
		scrt.Data[CertChainID] = bs

		controller.scrtUpdated(nil, scrt)

		if err := checkActions(client.Actions(), tc.expectedActions); err != nil {
			t.Errorf("Case %q: %s", k, err.Error())
		}
	}
}

func checkActions(actual, expected []ktesting.Action) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("unexpected number of actions, want %d but got %d", len(expected), len(actual))
	}

	for i, action := range actual {
		expectedAction := expected[i]
		verb := expectedAction.GetVerb()
		resource := expectedAction.GetResource().Resource
		if !action.Matches(verb, resource) {
			return fmt.Errorf("unexpected %dth action, want %q but got %q", i, expectedAction, action)
		}
	}

	return nil
}

func createFakeCA() *mockca.FakeCA {
	return &mockca.FakeCA{
		SignedCert: []byte("fake signed cert"),
		SignErr:    nil,
		KeyCertBundle: &mockutil.FakeKeyCertBundle{
			CertBytes:      []byte("fake CA cert"),
			PrivKeyBytes:   []byte("fake private key"),
			CertChainBytes: []byte("fake cert chain"),
			RootCertBytes:  []byte("fake root cert"),
		},
	}
}

func createSecret(saName, scrtName, namespace string) *v1.Secret {
	return &v1.Secret{
		Data: map[string][]byte{
			CertChainID:  []byte("fake cert chain"),
			PrivateKeyID: []byte("fake key"),
			RootCertID:   []byte("fake root cert"),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"istio.io/service-account.name": saName},
			Name:        scrtName,
			Namespace:   namespace,
		},
		Type: IstioSecretType,
	}
}

func createServiceAccount(name, namespace string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

type updatedSas struct {
	curSa *v1.ServiceAccount
	oldSa *v1.ServiceAccount
}
