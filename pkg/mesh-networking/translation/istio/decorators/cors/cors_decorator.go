package cors

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/trafficpolicyutils"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	decoratorName = "cors"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewCorsDecorator()
}

// handles setting Cors on a VirtualService
type corsDecorator struct{}

var _ decorators.TrafficPolicyVirtualServiceDecorator = &corsDecorator{}

func NewCorsDecorator() *corsDecorator {
	return &corsDecorator{}
}

func (d *corsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *corsDecorator) ApplyTrafficPolicyToVirtualService(
	appliedPolicy *v1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	_ *discoveryv1.MeshInstallation,
	output *networkingv1alpha3spec.HTTPRoute,
	registerField decorators.RegisterField,
) error {
	cors, err := d.translateCors(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if cors != nil {
		if err := registerField(&output.CorsPolicy, cors); err != nil {
			return err
		}
		output.CorsPolicy = cors
	}
	return nil
}

func (d *corsDecorator) translateCors(
	trafficPolicy *v1.TrafficPolicySpec,
) (*networkingv1alpha3spec.CorsPolicy, error) {
	corsPolicy := trafficPolicy.GetPolicy().GetCorsPolicy()
	return trafficpolicyutils.TranslateCorsPolicy(corsPolicy)
}
