package timeout

import (
	"github.com/gogo/protobuf/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "timeout"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(params decorators.Parameters) decorators.Decorator {
	return NewTimeoutDecorator()
}

// handles setting Timeout on a VirtualService
type timeoutDecorator struct {
}

var _ trafficpolicy.VirtualServiceDecorator = &timeoutDecorator{}

func NewTimeoutDecorator() *timeoutDecorator {
	return &timeoutDecorator{}
}

func (p *timeoutDecorator) DecoratorName() string {
	return decoratorName
}

func (p *timeoutDecorator) DecorateVirtualService(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	timeout, err := p.translateTimeout(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if timeout != nil {
		if err := registerField(&output.Timeout, timeout); err != nil {
			return err
		}
		output.Timeout = timeout
	}
	return nil
}

func (p *timeoutDecorator) translateTimeout(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) (*types.Duration, error) {
	return trafficPolicy.RequestTimeout, nil
}
