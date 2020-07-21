package accesspolicy

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/decorators"
	securityv1beta1spec "istio.io/api/security/v1beta1"
)

/*
	Interface definitions for decorators which take AccessPolicy as an input and
	decorate a given output resource.
*/

// AuthorizationPolicyDecorators modify the AuthorizationPolicy based on a AccessPolicy which applies to the MeshService.
type AuthorizationPolicyDecorator interface {
	decorators.Decorator

	ApplyToAuthorizationPolicy(
		appliedPolicy *v1alpha2.MeshServiceStatus_AppliedAccessPolicy,
		service *v1alpha2.MeshService,
		output *securityv1beta1spec.Operation,
		registerField decorators.RegisterField,
	) error
}
