package trafficpolicyutils

import discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"

func ContainsPort(
	destKubeServicePorts []*discoveryv1.DestinationSpec_KubeService_KubeServicePort,
	port uint32,
) bool {
	for _, destPort := range destKubeServicePorts {
		if destPort.Port == port {
			return true
		}
	}
	return false
}
