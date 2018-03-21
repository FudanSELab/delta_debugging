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

package crd

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"

	"istio.io/istio/mixer/pkg/config/store"
	"istio.io/istio/pilot/pkg/kube/admit/testcerts"
	"istio.io/istio/tests/k8s"
)

const (
	watchedNamespace    = "watched"
	nonWatchedNamespace = "not-watched"
)

const (
	// testAdmissionHookName is the default name for the
	// ExternalAdmissionHooks.
	testAdmissionHookName = "mixer.config.istio.io"

	// testAdmissionServiceName is the default service of the
	// validation webhook.
	testAdmissionServiceName = "istio-mixer"
)

type fakeValidator struct {
	err error
}

func (fv *fakeValidator) Validate(ev *store.BackendEvent) error {
	return fv.err
}

func makeConfig(t *testing.T, namespace string, i int) []byte {
	uns := &unstructured.Unstructured{}
	name := fmt.Sprintf("%s%d", "mock-config", i)
	uns.SetName(name)
	uns.SetNamespace(namespace)
	uns.SetKind("mock")
	uns.Object["spec"] = map[string]interface{}{"foo": 1}
	raw, err := json.Marshal(uns)
	if err != nil {
		t.Fatalf("Marshal(%v) failed: %v", uns, err)
	}
	return raw
}

func TestAdmissionController(t *testing.T) {
	valid := makeConfig(t, watchedNamespace, 0)
	nonWatchedInvalid := makeConfig(t, nonWatchedNamespace, 0)

	cases := []struct {
		name            string
		in              *admissionv1beta1.AdmissionRequest
		want            *admissionv1beta1.AdmissionResponse
		useNamespaceAll bool
		v               store.BackendValidator
	}{
		{
			name: "valid create",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Object: runtime.RawExtension{
					Raw: valid,
				},
				Namespace: watchedNamespace,
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
			useNamespaceAll: true,
			v:               &fakeValidator{},
		},
		{
			name: "valid update",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Object: runtime.RawExtension{
					Raw: valid,
				},
				Namespace: watchedNamespace,
				Operation: admissionv1beta1.Update,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
			useNamespaceAll: true,
			v:               &fakeValidator{},
		},
		{
			name: "valid delete",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Namespace: watchedNamespace,
				Name:      "to-be-deleted",
				Operation: admissionv1beta1.Delete,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
			useNamespaceAll: true,
			v:               &fakeValidator{},
		},
		{
			name: "validation failure",
			in: &admissionv1beta1.AdmissionRequest{

				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Object: runtime.RawExtension{
					Raw: valid,
				},
				Namespace: watchedNamespace,
				Operation: admissionv1beta1.Update,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: false,
			},
			useNamespaceAll: true,
			v:               &fakeValidator{errors.New("fail")},
		},
		{
			name: "validation failure on delete",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Namespace: watchedNamespace,
				Name:      "to-be-deleted",
				Operation: admissionv1beta1.Delete,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: false,
			},
			useNamespaceAll: true,
			v:               &fakeValidator{errors.New("fail")},
		},
		{
			name: "valid in NamespaceAll",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Object: runtime.RawExtension{
					Raw: nonWatchedInvalid,
				},
				Namespace: nonWatchedNamespace,
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
			useNamespaceAll: false,
			v:               &fakeValidator{errors.New("fail")},
		},
		{
			name: "invalid in NamespaceAll",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Object: runtime.RawExtension{
					Raw: nonWatchedInvalid,
				},
				Namespace: nonWatchedNamespace,
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: false,
			},
			useNamespaceAll: true,
			v:               &fakeValidator{errors.New("fail")},
		},
		{
			name: "invalid create",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Namespace: watchedNamespace,
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: false,
			},
		},
		{
			name: "invalid delete",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Kind: "mock",
				},
				Namespace: watchedNamespace,
				Operation: admissionv1beta1.Delete,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: false,
			},
		},
	}

	for _, c := range cases {
		namespaces := []string{watchedNamespace}
		if c.useNamespaceAll {
			namespaces = []string{metav1.NamespaceAll}
		}
		testAdmissionController, err := NewController(nil, ControllerOptions{
			ExternalAdmissionWebhookName: testAdmissionHookName,
			ServiceName:                  testAdmissionServiceName,
			ServiceNamespace:             "istio-system",
			ValidateNamespaces:           namespaces,
			Validator:                    c.v,
		})
		if err != nil {
			t.Fatal(err.Error())
		}

		got := testAdmissionController.admit(c.in)
		if got.Allowed != c.want.Allowed {
			t.Errorf("%v: AdmissionResponse.Allowed is wrong : got %v want %v",
				c.name, got.Allowed, c.want.Allowed)
		}
	}
}

func makeTestData(t *testing.T) []byte {
	review := admissionv1beta1.AdmissionReview{
		Request: &admissionv1beta1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{
				Kind: "mock",
			},
			Object: runtime.RawExtension{
				Raw: makeConfig(t, watchedNamespace, 0),
			},
			Namespace: watchedNamespace,
			Operation: admissionv1beta1.Create,
		},
	}
	reviewJSON, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("Failed to create AdmissionRequest: %v", err)
	}
	return reviewJSON
}

func makeTestClient() (*http.Client, error) {
	clientCert, err := tls.X509KeyPair(testcerts.ClientCert, testcerts.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509KeyPair: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(testcerts.CACert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 5 * time.Second,
	}, nil
}

func makeTestServer(handler http.Handler, tlsConfig *tls.Config) (*http.Server, net.Listener, error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, nil, fmt.Errorf("net.Listen failed: %v", err)
	}
	server := &http.Server{
		Handler:   handler,
		TLSConfig: tlsConfig,
	}
	return server, tls.NewListener(ln, tlsConfig), nil
}

func TestServe(t *testing.T) {
	review := makeTestData(t)

	cases := []struct {
		name           string
		body           []byte
		contentType    string
		validator      store.BackendValidator
		wantAllowed    bool
		wantStatusCode int
	}{
		{
			name:           "valid",
			body:           review,
			contentType:    "application/json",
			validator:      &fakeValidator{},
			wantAllowed:    true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid",
			body:           review,
			contentType:    "application/json",
			validator:      &fakeValidator{errors.New("fake error")},
			wantAllowed:    false,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "wrong content-type",
			body:           review,
			contentType:    "application/yaml",
			validator:      &fakeValidator{},
			wantAllowed:    false,
			wantStatusCode: http.StatusUnsupportedMediaType,
		},
		{
			name:           "bad content",
			body:           []byte{0, 1, 2, 3, 4, 5}, // random data
			contentType:    "application/json",
			validator:      &fakeValidator{},
			wantAllowed:    false,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	testAdmissionController, err := NewController(nil, ControllerOptions{
		ExternalAdmissionWebhookName: testAdmissionHookName,
		ServiceName:                  testAdmissionServiceName,
		ServiceNamespace:             "istio-system",
		ValidateNamespaces:           []string{watchedNamespace},
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	tlsConfig, err := makeTLSConfig(testcerts.ServerCert, testcerts.ServerKey, testcerts.CACert)
	if err != nil {
		t.Fatalf("MakeTLSConfig failed: %v", err)
	}

	testServer, testListener, err := makeTestServer(testAdmissionController, tlsConfig)
	if err != nil {
		t.Fatalf("Could not create test server: %v", err)
	}
	go func() {
		if serverErr := testServer.Serve(testListener); serverErr != nil {
			t.Log(serverErr.Error())
		}
	}()
	defer testServer.Close() // nolint: errcheck

	testURL := fmt.Sprintf("https://%v", testListener.Addr().String())

	testClient, err := makeTestClient()
	if err != nil {
		t.Fatalf("Could not create test client: %v", err)
	}

	for _, c := range cases {
		testAdmissionController.options.Validator = c.validator
		res, err := testClient.Post(testURL, c.contentType, bytes.NewReader(c.body))
		if err != nil {
			t.Errorf("%v: Post(%v, %v) failed %v", c.name, c.contentType, string(c.body), err)
			continue
		}

		if res.StatusCode != c.wantStatusCode {
			t.Errorf("%v: wrong status code: \ngot %v \nwant %v", c.name, res.StatusCode, c.wantStatusCode)
		}

		if res.StatusCode != http.StatusOK {
			continue
		}

		gotBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("%v: could not read body: %v", c.name, err)
			continue
		}
		var gotReview admissionv1beta1.AdmissionReview
		if err := json.Unmarshal(gotBody, &gotReview); err != nil {
			t.Errorf("%v: could not decode response body: %v", c.name, err)
		}
		if gotReview.Response.Allowed != c.wantAllowed {
			t.Errorf("%v: AdmissionRequest.Status.Allowed is wrong : got %v want %v",
				c.name, gotReview.Response.Allowed, c.wantAllowed)
		}
		_ = res.Body.Close()
	}
}

func TestRegister(t *testing.T) {
	testAdmissionController, err := NewController(nil, ControllerOptions{
		ExternalAdmissionWebhookName: testAdmissionHookName,
		ServiceName:                  testAdmissionServiceName,
		ServiceNamespace:             "istio-system",
		ValidateNamespaces:           []string{watchedNamespace},
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	fakeClient := fake.NewSimpleClientset(&admissionregistrationv1beta1.ValidatingWebhookConfiguration{})

	fakeAdmissionClient := fakeClient.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations()
	if err := testAdmissionController.register(fakeAdmissionClient, []byte("fake cert")); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	wantVerbs := []string{"delete", "create"}
	actions := fakeClient.Actions()
	if len(actions) != len(wantVerbs) {
		t.Fatalf("register: unexpected number of actions: got %v want %v number actions", len(actions), len(wantVerbs))
	}
	for i, verb := range wantVerbs {
		if actions[i].GetResource().Resource != "validatingwebhookconfigurations" {
			t.Errorf("register: unexpected action: got %v want %v",
				actions[i].GetResource().Resource, "validatingwebhookconfigurations")
		}
		if actions[i].GetVerb() != verb {
			t.Errorf("register: unexpected action: got %v want %v", actions[i], verb)
		}
	}
	fakeClient.ClearActions()

	if err := testAdmissionController.unregister(fakeAdmissionClient); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	wantVerbs = []string{"delete"}
	actions = fakeClient.Actions()
	if len(actions) != len(wantVerbs) {
		t.Fatalf("unregister: unexpected number of actions: got %v want %v number actions", len(actions), len(wantVerbs))
	}
	for i, verb := range wantVerbs {
		if actions[i].GetResource().Resource != "validatingwebhookconfigurations" {
			t.Errorf("unregister: unexpected action: got %v want %v",
				actions[i].GetResource().Resource, "validatingwebhookconfigurations")
		}
		if actions[i].GetVerb() != verb {
			t.Errorf("unregister: unexpected action: got %v want %v", actions[i], verb)
		}
	}
	fakeClient.ClearActions()
}

func makeClient(t *testing.T) kubernetes.Interface {
	kubeconfig := k8s.Kubeconfig("/../../../../pilot/pkg/kube/config")
	conf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	cl, err := kubernetes.NewForConfig(conf)
	if err != nil {
		t.Fatal(err)
	}
	return cl
}

func TestGetAPIServerExtensionCACert(t *testing.T) {
	cl := makeClient(t)
	if _, err := getAPIServerExtensionCACert(cl); err != nil {
		t.Errorf("GetAPIServerExtensionCACert() failed: %v", err)
	}
}
