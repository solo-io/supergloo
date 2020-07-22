package methods

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy"
	securityv1beta1spec "istio.io/api/security/v1beta1"
)

const (
	decoratorName = "methods"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewMethodsDecorator()
}

// handles setting Cors on a VirtualService
type methodsDecorator struct{}

var _ accesspolicy.AuthorizationPolicyDecorator = &methodsDecorator{}

func NewMethodsDecorator() *methodsDecorator {
	return &methodsDecorator{}
}

func (d *methodsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *methodsDecorator) ApplyToAuthorizationPolicy(
	appliedPolicy *v1alpha2.MeshServiceStatus_AppliedAccessPolicy,
	_ *v1alpha2.MeshService,
	output *securityv1beta1spec.Operation,
	registerField decorators.RegisterField,
) error {
	allowedMethods := convertHttpMethodsToStrings(appliedPolicy.Spec.AllowedMethods)
	if err := registerField(&output.Methods, allowedMethods); err != nil {
		return err
	}
	output.Methods = allowedMethods
	return nil
}

func convertHttpMethodsToStrings(methodEnums []types.HttpMethodValue) []string {
	var methods []string
	for _, methodEnum := range methodEnums {
		methods = append(methods, methodEnum.String())
	}
	return methods
}
