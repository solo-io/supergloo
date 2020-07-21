package paths

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy"
	securityv1beta1spec "istio.io/api/security/v1beta1"
)

const (
	decoratorName = "paths"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewPathsDecorator()
}

// handles setting Cors on a VirtualService
type pathsDecorator struct{}

var _ accesspolicy.AuthorizationPolicyDecorator = &pathsDecorator{}

func NewPathsDecorator() *pathsDecorator {
	return &pathsDecorator{}
}

func (d *pathsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *pathsDecorator) ApplyToAuthorizationPolicy(
	appliedPolicy *v1alpha2.MeshServiceStatus_AppliedAccessPolicy,
	_ *v1alpha2.MeshService,
	output *securityv1beta1spec.Operation,
	registerField decorators.RegisterField,
) error {
	allowedPaths := appliedPolicy.Spec.AllowedPaths
	if err := registerField(&output.Paths, allowedPaths); err != nil {
		return err
	}
	output.Paths = allowedPaths
	return nil
}
