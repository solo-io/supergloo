package timeout

import (
	"github.com/gogo/protobuf/types"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
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

var _ decorators.TrafficPolicyVirtualServiceDecorator = &timeoutDecorator{}

func NewTimeoutDecorator() *timeoutDecorator {
	return &timeoutDecorator{}
}

func (d *timeoutDecorator) DecoratorName() string {
	return decoratorName
}

func (d *timeoutDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *v1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	_ *discoveryv1.MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	timeout, err := d.translateTimeout(appliedPolicy.Spec)
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

func (d *timeoutDecorator) translateTimeout(
	trafficPolicy *v1.TrafficPolicySpec,
) (*types.Duration, error) {
	return trafficpolicyutils.TranslateTimeout(trafficPolicy.GetPolicy().GetRequestTimeout()), nil
}
