package k8s_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_k8s_apps_clients "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/mocks"
	mock_k8s_core_clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	mock_kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
	mock_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s/mocks"
	apps_v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Mesh Finder", func() {
	var (
		ctrl                          *gomock.Controller
		ctx                           = context.TODO()
		clusterName                   = "cluster-name"
		mockMeshScanner               *mock_discovery.MockMeshScanner
		mockMeshClient                *mock_core.MockMeshClient
		mockAppsMulticlusterClientset *mock_k8s_apps_clients.MockMulticlusterClientset
		mockCoreMulticlusterClientset *mock_k8s_core_clients.MockMulticlusterClientset
		mockAppsClientset             *mock_k8s_apps_clients.MockClientset
		mockCoreClientset             *mock_k8s_core_clients.MockClientset
		mockDeploymentClient          *mock_k8s_apps_clients.MockDeploymentClient
		mockConfigMapClient           *mock_kubernetes_core.MockConfigMapClient
		meshDiscovery                 k8s.MeshDiscovery
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockMeshScanner = mock_discovery.NewMockMeshScanner(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockAppsMulticlusterClientset = mock_k8s_apps_clients.NewMockMulticlusterClientset(ctrl)
		mockCoreMulticlusterClientset = mock_k8s_core_clients.NewMockMulticlusterClientset(ctrl)
		mockAppsClientset = mock_k8s_apps_clients.NewMockClientset(ctrl)
		mockCoreClientset = mock_k8s_core_clients.NewMockClientset(ctrl)
		mockDeploymentClient = mock_k8s_apps_clients.NewMockDeploymentClient(ctrl)
		mockConfigMapClient = mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		meshDiscovery = k8s.NewMeshDiscovery(
			[]k8s.MeshScanner{mockMeshScanner},
			mockMeshClient,
			mockAppsMulticlusterClientset,
			mockCoreMulticlusterClientset,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectDiscoverMesh = func() {
		mockAppsMulticlusterClientset.EXPECT().Cluster(clusterName).Return(mockAppsClientset, nil)
		mockCoreMulticlusterClientset.EXPECT().Cluster(clusterName).Return(mockCoreClientset, nil)
		mockAppsClientset.EXPECT().Deployments().Return(mockDeploymentClient)
		mockCoreClientset.EXPECT().ConfigMaps().Return(mockConfigMapClient)
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
			ScanDeployment(ctx, clusterName, deployment1, mockConfigMapClient).
			Return(mesh1, nil)
		mockMeshScanner.
			EXPECT().
			ScanDeployment(ctx, clusterName, deployment2, mockConfigMapClient).
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
