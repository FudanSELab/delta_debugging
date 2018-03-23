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
	"bytes"
	"fmt"
	"reflect"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"

	"istio.io/istio/pkg/log"
	"istio.io/istio/security/pkg/pki/ca"
	"istio.io/istio/security/pkg/pki/util"
)

/* #nosec: disable gas linter */
const (
	// The Istio secret annotation type
	IstioSecretType = "istio.io/key-and-cert"

	// The ID/name for the certificate chain file.
	CertChainID = "cert-chain.pem"
	// The ID/name for the private key file.
	PrivateKeyID = "key.pem"
	// The ID/name for the CA root certificate file.
	RootCertID = "root-cert.pem"

	secretNamePrefix   = "istio."
	secretResyncPeriod = time.Minute

	serviceAccountNameAnnotationKey = "istio.io/service-account.name"

	recommendedMinGracePeriodRatio = 0.3
	recommendedMaxGracePeriodRatio = 0.8

	// The size of a private key for a leaf certificate.
	keySize = 2048
)

// SecretController manages the service accounts' secrets that contains Istio keys and certificates.
type SecretController struct {
	ca      ca.CertificateAuthority
	certTTL time.Duration
	core    corev1.CoreV1Interface
	// Length of the grace period for the certificate rotation.
	gracePeriodRatio float32
	minGracePeriod   time.Duration

	// Controller and store for service account objects.
	saController cache.Controller
	saStore      cache.Store

	// Controller and store for secret objects.
	scrtController cache.Controller
	scrtStore      cache.Store
}

// NewSecretController returns a pointer to a newly constructed SecretController instance.
func NewSecretController(ca ca.CertificateAuthority, certTTL time.Duration, gracePeriodRatio float32, minGracePeriod time.Duration,
	core corev1.CoreV1Interface, namespace string) (*SecretController, error) {

	if gracePeriodRatio < 0 || gracePeriodRatio > 1 {
		return nil, fmt.Errorf("grace period ratio %f should be within [0, 1]", gracePeriodRatio)
	}
	if gracePeriodRatio < recommendedMinGracePeriodRatio || gracePeriodRatio > recommendedMaxGracePeriodRatio {
		log.Warnf("grace period ratio %f is out of the recommended window [%.2f, %.2f]",
			gracePeriodRatio, recommendedMinGracePeriodRatio, recommendedMaxGracePeriodRatio)
	}

	c := &SecretController{
		ca:               ca,
		certTTL:          certTTL,
		gracePeriodRatio: gracePeriodRatio,
		minGracePeriod:   minGracePeriod,
		core:             core,
	}

	saLW := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return core.ServiceAccounts(namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return core.ServiceAccounts(namespace).Watch(options)
		},
	}
	rehf := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.saAdded,
		DeleteFunc: c.saDeleted,
		UpdateFunc: c.saUpdated,
	}
	c.saStore, c.saController = cache.NewInformer(saLW, &v1.ServiceAccount{}, time.Minute, rehf)

	istioSecretSelector := fields.SelectorFromSet(map[string]string{"type": IstioSecretType}).String()
	scrtLW := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = istioSecretSelector
			return core.Secrets(namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = istioSecretSelector
			return core.Secrets(namespace).Watch(options)
		},
	}
	c.scrtStore, c.scrtController =
		cache.NewInformer(scrtLW, &v1.Secret{}, secretResyncPeriod, cache.ResourceEventHandlerFuncs{
			DeleteFunc: c.scrtDeleted,
			UpdateFunc: c.scrtUpdated,
		})

	return c, nil
}

// Run starts the SecretController until a value is sent to stopCh.
func (sc *SecretController) Run(stopCh chan struct{}) {
	go sc.scrtController.Run(stopCh)
	go sc.saController.Run(stopCh)
}

// Handles the event where a service account is added.
func (sc *SecretController) saAdded(obj interface{}) {
	acct := obj.(*v1.ServiceAccount)
	sc.upsertSecret(acct.GetName(), acct.GetNamespace())
}

// Handles the event where a service account is deleted.
func (sc *SecretController) saDeleted(obj interface{}) {
	acct := obj.(*v1.ServiceAccount)
	sc.deleteSecret(acct.GetName(), acct.GetNamespace())
}

// Handles the event where a service account is updated.
func (sc *SecretController) saUpdated(oldObj, curObj interface{}) {
	if reflect.DeepEqual(oldObj, curObj) {
		// Nothing is changed. The method is invoked by periodical re-sync with the apiserver.
		return
	}
	oldSa := oldObj.(*v1.ServiceAccount)
	curSa := curObj.(*v1.ServiceAccount)

	curName := curSa.GetName()
	curNamespace := curSa.GetNamespace()
	oldName := oldSa.GetName()
	oldNamespace := oldSa.GetNamespace()

	// We only care the name and namespace of a service account.
	if curName != oldName || curNamespace != oldNamespace {
		sc.deleteSecret(oldName, oldNamespace)
		sc.upsertSecret(curName, curNamespace)

		log.Infof("Service account \"%s\" in namespace \"%s\" has been updated to \"%s\" in namespace \"%s\"",
			oldName, oldNamespace, curName, curNamespace)
	}
}

func (sc *SecretController) upsertSecret(saName, saNamespace string) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{serviceAccountNameAnnotationKey: saName},
			Name:        getSecretName(saName),
			Namespace:   saNamespace,
		},
		Type: IstioSecretType,
	}

	_, exists, err := sc.scrtStore.Get(secret)
	if err != nil {
		log.Errorf("Failed to get secret from the store (error %v)", err)
	}

	if exists {
		// Do nothing for existing secrets. Rotating expiring certs are handled by the `scrtUpdated` method.
		return
	}

	// Now we know the secret does not exist yet. So we create a new one.
	chain, key, err := sc.generateKeyAndCert(saName, saNamespace)
	if err != nil {
		log.Errorf("Failed to generate key and certificate for service account %q in namespace %q (error %v)",
			saName, saNamespace, err)

		return
	}
	rootCert := sc.ca.GetRootCertificate()
	secret.Data = map[string][]byte{
		CertChainID:  chain,
		PrivateKeyID: key,
		RootCertID:   rootCert,
	}
	_, err = sc.core.Secrets(saNamespace).Create(secret)
	if err != nil {
		log.Errorf("Failed to create secret (error: %s)", err)
		return
	}

	log.Infof("Istio secret for service account \"%s\" in namespace \"%s\" has been created", saName, saNamespace)
}

func (sc *SecretController) deleteSecret(saName, saNamespace string) {
	err := sc.core.Secrets(saNamespace).Delete(getSecretName(saName), nil)
	// kube-apiserver returns NotFound error when the secret is successfully deleted.
	if err == nil || errors.IsNotFound(err) {
		log.Infof("Istio secret for service account \"%s\" in namespace \"%s\" has been deleted", saName, saNamespace)
		return
	}

	log.Errorf("Failed to delete Istio secret for service account \"%s\" in namespace \"%s\" (error: %s)",
		saName, saNamespace, err)
}

func (sc *SecretController) scrtDeleted(obj interface{}) {
	scrt, ok := obj.(*v1.Secret)
	if !ok {
		log.Warnf("Failed to convert to secret object: %v", obj)
		return
	}

	log.Infof("Re-create deleted Istio secret")

	saName := scrt.Annotations[serviceAccountNameAnnotationKey]
	sc.upsertSecret(saName, scrt.GetNamespace())
}

func (sc *SecretController) generateKeyAndCert(saName string, saNamespace string) ([]byte, []byte, error) {
	id := fmt.Sprintf("%s://cluster.local/ns/%s/sa/%s", util.URIScheme, saNamespace, saName)
	options := util.CertOptions{
		Host:       id,
		RSAKeySize: keySize,
	}

	csrPEM, keyPEM, err := util.GenCSR(options)
	if err != nil {
		return nil, nil, err
	}

	certPEM, err := sc.ca.Sign(csrPEM, sc.certTTL, false)
	if err != nil {
		return nil, nil, err
	}

	return certPEM, keyPEM, nil
}

func (sc *SecretController) scrtUpdated(oldObj, newObj interface{}) {
	scrt, ok := newObj.(*v1.Secret)
	if !ok {
		log.Warnf("Failed to convert to secret object: %v", newObj)
		return
	}

	certBytes := scrt.Data[CertChainID]
	cert, err := util.ParsePemEncodedCertificate(certBytes)
	if err != nil {
		// TODO: we should refresh secret in this case since the secret contains an
		// invalid cert.
		log.Errora(err)
		return
	}

	certLifeTimeLeft := time.Until(cert.NotAfter)
	certLifeTime := cert.NotAfter.Sub(cert.NotBefore)
	// TODO(myidpt): we may introduce a minimum gracePeriod, without making the config too complex.
	gracePeriod := time.Duration(sc.gracePeriodRatio) * certLifeTime
	if gracePeriod < sc.minGracePeriod {
		log.Warnf("gracePeriod (%v * %f) = %v is less than minGracePeriod %v. Apply minGracePeriod.",
			certLifeTime, sc.gracePeriodRatio, sc.minGracePeriod)
		gracePeriod = sc.minGracePeriod
	}
	rootCertificate := sc.ca.GetRootCertificate()

	// Refresh the secret if 1) the certificate contained in the secret is about
	// to expire, or 2) the root certificate in the secret is different than the
	// one held by the ca (this may happen when the CA is restarted and
	// a new self-signed CA cert is generated).
	if certLifeTimeLeft < gracePeriod || !bytes.Equal(rootCertificate, scrt.Data[RootCertID]) {
		namespace := scrt.GetNamespace()
		name := scrt.GetName()

		log.Infof("Refreshing secret %s/%s, either the leaf certificate is about to expire "+
			"or the root certificate is outdated", namespace, name)

		saName := scrt.Annotations[serviceAccountNameAnnotationKey]

		chain, key, err := sc.generateKeyAndCert(saName, namespace)
		if err != nil {
			log.Errorf("Failed to generate key and certificate for service account %q in namespace %q (error %v)",
				saName, namespace, err)

			return
		}

		scrt.Data[CertChainID] = chain
		scrt.Data[PrivateKeyID] = key
		scrt.Data[RootCertID] = rootCertificate

		if _, err = sc.core.Secrets(namespace).Update(scrt); err != nil {
			log.Errorf("Failed to update secret %s/%s (error: %s)", namespace, name, err)
		}
	}
}

func getSecretName(saName string) string {
	return secretNamePrefix + saName
}
