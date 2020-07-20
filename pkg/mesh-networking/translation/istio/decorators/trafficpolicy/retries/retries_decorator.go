package retries

import (
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
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

func (p *retriesDecorator) DecoratorName() string {
	return decoratorName
}

func (p *retriesDecorator) ApplyToVirtualService(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	retries, err := p.translateRetries(appliedPolicy.Spec)
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

func (p *retriesDecorator) translateRetries(
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
