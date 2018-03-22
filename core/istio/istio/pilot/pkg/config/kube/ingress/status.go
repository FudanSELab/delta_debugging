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

package ingress

import (
	"errors"
	"fmt"
	"os"
	// TODO(nmittler): Remove this
	_ "github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	betaext "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/ingress/core/pkg/ingress/status"
	"k8s.io/ingress/core/pkg/ingress/store"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/serviceregistry/kube"
	"istio.io/istio/pkg/log"
)

const ingressElectionID = "istio-ingress-controller-leader"

// StatusSyncer keeps the status IP in each Ingress resource updated
type StatusSyncer struct {
	sync     status.Sync
	informer cache.SharedIndexInformer
}

// Run the syncer until stopCh is closed
func (s *StatusSyncer) Run(stopCh <-chan struct{}) {
	go func() {
		s.sync.Run(stopCh)
		s.sync.Shutdown()
	}()
	go s.informer.Run(stopCh)
	<-stopCh
}

// NewStatusSyncer creates a new instance
func NewStatusSyncer(mesh *meshconfig.MeshConfig,
	client kubernetes.Interface,
	ingressNamespace string,
	options kube.ControllerOptions) (*StatusSyncer, error) {
	if _, exists := os.LookupEnv("POD_NAME"); !exists {
		return nil, errors.New("POD_NAME environment variable must be defined")
	}
	if _, exists := os.LookupEnv("POD_NAMESPACE"); !exists {
		return nil, errors.New("POD_NAMESPACE environment variable must be defined")
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts meta_v1.ListOptions) (runtime.Object, error) {
				return client.ExtensionsV1beta1().Ingresses(options.WatchedNamespace).List(opts)
			},
			WatchFunc: func(opts meta_v1.ListOptions) (watch.Interface, error) {
				return client.ExtensionsV1beta1().Ingresses(options.WatchedNamespace).Watch(opts)
			},
		},
		&v1beta1.Ingress{}, options.ResyncPeriod, cache.Indexers{},
	)

	var publishService string
	if mesh.IngressService != "" {
		publishService = fmt.Sprintf("%v/%v", ingressNamespace, mesh.IngressService)
	}
	log.Infof("ingress status syncer publishService %s", publishService)
	ingressClass, defaultIngressClass := convertIngressControllerMode(mesh.IngressControllerMode, mesh.IngressClass)

	customIngressStatus := func(*betaext.Ingress) []v1.LoadBalancerIngress {
		return nil
	}

	sync := status.NewStatusSyncer(status.Config{
		Client:              client,
		IngressLister:       store.IngressLister{Store: informer.GetStore()},
		ElectionID:          ingressElectionID, // TODO: configurable?
		PublishService:      publishService,
		DefaultIngressClass: defaultIngressClass,
		IngressClass:        ingressClass,
		CustomIngressStatus: customIngressStatus,
	})

	return &StatusSyncer{
		sync:     sync,
		informer: informer,
	}, nil
}

// convertIngressControllerMode converts Ingress controller mode into k8s ingress status syncer ingress class and
// default ingress class. Ingress class and default ingress class are used by the syncer to determine whether or not to
// update the IP of a ingress resource.
func convertIngressControllerMode(mode meshconfig.MeshConfig_IngressControllerMode,
	class string) (string, string) {
	var ingressClass, defaultIngressClass string
	switch mode {
	case meshconfig.MeshConfig_DEFAULT:
		defaultIngressClass = class
		ingressClass = class
	case meshconfig.MeshConfig_STRICT:
		ingressClass = class
	}
	return ingressClass, defaultIngressClass
}
