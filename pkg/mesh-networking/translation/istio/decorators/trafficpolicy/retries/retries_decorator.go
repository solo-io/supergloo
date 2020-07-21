package retries

import (
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "retries"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(params decorators.Parameters) decorators.Decorator {
	return NewRetriesDecorator()
}

// handles setting Retries on a VirtualService
type retriesDecorator struct {
}

var _ trafficpolicy.VirtualServiceDecorator = &retriesDecorator{}

func NewRetriesDecorator() *retriesDecorator {
	return &retriesDecorator{}
}

func (d *retriesDecorator) DecoratorName() string {
	return decoratorName
}

func (d *retriesDecorator) ApplyToVirtualService(
	appliedPolicy *discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha2.MeshService,
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
	trafficPolicy *v1alpha2.TrafficPolicySpec,
) (*networkingv1alpha3spec.HTTPRetry, error) {
	retries := trafficPolicy.Retries
	if retries == nil {
		return nil, nil
	}
	return &networkingv1alpha3spec.HTTPRetry{
		Attempts:      retries.GetAttempts(),
		PerTryTimeout: retries.GetPerTryTimeout(),
	}, nil
}
