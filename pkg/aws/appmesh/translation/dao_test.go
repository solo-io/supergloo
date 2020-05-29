package translation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	types3 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws/appmesh"
	mock_appmesh "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/aws/appmesh/translation"
	"github.com/solo-io/service-mesh-hub/pkg/collections/sets"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Dao", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		mockMeshServiceClient  *mock_core.MockMeshServiceClient
		mockMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockAcpClient          *mock_zephyr_networking.MockAccessControlPolicyClient
		mockResourceSelector   *mock_selector.MockResourceSelector
		mockAppmeshClient      *mock_appmesh.MockAppmeshClient
		dao                    translation.AppmeshTranslationDao
		mesh                   = &zephyr_discovery.Mesh{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mesh-name",
				Namespace: "mesh-namespace",
			},
			Spec: types.MeshSpec{
				MeshType: &types.MeshSpec_AwsAppMesh_{
					AwsAppMesh: &types.MeshSpec_AwsAppMesh{
						Name: "appmesh-name",
					},
				},
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockAcpClient = mock_zephyr_networking.NewMockAccessControlPolicyClient(ctrl)
		mockResourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		mockAppmeshClient = mock_appmesh.NewMockAppmeshClient(ctrl)
		dao = translation.NewAppmeshAccessControlDao(
			mockMeshServiceClient,
			mockMeshWorkloadClient,
			mockResourceSelector,
			func(mesh *zephyr_discovery.Mesh) (appmesh.AppmeshClient, error) {
				return mockAppmeshClient, nil
			},
			mockAcpClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectListMeshWorkloads = func() []*zephyr_discovery.MeshWorkload {
		mw1 := &zephyr_discovery.MeshWorkload{
			ObjectMeta: v1.ObjectMeta{
				Name: "mw1",
			},
			Spec: types.MeshWorkloadSpec{
				Mesh: &types2.ResourceRef{
					Name:      mesh.GetName(),
					Namespace: mesh.GetNamespace(),
				},
				KubeController: &types.MeshWorkloadSpec_KubeController{
					Labels: map[string]string{
						"k1": "v1",
					},
				},
			},
		}
		mw2 := &zephyr_discovery.MeshWorkload{
			ObjectMeta: v1.ObjectMeta{
				Name: "mw2",
			},
			Spec: types.MeshWorkloadSpec{
				Mesh: &types2.ResourceRef{
					Name:      mesh.GetName(),
					Namespace: mesh.GetNamespace(),
				},
				KubeController: &types.MeshWorkloadSpec_KubeController{
					Labels: map[string]string{
						"k2": "v2",
					},
				},
			},
		}
		mw3 := &zephyr_discovery.MeshWorkload{
			ObjectMeta: v1.ObjectMeta{
				Name: "mw3",
			},
			Spec: types.MeshWorkloadSpec{
				Mesh: &types2.ResourceRef{
					Name:      mesh.GetName(),
					Namespace: mesh.GetNamespace(),
				},
			},
		}
		mw4 := &zephyr_discovery.MeshWorkload{}
		mockMeshWorkloadClient.
			EXPECT().
			ListMeshWorkload(ctx).
			Return(&zephyr_discovery.MeshWorkloadList{
				Items: []zephyr_discovery.MeshWorkload{*mw1, *mw2, *mw3, *mw4},
			}, nil)

		return []*zephyr_discovery.MeshWorkload{mw1, mw2, mw3}
	}

	var expectListMeshServices = func() []*zephyr_discovery.MeshService {
		ms1 := &zephyr_discovery.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name: "ms1",
			},
			Spec: types.MeshServiceSpec{
				Mesh: &types2.ResourceRef{
					Name:      mesh.GetName(),
					Namespace: mesh.GetNamespace(),
				},
				KubeService: &types.MeshServiceSpec_KubeService{
					WorkloadSelectorLabels: map[string]string{
						"k1": "v1",
					},
				},
			},
		}
		ms2 := &zephyr_discovery.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name: "ms2",
			},
			Spec: types.MeshServiceSpec{
				Mesh: &types2.ResourceRef{
					Name:      mesh.GetName(),
					Namespace: mesh.GetNamespace(),
				},
				KubeService: &types.MeshServiceSpec_KubeService{
					WorkloadSelectorLabels: map[string]string{
						"k2": "v2",
					},
				},
			},
		}
		ms3 := &zephyr_discovery.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name: "ms3",
			},
			Spec: types.MeshServiceSpec{
				Mesh: &types2.ResourceRef{
					Name:      mesh.GetName(),
					Namespace: mesh.GetNamespace(),
				},
			},
		}
		ms4 := &zephyr_discovery.MeshService{}
		mockMeshServiceClient.
			EXPECT().
			ListMeshService(ctx).
			Return(&zephyr_discovery.MeshServiceList{
				Items: []zephyr_discovery.MeshService{*ms1, *ms2, *ms3, *ms4},
			}, nil)
		return []*zephyr_discovery.MeshService{ms1, ms2, ms3}
	}

	var expectListMeshWorkloadsAndServicesForMesh = func() ([]*zephyr_discovery.MeshWorkload, []*zephyr_discovery.MeshService) {
		return expectListMeshWorkloads(), expectListMeshServices()
	}

	It("should GetAllServiceWorkloadPairsForMesh", func() {
		workloads, services := expectListMeshWorkloadsAndServicesForMesh()
		// Expect map is defined below, gomega isn't smart enough to do equality checking on complex maps.
		//expectedServiceToWorkloads := map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload{
		//	services[0]: {workloads[0]},
		//	services[1]: {workloads[1]},
		//	services[2]: nil,
		//}
		//expectedWorkloadsToServices := map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService{
		//	workloads[0]: {services[0]},
		//	workloads[1]: {services[1]},
		//	workloads[2]: nil,
		//}
		servicesToWorkloads, workloadsToServices, err := dao.GetAllServiceWorkloadPairsForMesh(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(servicesToWorkloads).To(HaveKeyWithValue(services[0], []*zephyr_discovery.MeshWorkload{workloads[0]}))
		Expect(servicesToWorkloads).To(HaveKeyWithValue(services[1], []*zephyr_discovery.MeshWorkload{workloads[1]}))
		Expect(servicesToWorkloads).To(HaveKeyWithValue(services[2], BeNil()))

		Expect(workloadsToServices).To(HaveKeyWithValue(workloads[0], []*zephyr_discovery.MeshService{services[0]}))
		Expect(workloadsToServices).To(HaveKeyWithValue(workloads[1], []*zephyr_discovery.MeshService{services[1]}))
		Expect(workloadsToServices).To(HaveKeyWithValue(workloads[2], BeNil()))
	})

	It("should GetWorkloadsToAllUpstreamServices", func() {
		workloads, services := expectListMeshWorkloadsAndServicesForMesh()
		workloadsToAllUpstreamServices, err := dao.GetWorkloadsToAllUpstreamServices(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadsToAllUpstreamServices).To(HaveKey(workloads[0]))
		Expect(workloadsToAllUpstreamServices).To(HaveKey(workloads[1]))
		Expect(workloadsToAllUpstreamServices).To(HaveKey(workloads[2]))

		expectedSet := sets.NewMeshServiceSet(services[0], services[1], services[2])
		Expect(workloadsToAllUpstreamServices).To(HaveKeyWithValue(workloads[0], sets.MatchMeshServiceSet(expectedSet)))
		Expect(workloadsToAllUpstreamServices).To(HaveKeyWithValue(workloads[1], sets.MatchMeshServiceSet(expectedSet)))
		Expect(workloadsToAllUpstreamServices).To(HaveKeyWithValue(workloads[2], sets.MatchMeshServiceSet(expectedSet)))
	})

	It("should GetServicesWithACP", func() {
		services := expectListMeshServices()
		acp1 := &v1alpha1.AccessControlPolicy{
			Spec: types3.AccessControlPolicySpec{
				DestinationSelector: &types2.ServiceSelector{},
			},
		}
		acp2 := &v1alpha1.AccessControlPolicy{
			Spec: types3.AccessControlPolicySpec{
				DestinationSelector: &types2.ServiceSelector{},
			},
		}
		mockAcpClient.
			EXPECT().
			ListAccessControlPolicy(ctx).
			Return(&v1alpha1.AccessControlPolicyList{Items: []v1alpha1.AccessControlPolicy{*acp1, *acp2}}, nil)
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, acp1.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{{ObjectMeta: v1.ObjectMeta{Name: "other-service"}}, services[0]}, nil)
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, acp1.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{services[1]}, nil)
		expectedMeshServices := sets.NewMeshServiceSet(services[0], services[1])
		acpServicesInMesh, err := dao.GetServicesWithACP(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(acpServicesInMesh).To(sets.MatchMeshServiceSet(expectedMeshServices))
	})

	It("should GetWorkloadsToUpstreamServicesWithACP", func() {
		services := []*zephyr_discovery.MeshService{
			{ObjectMeta: v1.ObjectMeta{Name: "s1"}},
			{ObjectMeta: v1.ObjectMeta{Name: "s2"}},
			{ObjectMeta: v1.ObjectMeta{Name: "s3"}},
		}
		workloads := expectListMeshWorkloads()
		acp1 := &v1alpha1.AccessControlPolicy{
			Spec: types3.AccessControlPolicySpec{
				SourceSelector:      &types2.IdentitySelector{},
				DestinationSelector: &types2.ServiceSelector{},
			},
		}
		acp2 := &v1alpha1.AccessControlPolicy{
			Spec: types3.AccessControlPolicySpec{
				SourceSelector:      &types2.IdentitySelector{},
				DestinationSelector: &types2.ServiceSelector{},
			},
		}
		mockAcpClient.
			EXPECT().
			ListAccessControlPolicy(ctx).
			Return(&v1alpha1.AccessControlPolicyList{Items: []v1alpha1.AccessControlPolicy{*acp1, *acp2}}, nil)
		// workloads[0] -> services[0], services[1], services[2]
		// workloads[1] -> services[1], services[2]
		mockResourceSelector.
			EXPECT().
			GetMeshWorkloadsByIdentitySelector(ctx, acp1.Spec.GetSourceSelector()).
			Return([]*zephyr_discovery.MeshWorkload{{ObjectMeta: v1.ObjectMeta{Name: "other-workload"}}, workloads[0]}, nil)
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, acp1.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{services[0]}, nil)
		mockResourceSelector.
			EXPECT().
			GetMeshWorkloadsByIdentitySelector(ctx, acp2.Spec.GetSourceSelector()).
			Return([]*zephyr_discovery.MeshWorkload{workloads[0], workloads[1]}, nil)
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, acp2.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{services[1], services[2]}, nil)
		expectedDeclaredWorkloads := sets.NewMeshWorkloadSet(workloads[0], workloads[1])
		declaredWorkloads, workloadsToUpstreamServices, err := dao.GetWorkloadsToUpstreamServicesWithACP(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(declaredWorkloads).To(sets.MatchMeshWorkloadSet(expectedDeclaredWorkloads))
		Expect(workloadsToUpstreamServices).To(
			HaveKeyWithValue(workloads[0], sets.MatchMeshServiceSet(sets.NewMeshServiceSet(services[0], services[1], services[2]))))
		Expect(workloadsToUpstreamServices).To(
			HaveKeyWithValue(workloads[1], sets.MatchMeshServiceSet(sets.NewMeshServiceSet(services[1], services[2]))))
	})
})
