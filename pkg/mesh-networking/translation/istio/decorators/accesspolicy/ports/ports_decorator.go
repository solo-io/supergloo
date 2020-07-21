package ports

import (
	"strconv"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/accesspolicy"
	securityv1beta1spec "istio.io/api/security/v1beta1"
)

const (
	decoratorName = "ports"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewPortsDecorator()
}

// handles setting Cors on a VirtualService
type portsDecorator struct{}

var _ accesspolicy.AuthorizationPolicyDecorator = &portsDecorator{}

func NewPortsDecorator() *portsDecorator {
	return &portsDecorator{}
}

func (d *portsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *portsDecorator) ApplyToAuthorizationPolicy(
	appliedPolicy *v1alpha2.MeshServiceStatus_AppliedAccessPolicy,
	_ *v1alpha2.MeshService,
	output *securityv1beta1spec.Operation,
	registerField decorators.RegisterField,
) error {
	allowedPorts := convertIntsToStrings(appliedPolicy.Spec.AllowedPorts)
	if err := registerField(&output.Ports, allowedPorts); err != nil {
		return err
	}
	output.Ports = allowedPorts
	return nil
}

func convertIntsToStrings(ints []uint32) []string {
	var strings []string
	for _, i := range ints {
		strings = append(strings, strconv.Itoa(int(i)))
	}
	return strings
}
