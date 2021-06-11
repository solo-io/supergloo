package utils

import (
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
)

// Get the workload labels and TLS port name used to detect ingress gateways in the given cluster.
func GetIngressGatewayDetector(settings *settingsv1.DiscoverySettings, clusterName string) (*settingsv1.DiscoverySettings_Istio_IngressGatewayDetector, error) {
	var deprecatedLabels map[string]string
	var labelSets []*commonv1.LabelSet
	var portName string

	gatewayDetectors := settings.GetIstio().GetIngressGatewayDetectors()
	if gatewayDetectors != nil {
		// First, check if cluster-specific values are set
		if gatewayDetector := gatewayDetectors[clusterName]; gatewayDetector != nil {
			deprecatedLabels = gatewayDetector.GetGatewayWorkloadLabels()
			labelSets = gatewayDetector.GetGatewayWorkloadLabelSets()
			portName = gatewayDetector.GetGatewayTlsPortName()
		}

		// Check the wildcard (all clusters) entry for any unspecified fields
		if wildcardGatewayDetector := gatewayDetectors["*"]; wildcardGatewayDetector != nil {
			if len(deprecatedLabels) < 1 {
				deprecatedLabels = wildcardGatewayDetector.GetGatewayWorkloadLabels()
			}
			if len(labelSets) < 1 {
				labelSets = wildcardGatewayDetector.GetGatewayWorkloadLabelSets()
			}
			if portName == "" {
				portName = wildcardGatewayDetector.GetGatewayTlsPortName()
			}
		}
	}

	// if discovery parameters not specified in Settings, use defaults
	if len(labelSets) < 1 {
		// respect deprecated `GatewayWorkloadLabels` field only if new `GatewayWorkloadLabelSets` field isn't specified
		if len(deprecatedLabels) > 0 {
			labelSets = []*commonv1.LabelSet{
				{
					Labels: deprecatedLabels,
				},
			}
		} else {
			// default workload labels
			labelSets = []*commonv1.LabelSet{
				{
					Labels: defaults.DefaultGatewayWorkloadLabels,
				},
			}
		}
	}
	if portName == "" {
		portName = defaults.DefaultGatewayPortName
	}

	return &settingsv1.DiscoverySettings_Istio_IngressGatewayDetector{
		GatewayWorkloadLabelSets: labelSets,
		GatewayTlsPortName:       portName,
	}, nil
}
