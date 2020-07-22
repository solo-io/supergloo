package operations

import (
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy"
	securityv1beta1spec "istio.io/api/security/v1beta1"
)

const (
	decoratorName = "operations"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewPathsDecorator()
}

// handles setting Cors on a VirtualService
type operationsDecorator struct{}

var _ accesspolicy.AuthorizationPolicyDecorator = &operationsDecorator{}

func NewPathsDecorator() *operationsDecorator {
	return &operationsDecorator{}
}

func (d *operationsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *operationsDecorator) ApplyToAuthorizationPolicy(
	appliedPolicy *v1alpha2.MeshServiceStatus_AppliedAccessPolicy,
	_ *v1alpha2.MeshService,
	output *securityv1beta1spec.Operation,
	registerField decorators.RegisterField,
) error {
	var errs *multierror.Error
	// allowed paths
	allowedPaths := appliedPolicy.Spec.AllowedPaths
	if err := registerField(&output.Paths, allowedPaths); err != nil {
		errs = multierror.Append(errs, err)
	}
	output.Paths = allowedPaths
	// allowed methods
	allowedMethods := convertHttpMethodsToStrings(appliedPolicy.Spec.AllowedMethods)
	if err := registerField(&output.Methods, allowedMethods); err != nil {
		errs = multierror.Append(errs, err)
	}
	output.Methods = allowedMethods
	// allowed ports
	allowedPorts := convertIntsToStrings(appliedPolicy.Spec.AllowedPorts)
	if err := registerField(&output.Ports, allowedPorts); err != nil {
		errs = multierror.Append(errs, err)
	}
	output.Ports = allowedPorts
	return errs.ErrorOrNil()
}

func convertHttpMethodsToStrings(allowedMethods []types.HttpMethodValue) []string {
	var methods []string
	// Istio considers AuthorizationPolicies without at least one defined To.Operation invalid,
	// The workaround is to populate a dummy "*" for METHOD if not user specified. This guarantees existence of at least
	// one To.Operation.
	if len(allowedMethods) < 1 {
		methods = []string{"*"}
		return methods
	}
	for _, methodEnum := range allowedMethods {
		methods = append(methods, methodEnum.String())
	}
	return methods
}

func convertIntsToStrings(ints []uint32) []string {
	var strings []string
	for _, i := range ints {
		strings = append(strings, strconv.Itoa(int(i)))
	}
	return strings
}
