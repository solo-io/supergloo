package utils

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/go-utils/contextutils"
)

// TODO this is copied from networking snapshot_utils. remove this when functionality is added to skv2
// Safely fetch the single Settings object from snapshot. Log and error if not singleton.
func getSingletonSettings(ctx context.Context, in input.Snapshot) (*v1alpha2.Settings, error) {
	settings := in.Settings().List()
	n := len(settings)
	if n != 1 {
		err := eris.Errorf("Snapshot does not contain single Settings object, %d found.", n)
		contextutils.LoggerFrom(ctx).Errorf("%+v", err)
		return nil, err
	}
	return settings[0], nil
}

// Get the workload labels and TLS port name used to detect ingress gateways in the given cluster.
func GetIngressGatewayMatcher(ctx context.Context, in input.Snapshot, clusterName string) (*v1alpha2.SettingsSpec_Istio_IngressGatewayMatcher, error) {
	settings, err := getSingletonSettings(ctx, in)
	if err != nil {
		return nil, err
	}

	var labels map[string]string
	var portName string

	// First, check if cluster-specific values are set
	clusterMatcherSettings := settings.Spec.GetIstio().GetIngressGatewayMatcherOverrides()
	if clusterMatcherSettings != nil && clusterMatcherSettings[clusterName] != nil {
		labels = clusterMatcherSettings[clusterName].GetGatewayWorkloadLabels()
		portName = clusterMatcherSettings[clusterName].GetGatewayTlsPortName()
	}

	// Check the top-level gateway matcher settings if needed
	if labels == nil || portName == "" {
		matcherSettings := settings.Spec.GetIstio().GetIngressGatewayMatcher()
		if matcherSettings != nil {
			if labels == nil {
				labels = matcherSettings.GetGatewayWorkloadLabels()
			}
			if portName == "" {
				portName = matcherSettings.GetGatewayTlsPortName()
			}
		}
	}

	// Fall back to default values if needed
	if labels == nil {
		labels = defaults.DefaultGatewayWorkloadLabels
	}
	if portName == "" {
		portName = defaults.DefaultGatewayPortName
	}

	return &v1alpha2.SettingsSpec_Istio_IngressGatewayMatcher{
		GatewayWorkloadLabels: labels,
		GatewayTlsPortName:    portName,
	}, nil
}
