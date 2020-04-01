package istio_enforcer_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	mock_istio_security "github.com/solo-io/mesh-projects/pkg/clients/istio/security/mock"
	"github.com/solo-io/mesh-projects/services/common/constants"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	istio_enforcer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-enforcer/istio-enforcer"
	security_v1beta1 "istio.io/api/security/v1beta1"
	client_security_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		ctx = context.TODO()
		dynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		authPolicyClient = mock_istio_security.NewMockAuthorizationPolicyClient(ctrl)
		istioEnforcer = istio_enforcer.NewIstioEnforcer(
			dynamicClientGetter,
			func(client client.Client) security.AuthorizationPolicyClient {
				return authPolicyClient
			})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var buildMeshes = func() []*discovery_v1alpha1.Mesh {
		clusterNames := []string{"cluster1", "cluster2"}
		installationNamespace := []string{"istio-system-1", "istio-system-2"}
		return []*discovery_v1alpha1.Mesh{
			{
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterNames[0],
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.MeshSpec_IstioMesh{
							Installation: &discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: installationNamespace[0],
							},
						},
					},
				},
			},
			{
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterNames[1],
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.MeshSpec_IstioMesh{
							Installation: &discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: installationNamespace[1],
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
				GetClientForCluster(mesh.Spec.GetCluster().GetName()).
				Return(nil, nil)
			globalAuthPolicy := &client_security_v1beta1.AuthorizationPolicy{
				ObjectMeta: v1.ObjectMeta{
					Name:      istio_enforcer.GlobalAccessControlAuthPolicyName,
					Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
					Labels:    constants.OwnedBySMHLabel,
				},
				Spec: security_v1beta1.AuthorizationPolicy{},
			}
			authPolicyClient.
				EXPECT().
				UpsertSpec(ctx, globalAuthPolicy).
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
				GetClientForCluster(mesh.Spec.GetCluster().GetName()).
				Return(nil, nil)
			authPolicyKey := client.ObjectKey{
				Name:      istio_enforcer.GlobalAccessControlAuthPolicyName,
				Namespace: mesh.Spec.GetIstio().GetInstallation().GetInstallationNamespace(),
			}
			// test
			if i != 0 {
				authPolicyClient.
					EXPECT().
					Get(ctx, authPolicyKey).
					Return(nil, nil)
				authPolicyClient.
					EXPECT().
					Delete(ctx,
						authPolicyKey,
					).
					Return(nil)
			} else {
				// Delete should not be called if no global auth policy exists
				authPolicyClient.
					EXPECT().
					Get(ctx, authPolicyKey).
					Return(nil, errors.NewNotFound(schema.GroupResource{}, ""))
			}
		}
		err := istioEnforcer.StopEnforcing(ctx, meshes)
		Expect(err).ToNot(HaveOccurred())
	})
})
