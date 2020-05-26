package istio_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/access-control/enforcer/istio"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	istio_security "github.com/solo-io/service-mesh-hub/pkg/api/istio/security/v1beta1"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s/mocks"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	istio_federation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/resolver/meshes/istio"
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
		dynamicClientGetter *mock_mc_manager.MockDynamicClientGetter
		authPolicyClient    *mock_istio_security.MockAuthorizationPolicyClient
		istioEnforcer       istio.IstioEnforcer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		dynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
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

	var buildMesh = func() *zephyr_discovery.Mesh {
		return &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: "cluster1",
				},
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{
					Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{
						Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: "istio-system-1",
						},
					},
				},
			},
		}
	}

	It("should start enforcing for Meshes", func() {
		mesh := buildMesh()
		dynamicClientGetter.
			EXPECT().
			GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName()).
			Return(nil, nil)
		globalAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      istio.GlobalAccessControlAuthPolicyName,
				Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
				Labels:    constants.OwnedBySMHLabel,
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
				Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
				Labels:    constants.OwnedBySMHLabel,
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
		err := istioEnforcer.StartEnforcing(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should disable for Meshes", func() {
		mesh := buildMesh()
		dynamicClientGetter.
			EXPECT().
			GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName()).
			Return(nil, nil)
		globalAuthPolicyKey := client.ObjectKey{
			Name:      istio.GlobalAccessControlAuthPolicyName,
			Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
		}
		ingressAuthPolicyKey := client.ObjectKey{
			Name:      istio.IngressGatewayAuthPolicy,
			Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
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
		err := istioEnforcer.StopEnforcing(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
	})
})
