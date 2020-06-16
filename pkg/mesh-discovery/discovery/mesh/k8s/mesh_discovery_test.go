package k8s_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	mock_k8s_apps_clients "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/mocks"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
	mock_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/mocks"
	mock_controller_runtime "github.com/solo-io/service-mesh-hub/test/mocks/controller-runtime"
	apps_v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Mesh Finder", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     = context.TODO()
		clusterName             = "cluster-name"
		mockMeshScanner         *mock_discovery.MockMeshScanner
		mockMeshClient          *mock_core.MockMeshClient
		mockDynamicClientGetter *mock_multicluster.MockDynamicClientGetter
		mockClusterClient       *mock_controller_runtime.MockClient
		mockDeploymentClient    *mock_k8s_apps_clients.MockDeploymentClient
		meshDiscovery           k8s.MeshDiscovery
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockMeshScanner = mock_discovery.NewMockMeshScanner(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockDynamicClientGetter = mock_multicluster.NewMockDynamicClientGetter(ctrl)
		mockDeploymentClient = mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
		mockClusterClient = mock_controller_runtime.NewMockClient(ctrl)
		meshDiscovery = k8s.NewMeshDiscovery(
			[]k8s.MeshScanner{mockMeshScanner},
			mockMeshClient,
			func(client client.Client) v1.DeploymentClient {
				return mockDeploymentClient
			},
			mockDynamicClientGetter,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectDiscoverMesh = func() {
		// Doesn't matter what is returned here, the mock factory will return mockDeploymentClient
		mockDynamicClientGetter.EXPECT().GetClientForCluster(ctx, clusterName).Return(mockClusterClient, nil)
		// To be created
		mesh1 := &smh_discovery.Mesh{
			ObjectMeta: metav1.ObjectMeta{Name: "mesh1", Namespace: "mesh1-namespace"},
		}
		// To be updated
		mesh2 := &smh_discovery.Mesh{
			ObjectMeta: metav1.ObjectMeta{Name: "mesh2", Namespace: "mesh2-namespace"},
		}
		mesh2discovered := &smh_discovery.Mesh{
			ObjectMeta: metav1.ObjectMeta{Name: "mesh2", Namespace: "mesh2-namespace"},
			Spec: types.MeshSpec{
				MeshType: &types.MeshSpec_Istio1_5_{},
			},
		}
		// To be deleted
		mesh3 := &smh_discovery.Mesh{
			ObjectMeta: metav1.ObjectMeta{Name: "mesh3", Namespace: "mesh3-namespace"},
		}
		mockMeshClient.
			EXPECT().
			ListMesh(ctx, client.MatchingLabels{kube.COMPUTE_TARGET: clusterName}).
			Return(&smh_discovery.MeshList{Items: []smh_discovery.Mesh{*mesh2, *mesh3}}, nil)
		deployment1 := &apps_v1.Deployment{}
		deployment2 := &apps_v1.Deployment{}
		mockDeploymentClient.
			EXPECT().
			ListDeployment(ctx).
			Return(&apps_v1.DeploymentList{Items: []apps_v1.Deployment{*deployment1, *deployment2}}, nil)
		mockMeshScanner.
			EXPECT().
			ScanDeployment(ctx, clusterName, deployment1, mockClusterClient).
			Return(mesh1, nil)
		mockMeshScanner.
			EXPECT().
			ScanDeployment(ctx, clusterName, deployment2, mockClusterClient).
			Return(mesh2discovered, nil)
		mockMeshClient.EXPECT().UpsertMesh(ctx, mesh1).Return(nil)
		mockMeshClient.EXPECT().UpsertMesh(ctx, mesh2discovered).Return(nil)
		mockMeshClient.EXPECT().DeleteMesh(ctx, selection.ObjectMetaToObjectKey(mesh3.ObjectMeta)).Return(nil)
	}

	Context("Create Event", func() {
		It("can discover a mesh", func() {
			expectDiscoverMesh()
			err := meshDiscovery.DiscoverMesh(ctx, clusterName)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
