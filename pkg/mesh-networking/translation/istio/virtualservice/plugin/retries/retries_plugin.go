package retries

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/fieldutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "retries"
)

// handles setting Retries on a VirtualService
type retriesPlugin struct {
}

func NewRetriesPlugin() *retriesPlugin {
	return &retriesPlugin{
	}
}

func (p *retriesPlugin) PluginName() string {
	return pluginName
}

func (p *retriesPlugin) ProcessTrafficPolicy(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	fieldRegistry fieldutils.FieldOwnershipRegistry,
) error {
	retries, err := p.translateRetries(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if retries != nil && !equalityutils.Equals(output.Retries, retries) {
		if err := fieldRegistry.RegisterFieldOwner(
			output.Retries,
			appliedPolicy.Ref,
			0,
		); err != nil {
			return err
		}
		output.Retries = retries
	}
	return nil
}

func (p *retriesPlugin) translateRetries(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) (*istiov1alpha3spec.HTTPRetry, error) {
	retries := trafficPolicy.Retries
	if retries == nil {
		return nil, nil
	}
	return &istiov1alpha3spec.HTTPRetry{
		Attempts:      retries.GetAttempts(),
		PerTryTimeout: retries.GetPerTryTimeout(),
	}, nil
}
