package retries

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "retries"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewRetriesDecorator()
}

// handles setting Retries on a VirtualService
type retriesDecorator struct {
}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &retriesDecorator{}

func NewRetriesDecorator() *retriesDecorator {
	return &retriesDecorator{}
}

func (d *retriesDecorator) DecoratorName() string {
	return decoratorName
}

func (d *retriesDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *v1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	_ *discoveryv1.MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	retries, err := d.translateRetries(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if retries != nil {
		if err := registerField(&output.Retries, retries); err != nil {
			return err
		}
		output.Retries = retries
	}
	return nil
}

func (d *retriesDecorator) translateRetries(
	trafficPolicy *v1.TrafficPolicySpec,
) (*networkingv1alpha3spec.HTTPRetry, error) {
	retries := trafficPolicy.GetPolicy().GetRetries()
	return trafficpolicyutils.TranslateRetries(retries), nil
}
