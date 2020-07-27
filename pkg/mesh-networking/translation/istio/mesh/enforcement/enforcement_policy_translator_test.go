package enforcement_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1beta1sets "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/sets"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/enforcement"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	securityv1beta1spec "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("EnforcementPolicyTranslator", func() {
	var (
		translator enforcement.Translator
	)

	BeforeEach(func() {
		translator = enforcement.NewTranslator()
	})

	It("should translate an AuthorizationPolicy for the ingress gateway and in the installation namespace", func() {
		mesh := &discoveryv1alpha2.Mesh{
			Spec: discoveryv1alpha2.MeshSpec{
				MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
					Istio: &discoveryv1alpha2.MeshSpec_Istio{
						Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
							Namespace: "istio-system",
						},
					},
				},
			},
			Status: discoveryv1alpha2.MeshStatus{
				AppliedVirtualMeshes: []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
					{
						Spec: &networkingv1alpha2.VirtualMeshSpec{
							EnforceAccessControl: networkingv1alpha2.VirtualMeshSpec_ENABLED,
						},
					},
				},
			},
		}
		expectedAuthPolicies := v1beta1sets.NewAuthorizationPolicySet(
			&securityv1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      enforcement.IngressGatewayAuthPolicyName,
					Namespace: "istio-system",
					Labels:    metautils.TranslatedObjectLabels(),
				},
				Spec: securityv1beta1spec.AuthorizationPolicy{
					Action: securityv1beta1spec.AuthorizationPolicy_ALLOW,
					// A single empty rule allows all traffic.
					// Reference: https://istio.io/docs/reference/config/security/authorization-policy/#AuthorizationPolicy
					Rules: []*securityv1beta1spec.Rule{{}},
					Selector: &v1beta1.WorkloadSelector{
						MatchLabels: defaults.DefaultGatewayWorkloadLabels,
					},
				},
			},
			&securityv1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      enforcement.GlobalAccessControlAuthPolicyName,
					Namespace: "istio-system",
					Labels:    metautils.TranslatedObjectLabels(),
				},
				Spec: securityv1beta1spec.AuthorizationPolicy{},
			},
		)
		authPolicies := translator.Translate(nil, mesh, mesh.Status.AppliedVirtualMeshes[0], nil)
		Expect(authPolicies).To(Equal(expectedAuthPolicies))
	})

	It("should not translate any AuthorizationPolicies", func() {
		mesh := &discoveryv1alpha2.Mesh{
			Spec: discoveryv1alpha2.MeshSpec{
				MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
					Istio: &discoveryv1alpha2.MeshSpec_Istio{
						Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
							Namespace: "istio-system",
						},
					},
				},
			},
			Status: discoveryv1alpha2.MeshStatus{
				AppliedVirtualMeshes: []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
					{
						Spec: &networkingv1alpha2.VirtualMeshSpec{
							EnforceAccessControl: networkingv1alpha2.VirtualMeshSpec_DISABLED,
						},
					},
				},
			},
		}
		authPolicies := translator.Translate(nil, mesh, mesh.Status.AppliedVirtualMeshes[0], nil)
		Expect(authPolicies).To(BeNil())
	})
})
