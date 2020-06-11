package dns

import (
	"context"

	"github.com/rotisserie/eris"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	UnsupportedServiceType = func(svc *k8s_core_types.Service, clusterName string) error {
		return eris.Errorf("Unsupported service type (%s) found for gateway service (%s.%s) on cluster (%s)",
			svc.Spec.Type, svc.GetName(), svc.GetNamespace(), clusterName)
	}
	NoExternallyResolvableIp = func(svc *k8s_core_types.Service, clusterName string) error {
		return eris.Errorf("Service (%s.%s) of type LoadBalancer on cluster (%s) is not yet externally accessible",
			svc.GetName(), svc.GetNamespace(), clusterName)
	}
	NoAvailableIngresses = func(svc *k8s_core_types.Service, clusterName string) error {
		return eris.Errorf("Service (%s.%s) of type LoadBalancer on cluster (%s) has no ingress",
			svc.GetName(), svc.GetNamespace(), clusterName)
	}
	NoScheduledPods = func(svc *k8s_core_types.Service, clusterName string) error {
		return eris.Errorf("no node found for the service's pods. ensure at least one pod has been deployed "+
			"for the (%s.%s) service on cluster (%s)", svc.GetName(), svc.GetNamespace(), clusterName)
	}
	NoActiveAddressesForNode = func(node *k8s_core_types.Node, clusterName string) error {
		return eris.Errorf("no active addresses found for node (%s) on cluster (%s)",
			node.GetName(), clusterName)
	}
	NoAvailablePorts = func(svc *k8s_core_types.Service, clusterName string) error {
		return eris.Errorf("Service (%s.%s) on cluster (%s) has no ports set",
			svc.GetName(), svc.GetNamespace(), clusterName)
	}
	NamedPortNotFound = func(svc *k8s_core_types.Service, clusterName, portName string) error {
		return eris.Errorf("Service (%s.%s) on cluster (%s) has no ports named %s",
			svc.GetName(), svc.GetNamespace(), clusterName, portName)
	}
)

func NewExternalAccessPointGetter(
	dynamicClientGetter multicluster.DynamicClientGetter,
	podClientFactory k8s_core.PodClientFactory,
	nodeClientFactory k8s_core.NodeClientFactory,
) ExternalAccessPointGetter {
	return &externalAccessPointGetter{
		dynamicClientGetter: dynamicClientGetter,
		podClientFactory:    podClientFactory,
		nodeClientFactory:   nodeClientFactory,
	}
}

type externalAccessPointGetter struct {
	dynamicClientGetter multicluster.DynamicClientGetter
	podClientFactory    k8s_core.PodClientFactory
	nodeClientFactory   k8s_core.NodeClientFactory
}

func (f *externalAccessPointGetter) GetExternalAccessPointForService(
	ctx context.Context,
	svc *k8s_core_types.Service,
	portName, clusterName string,
) (eap ExternalAccessPoint, err error) {
	if len(svc.Spec.Ports) == 0 {
		return eap, NoAvailablePorts(svc, clusterName)
	}

	var servicePort *k8s_core_types.ServicePort
	for _, p := range svc.Spec.Ports {
		if p.Name == portName {
			servicePort = &p
			break
		}
	}

	if servicePort == nil {
		return eap, NamedPortNotFound(svc, clusterName, portName)
	}

	switch svc.Spec.Type {
	case k8s_core_types.ServiceTypeLoadBalancer:
		ingress := svc.Status.LoadBalancer.Ingress
		if len(ingress) == 0 {
			return eap, NoAvailableIngresses(svc, clusterName)
		}

		// depending on the environment, the svc may have either an IP or a Hostname
		// https://istio.io/docs/tasks/traffic-management/ingress/ingress-control/#determining-the-ingress-ip-and-ports
		eap.Address = ingress[0].IP
		if eap.Address == "" {
			eap.Address = ingress[0].Hostname
		}
		if eap.Address == "" {
			return eap, NoExternallyResolvableIp(svc, clusterName)
		}

		eap.Port = uint32(servicePort.Port)
	case k8s_core_types.ServiceTypeNodePort:
		eap.Address, err = f.getNodeIp(ctx, svc, clusterName)
		if err != nil {
			return eap, err
		}
		eap.Port = uint32(servicePort.NodePort)
	default:
		return eap, UnsupportedServiceType(svc, clusterName)
	}

	return eap, nil
}

func (f *externalAccessPointGetter) getNodeIp(ctx context.Context, svc *k8s_core_types.Service, clusterName string) (string, error) {
	dynamicClient, err := f.dynamicClientGetter.GetClientForCluster(ctx, clusterName)
	if err != nil {
		return "", err
	}

	podClient := f.podClientFactory(dynamicClient)
	pods, err := podClient.ListPod(ctx, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector),
		Namespace:     svc.Namespace,
	})
	if err != nil {
		return "", err
	}
	var nodeName string
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			nodeName = pod.Spec.NodeName
			break
		}
	}
	if nodeName == "" {
		return "", NoScheduledPods(svc, clusterName)
	}

	nodeClient := f.nodeClientFactory(dynamicClient)
	node, err := nodeClient.GetNode(ctx, client.ObjectKey{Name: nodeName})
	if err != nil {
		return "", err
	}

	for _, addr := range node.Status.Addresses {
		return addr.Address, nil
	}

	return "", NoActiveAddressesForNode(node, clusterName)
}
