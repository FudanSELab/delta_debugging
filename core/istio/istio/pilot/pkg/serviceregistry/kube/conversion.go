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

package kube

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/model"
)

type kubeServiceNode struct {
	// PodName Specifies the name of the POD
	PodName string

	// Namespace specifies the name of the namespace the pod belongs to
	Namespace string

	// Domain specifies the pod's domain
	Domain string
}

const (
	// IngressClassAnnotation is the annotation on ingress resources for the class of controllers
	// responsible for it
	IngressClassAnnotation = "kubernetes.io/ingress.class"

	// KubeServiceAccountsOnVMAnnotation is to specify the K8s service accounts that are allowed to run
	// this service on the VMs
	KubeServiceAccountsOnVMAnnotation = "alpha.istio.io/kubernetes-serviceaccounts"

	// CanonicalServiceAccountsOnVMAnnotation is to specify the non-Kubernetes service accounts that
	// are allowed to run this service on the VMs
	CanonicalServiceAccountsOnVMAnnotation = "alpha.istio.io/canonical-serviceaccounts"

	// IstioURIPrefix is the URI prefix in the Istio service account scheme
	IstioURIPrefix = "spiffe"

	// PortAuthenticationAnnotationKeyPrefix is the annotation key prefix that used to define
	// authentication policy.
	PortAuthenticationAnnotationKeyPrefix = "auth.istio.io"
)

func convertLabels(obj meta_v1.ObjectMeta) model.Labels {
	out := make(model.Labels, len(obj.Labels))
	for k, v := range obj.Labels {
		out[k] = v
	}
	return out
}

// Extracts security option for given port from annotation. If there is no such
// annotation, or the annotation value is not recognized, returns
// meshconfig.AuthenticationPolicy_INHERIT
func extractAuthenticationPolicy(port v1.ServicePort, obj meta_v1.ObjectMeta) meshconfig.AuthenticationPolicy {
	if obj.Annotations == nil {
		return meshconfig.AuthenticationPolicy_INHERIT
	}
	if val, ok := meshconfig.AuthenticationPolicy_value[obj.Annotations[portAuthenticationAnnotationKey(int(port.Port))]]; ok {
		return meshconfig.AuthenticationPolicy(val)
	}
	return meshconfig.AuthenticationPolicy_INHERIT
}

func convertPort(port v1.ServicePort, obj meta_v1.ObjectMeta) *model.Port {
	return &model.Port{
		Name:                 port.Name,
		Port:                 int(port.Port),
		Protocol:             ConvertProtocol(port.Name, port.Protocol),
		AuthenticationPolicy: extractAuthenticationPolicy(port, obj),
	}
}

func convertService(svc v1.Service, domainSuffix string) *model.Service {
	addr, external := "", ""
	if svc.Spec.ClusterIP != "" && svc.Spec.ClusterIP != v1.ClusterIPNone {
		addr = svc.Spec.ClusterIP
	}

	if svc.Spec.Type == v1.ServiceTypeExternalName && svc.Spec.ExternalName != "" {
		external = svc.Spec.ExternalName
	}

	ports := make([]*model.Port, 0, len(svc.Spec.Ports))
	for _, port := range svc.Spec.Ports {
		ports = append(ports, convertPort(port, svc.ObjectMeta))
	}

	loadBalancingDisabled := addr == "" && external == "" // headless services should not be load balanced

	serviceaccounts := make([]string, 0)
	if svc.Annotations != nil {
		if svc.Annotations[CanonicalServiceAccountsOnVMAnnotation] != "" {
			for _, csa := range strings.Split(svc.Annotations[CanonicalServiceAccountsOnVMAnnotation], ",") {
				serviceaccounts = append(serviceaccounts, canonicalToIstioServiceAccount(csa))
			}
		}
		if svc.Annotations[KubeServiceAccountsOnVMAnnotation] != "" {
			for _, ksa := range strings.Split(svc.Annotations[KubeServiceAccountsOnVMAnnotation], ",") {
				serviceaccounts = append(serviceaccounts, kubeToIstioServiceAccount(ksa, svc.Namespace, domainSuffix))
			}
		}
	}
	sort.Sort(sort.StringSlice(serviceaccounts))

	return &model.Service{
		Hostname:              serviceHostname(svc.Name, svc.Namespace, domainSuffix),
		Ports:                 ports,
		Address:               addr,
		ExternalName:          external,
		ServiceAccounts:       serviceaccounts,
		LoadBalancingDisabled: loadBalancingDisabled,
	}
}

// serviceHostname produces FQDN for a k8s service
func serviceHostname(name, namespace, domainSuffix string) string {
	return fmt.Sprintf("%s.%s.svc.%s", name, namespace, domainSuffix)
}

// canonicalToIstioServiceAccount converts a Canonical service account to an Istio service account
func canonicalToIstioServiceAccount(saname string) string {
	return fmt.Sprintf("%v://%v", IstioURIPrefix, saname)
}

func portAuthenticationAnnotationKey(port int) string {
	return fmt.Sprintf("%s/%d", PortAuthenticationAnnotationKeyPrefix, port)
}

// kubeToIstioServiceAccount converts a K8s service account to an Istio service account
func kubeToIstioServiceAccount(saname string, ns string, domain string) string {
	return fmt.Sprintf("%v://%v/ns/%v/sa/%v", IstioURIPrefix, domain, ns, saname)
}

// KeyFunc is the internal API key function that returns "namespace"/"name" or
// "name" if "namespace" is empty
func KeyFunc(name, namespace string) string {
	if len(namespace) == 0 {
		return name
	}
	return namespace + "/" + name
}

// parseHostname extracts service name and namespace from the service hostname
func parseHostname(hostname string) (name string, namespace string, err error) {
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		err = fmt.Errorf("missing service name and namespace from the service hostname %q", hostname)
		return
	}
	name = parts[0]
	namespace = parts[1]
	return
}

// parsePodID extracts POD name and namespace from the service node ID
func parsePodID(nodeID string) (podname string, namespace string, err error) {
	parts := strings.Split(nodeID, ".")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid ID %q. Should be <pod name>.<namespace>", nodeID)
		return
	}
	podname = parts[0]
	namespace = parts[1]
	return
}

// parseDomain extracts the service node's domain
func parseDomain(nodeDomain string) (namespace string, err error) {
	parts := strings.Split(nodeDomain, ".")
	if len(parts) != 4 {
		err = fmt.Errorf("invalid node domain format %q. Should be <namespace>.svc.cluster.local", nodeDomain)
		return
	}
	if parts[1] != "svc" || parts[2] != "cluster" || parts[3] != "local" {
		err = fmt.Errorf("invalid node domain %q. Should be <namespace>.svc.cluster.local", nodeDomain)
		return
	}
	namespace = parts[0]
	return
}

func parseKubeServiceNode(IPAddress string, node *model.Proxy, kubeNodes map[string]*kubeServiceNode) (err error) {
	podname, namespace, err := parsePodID(node.ID)
	if err != nil {
		return
	}
	namespace1, err := parseDomain(node.Domain)
	if err != nil {
		return
	}
	if namespace != namespace1 {
		err = fmt.Errorf("namespace in ID %q must be equal to that in domain %q", node.ID, node.Domain)
	}
	kubeNodes[IPAddress] = &kubeServiceNode{
		PodName:   podname,
		Namespace: namespace,
		Domain:    node.Domain}
	return
}

// ConvertProtocol from k8s protocol and port name
func ConvertProtocol(name string, proto v1.Protocol) model.Protocol {
	out := model.ProtocolTCP
	switch proto {
	case v1.ProtocolUDP:
		out = model.ProtocolUDP
	case v1.ProtocolTCP:
		prefix := name
		i := strings.Index(name, "-")
		if i >= 0 {
			prefix = name[:i]
		}
		protocol := model.ConvertCaseInsensitiveStringToProtocol(prefix)
		if protocol != model.ProtocolUDP && protocol != model.ProtocolUnsupported {
			out = protocol
		}
	}
	return out
}

func convertProbePort(c v1.Container, handler *v1.Handler) (*model.Port, error) {
	if handler == nil {
		return nil, nil
	}

	var protocol model.Protocol
	var portVal intstr.IntOrString
	var port int

	// Only one type of handler is allowed by Kubernetes (HTTPGet or TCPSocket)
	if handler.HTTPGet != nil {
		portVal = handler.HTTPGet.Port
		protocol = model.ProtocolHTTP
	} else if handler.TCPSocket != nil {
		portVal = handler.TCPSocket.Port
		protocol = model.ProtocolTCP
	} else {
		return nil, nil
	}

	switch portVal.Type {
	case intstr.Int:
		port = portVal.IntValue()
		return &model.Port{
			Name:     "mgmt-" + strconv.Itoa(port),
			Port:     port,
			Protocol: protocol,
		}, nil
	case intstr.String:
		for _, named := range c.Ports {
			if named.Name == portVal.String() {
				port = int(named.ContainerPort)
				return &model.Port{
					Name:     "mgmt-" + strconv.Itoa(port),
					Port:     port,
					Protocol: protocol,
				}, nil
			}
		}
		return nil, fmt.Errorf("missing named port %q", portVal)
	default:
		return nil, fmt.Errorf("incorrect port type %q", portVal)
	}
}

// convertProbesToPorts returns a PortList consisting of the ports where the
// pod is configured to do Liveness and Readiness probes
func convertProbesToPorts(t *v1.PodSpec) (model.PortList, error) {
	set := make(map[string]*model.Port)
	var errs error
	for _, container := range t.Containers {
		for _, probe := range []*v1.Probe{container.LivenessProbe, container.ReadinessProbe} {
			if probe == nil {
				continue
			}

			p, err := convertProbePort(container, &probe.Handler)
			if err != nil {
				errs = multierror.Append(errs, err)
			} else if p != nil && set[p.Name] == nil {
				// Deduplicate along the way. We don't differentiate between HTTP vs TCP mgmt ports
				set[p.Name] = p
			}
		}
	}

	mgmtPorts := make(model.PortList, 0, len(set))
	for _, p := range set {
		mgmtPorts = append(mgmtPorts, p)
	}
	sort.Slice(mgmtPorts, func(i, j int) bool { return mgmtPorts[i].Port < mgmtPorts[j].Port })

	return mgmtPorts, errs
}
