package istio_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	istio_security "github.com/solo-io/service-mesh-hub/pkg/api/istio/security/v1beta1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/access-control/enforcer/istio"
	istio_federation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/resolver/meshes/istio"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster/mocks"
	mock_istio_security "github.com/solo-io/service-mesh-hub/test/mocks/clients/istio/security/v1alpha3"
	security_v1beta1 "istio.io/api/security/v1beta1"
	"istio.io/api/type/v1beta1"
	client_security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("IstioEnforcer", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		dynamicClientGetter *mock_multicluster.MockDynamicClientGetter
		authPolicyClient    *mock_istio_security.MockAuthorizationPolicyClient
		istioEnforcer       istio.IstioEnforcer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		dynamicClientGetter = mock_multicluster.NewMockDynamicClientGetter(ctrl)
		authPolicyClient = mock_istio_security.NewMockAuthorizationPolicyClient(ctrl)
		istioEnforcer = istio.NewIstioEnforcer(
			dynamicClientGetter,
			func(client client.Client) istio_security.AuthorizationPolicyClient {
				return authPolicyClient
			})
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var buildMesh = func() *smh_discovery.Mesh {
		return &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				Cluster: &smh_core_types.ResourceRef{
					Name: "cluster1",
				},
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system-1",
							},
						},
					},
				},
			},
		}
	}

	It("should start enforcing for Meshes", func() {
		vm := &v1alpha1.VirtualMesh{
			Spec: types.VirtualMeshSpec{
				EnforceAccessControl: types.VirtualMeshSpec_ENABLED,
			},
		}
		mesh := buildMesh()
		dynamicClientGetter.
			EXPECT().
			GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName()).
			Return(nil, nil)
		globalAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      istio.GlobalAccessControlAuthPolicyName,
				Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
				Labels:    kube.OwnedBySMHLabel,
			},
			Spec: security_v1beta1.AuthorizationPolicy{},
		}
		authPolicyClient.
			EXPECT().
			UpsertAuthorizationPolicySpec(ctx, globalAuthPolicy).
			Return(nil)
		ingressAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      istio.IngressGatewayAuthPolicy,
				Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
				Labels:    kube.OwnedBySMHLabel,
			},
			Spec: security_v1beta1.AuthorizationPolicy{
				Action: security_v1beta1.AuthorizationPolicy_ALLOW,
				Selector: &v1beta1.WorkloadSelector{
					MatchLabels: istio_federation.BuildGatewayWorkloadSelector(),
				},
				Rules: []*security_v1beta1.Rule{{}},
			},
		}
		authPolicyClient.
			EXPECT().
			UpsertAuthorizationPolicySpec(ctx, ingressAuthPolicy).
			Return(nil)
		err := istioEnforcer.ReconcileAccessControl(ctx, mesh, vm)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should stop enforcing for Meshes", func() {
		vm := &v1alpha1.VirtualMesh{
			Spec: types.VirtualMeshSpec{
				EnforceAccessControl: types.VirtualMeshSpec_DISABLED,
			},
		}
		mesh := buildMesh()
		dynamicClientGetter.
			EXPECT().
			GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName()).
			Return(nil, nil)
		globalAuthPolicyKey := client.ObjectKey{
			Name:      istio.GlobalAccessControlAuthPolicyName,
			Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
		}
		ingressAuthPolicyKey := client.ObjectKey{
			Name:      istio.IngressGatewayAuthPolicy,
			Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
		}
		// Delete should not be called if no global auth policy exists
		authPolicyClient.
			EXPECT().
			GetAuthorizationPolicy(ctx, globalAuthPolicyKey).
			Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))

		authPolicyClient.
			EXPECT().
			GetAuthorizationPolicy(ctx, ingressAuthPolicyKey).
			Return(nil, nil)
		authPolicyClient.
			EXPECT().
			DeleteAuthorizationPolicy(ctx,
				ingressAuthPolicyKey,
			).
			Return(nil)
		err := istioEnforcer.ReconcileAccessControl(ctx, mesh, vm)
		Expect(err).ToNot(HaveOccurred())
	})
})
