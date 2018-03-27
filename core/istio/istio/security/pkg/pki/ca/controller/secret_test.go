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

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"istio.io/istio/security/pkg/pki/util"
)

const (
	defaultTTL              = time.Hour
	defaultGracePeriodRatio = 0.5
	defaultMinGracePeriod   = 10 * time.Minute
)

type fakeCa struct{}

func (ca *fakeCa) Sign([]byte, time.Duration, bool) ([]byte, error) {
	return []byte("fake cert chain"), nil
}

func (ca *fakeCa) GetRootCertificate() []byte {
	return []byte("fake root cert")
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
	}

	for k, tc := range testCases {
		client := fake.NewSimpleClientset()
		controller, err := NewSecretController(&fakeCa{}, defaultTTL, defaultGracePeriodRatio, defaultMinGracePeriod,
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

func TestRecoverFromDeletedIstioSecret(t *testing.T) {
	client := fake.NewSimpleClientset()
	controller, err := NewSecretController(&fakeCa{}, defaultTTL, defaultGracePeriodRatio, defaultMinGracePeriod,
		client.CoreV1(), metav1.NamespaceAll)
	if err != nil {
		t.Errorf("failed to create secret controller: %v", err)
	}
	scrt := createSecret("test", "istio.test", "test-ns")
	controller.scrtDeleted(scrt)

	gvr := schema.GroupVersionResource{
		Resource: "secrets",
		Version:  "v1",
	}
	expectedActions := []ktesting.Action{ktesting.NewCreateAction(gvr, "test-ns", scrt)}
	if err := checkActions(client.Actions(), expectedActions); err != nil {
		t.Error(err)
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
			gracePeriodRatio: defaultGracePeriodRatio,
			minGracePeriod:   defaultMinGracePeriod,
		},
		"Update secret in grace period": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              defaultTTL,
			gracePeriodRatio: 1, // Always in grace period
			minGracePeriod:   defaultMinGracePeriod,
		},
		"Update secret in min grace period": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              10 * time.Minute,
			gracePeriodRatio: defaultGracePeriodRatio,
			minGracePeriod:   time.Hour, // ttl is always in minGracePeriod
		},
		"Update expired secret": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              -time.Second,
			gracePeriodRatio: defaultGracePeriodRatio,
			minGracePeriod:   defaultMinGracePeriod,
		},
		"Update secret with different root cert": {
			expectedActions: []ktesting.Action{
				ktesting.NewUpdateAction(gvr, "test-ns", createSecret("test", "istio.test", "test-ns")),
			},
			ttl:              defaultTTL,
			gracePeriodRatio: defaultGracePeriodRatio,
			minGracePeriod:   defaultMinGracePeriod,
			rootCert:         []byte("Outdated root cert"),
		},
	}

	for k, tc := range testCases {
		client := fake.NewSimpleClientset()
		controller, err := NewSecretController(&fakeCa{}, time.Hour, tc.gracePeriodRatio, tc.minGracePeriod,
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
