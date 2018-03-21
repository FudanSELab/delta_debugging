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

package admit

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/kube/admit/testcerts"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/model/test"
	"istio.io/istio/pilot/pkg/serviceregistry/kube"
	"istio.io/istio/pilot/test/mock"
	"istio.io/istio/tests/k8s"
)

const (
	watchedNamespace    = "watched"
	nonWatchedNamespace = "not-watched"
)

const (
	// testAdmissionHookName is the default name for the
	// ExternalAdmissionHooks.
	testAdmissionHookName = "pilot.config.istio.io"

	// testAdmissionServiceName is the default service of the
	// validation webhook.
	testAdmissionServiceName = "istio-pilot"

	// testDomainSuffix is the default DNS domain suffix for Istio
	// CRD resources.
	testDomainSuffix = "local.cluster"
)

func makeConfig(t *testing.T, namespace string, i int, valid bool) []byte {
	t.Helper()

	var key string
	if valid {
		key = "key"
	}
	name := fmt.Sprintf("%s%d", "mock-config", i)
	config := model.Config{
		ConfigMeta: model.ConfigMeta{
			Type:      model.MockConfig.Type,
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"key": name,
			},
			Annotations: map[string]string{
				"annotationkey": name,
			},
		},
		Spec: &test.MockConfig{
			Key: key,
			Pairs: []*test.ConfigPair{{
				Key:   key,
				Value: strconv.Itoa(i),
			}},
		},
	}
	obj, err := crd.ConvertConfig(model.MockConfig, config)
	if err != nil {
		t.Fatalf("ConvertConfig(%v) failed: %v", config.Name, err)
	}
	raw, err := json.Marshal(&obj)
	if err != nil {
		t.Fatalf("Marshal(%v) failed: %v", config.Name, err)
	}
	return raw
}

func TestAdmissionController(t *testing.T) {
	valid := makeConfig(t, watchedNamespace, 0, true)
	invalid := makeConfig(t, watchedNamespace, 0, false)
	nonWatchedInvalid := makeConfig(t, nonWatchedNamespace, 0, false)

	cases := []struct {
		name            string
		in              *admissionv1beta1.AdmissionRequest
		want            *admissionv1beta1.AdmissionResponse
		useNamespaceAll bool
	}{
		{
			name: "valid create",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{}, // TODO
				Object: runtime.RawExtension{
					Raw: valid,
				},
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
		},
		{
			name: "valid update",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{}, // TODO
				Object: runtime.RawExtension{
					Raw: valid,
				},
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
		},
		{
			name: "invalid raw content",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{},
				Object: runtime.RawExtension{
					Raw: invalid,
				},
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: false,
			},
		},
		{
			name: "skip invalid raw content in non-watched namespace",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{},
				Object: runtime.RawExtension{
					Raw: nonWatchedInvalid,
				},
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
		},
		{
			name: "valid in NamespaceAll",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{},
				Object: runtime.RawExtension{
					Raw: nonWatchedInvalid,
				},
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
		},
		{
			name: "invalid in NamespaceAll",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{},
				Object: runtime.RawExtension{
					Raw: nonWatchedInvalid,
				},
				Operation: admissionv1beta1.Create,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: false,
			},
			useNamespaceAll: true,
		},
		{
			name: "don't reject delete of invalid",
			in: &admissionv1beta1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{},
				Object: runtime.RawExtension{
					Raw: nonWatchedInvalid,
				},
				Operation: admissionv1beta1.Delete,
			},
			want: &admissionv1beta1.AdmissionResponse{
				Allowed: true,
			},
			useNamespaceAll: true,
		},
	}

	for _, c := range cases {
		namespaces := []string{watchedNamespace}
		if c.useNamespaceAll {
			namespaces = []string{metav1.NamespaceAll}
		}
		testAdmissionController, err := NewController(nil, ControllerOptions{
			Descriptor:                   mock.Types,
			ExternalAdmissionWebhookName: testAdmissionHookName,
			ServiceName:                  testAdmissionServiceName,
			ServiceNamespace:             "istio-system",
			ValidateNamespaces:           namespaces,
			DomainSuffix:                 testDomainSuffix,
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

func makeTestData(t *testing.T, valid bool) []byte {
	t.Helper()
	review := admissionv1beta1.AdmissionReview{
		Request: &admissionv1beta1.AdmissionRequest{
			Kind: metav1.GroupVersionKind{},
			Object: runtime.RawExtension{
				Raw: makeConfig(t, watchedNamespace, 0, valid),
			},
			Operation: admissionv1beta1.Create,
		},
	}
	reviewJSON, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("Failed to create AdmissionReview: %v", err)
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
	validReview := makeTestData(t, true)
	invalidReview := makeTestData(t, false)

	cases := []struct {
		name           string
		body           []byte
		contentType    string
		wantAllowed    bool
		wantStatusCode int
	}{
		{
			name:           "valid",
			body:           validReview,
			contentType:    "application/json",
			wantAllowed:    true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid",
			body:           invalidReview,
			contentType:    "application/json",
			wantAllowed:    false,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "wrong content-type",
			body:           validReview,
			contentType:    "application/yaml",
			wantAllowed:    false,
			wantStatusCode: http.StatusUnsupportedMediaType,
		},
		{
			name:           "bad content",
			body:           []byte{0, 1, 2, 3, 4, 5}, // random data
			contentType:    "application/json",
			wantAllowed:    false,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	testAdmissionController, err := NewController(nil, ControllerOptions{
		Descriptor:                   mock.Types,
		ExternalAdmissionWebhookName: testAdmissionHookName,
		ServiceName:                  testAdmissionServiceName,
		ServiceNamespace:             "istio-system",
		ValidateNamespaces:           []string{watchedNamespace},
		DomainSuffix:                 testDomainSuffix,
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
			t.Errorf("%v: AdmissionReview.Response.Allowed is wrong : got %v want %v",
				c.name, gotReview.Response.Allowed, c.wantAllowed)
		}
		_ = res.Body.Close()
	}
}

func TestRegister(t *testing.T) {
	testAdmissionController, err := NewController(nil, ControllerOptions{
		Descriptor:                   mock.Types,
		ExternalAdmissionWebhookName: testAdmissionHookName,
		ServiceName:                  testAdmissionServiceName,
		ServiceNamespace:             "istio-system",
		ValidateNamespaces:           []string{watchedNamespace},
		DomainSuffix:                 testDomainSuffix,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	fakeClient := fake.NewSimpleClientset(&admissionregistrationv1beta1.ValidatingWebhookConfiguration{})

	fakeAdmissionClient := fakeClient.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations()
	if err := testAdmissionController.register(fakeAdmissionClient, testcerts.CACert); err != nil {
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
	// Not clear how it worked before - probably some fake-hermetic script copying or
	// mapping the real users' config
	kubeconfig := k8s.Kubeconfig("/../config")
	_, cl, err := kube.CreateInterface(kubeconfig)
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
