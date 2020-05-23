package istio_enforcer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	istio_security "github.com/solo-io/service-mesh-hub/pkg/api/istio/security/v1beta1"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s/mocks"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	istio_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer/istio-enforcer"
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
		istioEnforcer       istio_enforcer.IstioEnforcer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		dynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		authPolicyClient = mock_istio_security.NewMockAuthorizationPolicyClient(ctrl)
		istioEnforcer = istio_enforcer.NewIstioEnforcer(
			dynamicClientGetter,
			func(client client.Client) istio_security.AuthorizationPolicyClient {
				return authPolicyClient
			})
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var buildMeshes = func() []*zephyr_discovery.Mesh {
		clusterNames := []string{"cluster1", "cluster2"}
		installationNamespace := []string{"istio-system-1", "istio-system-2"}
		return []*zephyr_discovery.Mesh{
			{
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterNames[0],
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &zephyr_discovery_types.MeshSpec_Istio1_5{
							Metadata: &zephyr_discovery_types.MeshSpec_IstioMesh{
								Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: installationNamespace[0],
								},
							},
						},
					},
				},
			},
			{
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterNames[1],
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &zephyr_discovery_types.MeshSpec_Istio1_5{
							Metadata: &zephyr_discovery_types.MeshSpec_IstioMesh{
								Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: installationNamespace[1],
								},
							},
						},
					},
				},
			},
		}
	}

	It("should start enforcing for Meshes", func() {
		meshes := buildMeshes()
		for _, mesh := range meshes {
			dynamicClientGetter.
				EXPECT().
				GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName()).
				Return(nil, nil)
			globalAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      istio_enforcer.GlobalAccessControlAuthPolicyName,
					Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
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
					Name:      istio_enforcer.IngressGatewayAuthPolicy,
					Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
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
		}
		err := istioEnforcer.StartEnforcing(ctx, meshes)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should disable for Meshes", func() {
		meshes := buildMeshes()
		for i, mesh := range meshes {
			dynamicClientGetter.
				EXPECT().
				GetClientForCluster(ctx, mesh.Spec.GetCluster().GetName()).
				Return(nil, nil)
			globalAuthPolicyKey := client.ObjectKey{
				Name:      istio_enforcer.GlobalAccessControlAuthPolicyName,
				Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
			}
			ingressAuthPolicyKey := client.ObjectKey{
				Name:      istio_enforcer.IngressGatewayAuthPolicy,
				Namespace: mesh.Spec.GetIstio1_5().GetMetadata().GetInstallation().GetInstallationNamespace(),
			}
			// test
			if i != 0 {
				authPolicyClient.
					EXPECT().
					GetAuthorizationPolicy(ctx, globalAuthPolicyKey).
					Return(nil, nil)
				authPolicyClient.
					EXPECT().
					DeleteAuthorizationPolicy(ctx,
						globalAuthPolicyKey,
					).
					Return(nil)
				authPolicyClient.
					EXPECT().
					GetAuthorizationPolicy(ctx, ingressAuthPolicyKey).
					Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			} else {
				gomock.Any()
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
			}
		}
		err := istioEnforcer.StopEnforcing(ctx, meshes)
		Expect(err).ToNot(HaveOccurred())
	})
})
