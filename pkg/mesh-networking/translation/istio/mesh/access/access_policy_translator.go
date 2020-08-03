package access

import (
	"github.com/solo-io/go-utils/kubeutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	securityv1beta1spec "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./access_policy_translator.go -destination mocks/access_policy_translator.go

// The access control translator translates a VirtualMesh EnforcementPolicy into Istio AuthorizationPolicies.
type Translator interface {
	// Returns nil if no AuthorizationPolicies are required for the mesh (i.e. because access policy enforcement is disabled).
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		mesh *discoveryv1alpha2.Mesh,
		virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
		outputs output.Builder,
	)
}

const (
	IngressGatewayAuthPolicyName      = "allow-ingress-gateway"
	GlobalAccessControlAuthPolicyName = "global-access-control"
)

type translator struct{}

func NewTranslator() Translator {
	return &translator{}
}

func (t *translator) Translate(
	mesh *discoveryv1alpha2.Mesh,
	virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
	outputs output.Builder,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		return
	}

	// Istio's default access enforcement policy is disabled.
	if virtualMesh.Spec.GlobalAccessPolicy == v1alpha2.VirtualMeshSpec_MESH_DEFAULT ||
		virtualMesh.Spec.GlobalAccessPolicy == v1alpha2.VirtualMeshSpec_DISABLED {
		return
	}
	clusterName := istioMesh.Installation.Cluster
	installationNamespace := istioMesh.Installation.Namespace
	globalAuthPolicy := buildGlobalAuthPolicy(installationNamespace, clusterName)
	ingressGatewayAuthPolicies := buildAuthPoliciesForIngressgateways(installationNamespace, clusterName, istioMesh.IngressGateways)

	outputs.AddAuthorizationPolicies(globalAuthPolicy)
	outputs.AddAuthorizationPolicies(ingressGatewayAuthPolicies...)
}

// Creates an AuthorizationPolicy that allows all traffic into the "istio-ingressgateway" service
// which backs the Gateway used for multi cluster traffic.
func buildAuthPoliciesForIngressgateways(
	installationNamespace string,
	clusterName string,
	ingressGateways []*discoveryv1alpha2.MeshSpec_Istio_IngressGatewayInfo,
) []*securityv1beta1.AuthorizationPolicy {
	var authPolicies []*securityv1beta1.AuthorizationPolicy
	for _, ingressGateway := range ingressGateways {
		authPolicies = append(authPolicies, &securityv1beta1.AuthorizationPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name:        ingressGatewayAuthPolicyName(ingressGateway),
				Namespace:   installationNamespace,
				ClusterName: clusterName,
				Labels:      metautils.TranslatedObjectLabels(),
			},
			Spec: securityv1beta1spec.AuthorizationPolicy{
				Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
				// A single empty rule allows all traffic.
				// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
				Rules: []*securityv1beta1spec.Rule{{}},
				Selector: &v1beta1.WorkloadSelector{
					MatchLabels: ingressGateway.WorkloadLabels,
				},
			},
		})
	}
	return authPolicies
}

// Creates a global AuthorizationPolicy that denies all traffic within the Mesh unless explicitly allowed by SMH AccessControl resources.
func buildGlobalAuthPolicy(
	installationNamespace,
	clusterName string,
) *securityv1beta1.AuthorizationPolicy {
	// The following config denies all traffic in the mesh because it defaults to an ALLOW rule that doesn't match any requests,
	// set to the installation namespace so it affects all namespaces,
	// thereby denying traffic unless explicitly allowed by the user through additional AuthorizationPolicies.
	// https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
	return &securityv1beta1.AuthorizationPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:        GlobalAccessControlAuthPolicyName,
			Namespace:   installationNamespace,
			ClusterName: clusterName,
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: securityv1beta1spec.AuthorizationPolicy{},
	}
}

func ingressGatewayAuthPolicyName(ingressGateway *discoveryv1alpha2.MeshSpec_Istio_IngressGatewayInfo) string {
	return IngressGatewayAuthPolicyName + "-" + kubeutils.SanitizeNameV2(ingressGateway.ExternalAddress)
}
