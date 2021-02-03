package utils

import (
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
)

// Get the workload labels and TLS port name used to detect ingress gateways in the given cluster.
func GetIngressGatewayDetector(settings *settingsv1alpha2.DiscoverySettings, clusterName string) (*settingsv1alpha2.DiscoverySettings_Istio_IngressGatewayDetector, error) {
	var labels map[string]string
	var portName string

	gatewayDetectors := settings.GetIstio().GetIngressGatewayDetectors()
	if gatewayDetectors != nil {
		// First, check if cluster-specific values are set
		if gatewayDetectors[clusterName] != nil {
			labels = gatewayDetectors[clusterName].GetGatewayWorkloadLabels()
			portName = gatewayDetectors[clusterName].GetGatewayTlsPortName()
		}

		// Check the wildcard (all clusters) entry
		if (labels == nil || portName == "") && gatewayDetectors["*"] != nil {
			if labels == nil {
				labels = gatewayDetectors["*"].GetGatewayWorkloadLabels()
			}
			if portName == "" {
				portName = gatewayDetectors["*"].GetGatewayTlsPortName()
			}
		}
	}

	// Fall back to default values
	if labels == nil {
		labels = defaults.DefaultIngressGatewayWorkloadLabels
	}
	if portName == "" {
		portName = defaults.DefaultGatewayPortName
	}

	return &settingsv1alpha2.DiscoverySettings_Istio_IngressGatewayDetector{
		GatewayWorkloadLabels: labels,
		GatewayTlsPortName:    portName,
	}, nil
}
