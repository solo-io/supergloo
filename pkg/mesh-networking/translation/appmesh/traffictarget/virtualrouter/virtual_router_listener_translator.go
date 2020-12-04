package virtualrouter

import (
	"strings"

	appmeshv1beta2 "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
)

type listenerTranslator struct{}

func newListenerTranslator() *listenerTranslator {
	return &listenerTranslator{}
}

func (l *listenerTranslator) getListeners(trafficTarget *discoveryv1alpha2.TrafficTarget) []appmeshv1beta2.VirtualRouterListener {
	var listeners []appmeshv1beta2.VirtualRouterListener
	for _, port := range trafficTarget.Spec.GetKubeService().Ports {
		listener := appmeshv1beta2.VirtualRouterListener{
			PortMapping: appmeshv1beta2.PortMapping{
				Port:     appmeshv1beta2.PortNumber(port.Port),
				Protocol: appmeshv1beta2.PortProtocol(strings.ToLower(port.Protocol)),
			},
		}
		listeners = append(listeners, listener)
	}
	return listeners
}
