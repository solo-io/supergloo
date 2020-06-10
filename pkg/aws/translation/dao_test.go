package translation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	types3 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws/clients"
	mock_appmesh "github.com/solo-io/service-mesh-hub/pkg/aws/clients/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/aws/translation"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/kube/selection/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Dao", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		mockMeshServiceClient  *mock_core.MockMeshServiceClient
		mockMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockAcpClient          *mock_smh_networking.MockAccessControlPolicyClient
		mockResourceSelector   *mock_selector.MockResourceSelector
		mockAppmeshClient      *mock_appmesh.MockAppmeshClient
		dao                    translation.AppmeshTranslationDao
		mesh                   = &smh_discovery.Mesh{
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
		mockAcpClient = mock_smh_networking.NewMockAccessControlPolicyClient(ctrl)
		mockResourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		mockAppmeshClient = mock_appmesh.NewMockAppmeshClient(ctrl)
		dao = translation.NewAppmeshAccessControlDao(
			mockMeshServiceClient,
			mockMeshWorkloadClient,
			mockResourceSelector,
			func(mesh *smh_discovery.Mesh) (clients.AppmeshClient, error) {
				return mockAppmeshClient, nil
			},
			mockAcpClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectListMeshWorkloads = func() []*smh_discovery.MeshWorkload {
		mw1 := &smh_discovery.MeshWorkload{
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
		mw2 := &smh_discovery.MeshWorkload{
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
		mw3 := &smh_discovery.MeshWorkload{
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
		mw4 := &smh_discovery.MeshWorkload{}
		mockMeshWorkloadClient.
			EXPECT().
			ListMeshWorkload(ctx).
			Return(&smh_discovery.MeshWorkloadList{
				Items: []smh_discovery.MeshWorkload{*mw1, *mw2, *mw3, *mw4},
			}, nil)

		return []*smh_discovery.MeshWorkload{mw1, mw2, mw3}
	}

	var expectListMeshServices = func() []*smh_discovery.MeshService {
		ms1 := &smh_discovery.MeshService{
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
		ms2 := &smh_discovery.MeshService{
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
		ms3 := &smh_discovery.MeshService{
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
		ms4 := &smh_discovery.MeshService{}
		mockMeshServiceClient.
			EXPECT().
			ListMeshService(ctx).
			Return(&smh_discovery.MeshServiceList{
				Items: []smh_discovery.MeshService{*ms1, *ms2, *ms3, *ms4},
			}, nil)
		return []*smh_discovery.MeshService{ms1, ms2, ms3}
	}

	var expectListMeshWorkloadsAndServicesForMesh = func() ([]*smh_discovery.MeshWorkload, []*smh_discovery.MeshService) {
		return expectListMeshWorkloads(), expectListMeshServices()
	}

	It("should GetAllServiceWorkloadPairsForMesh", func() {
		workloads, services := expectListMeshWorkloadsAndServicesForMesh()
		// Expect map is defined below, gomega isn't smart enough to do equality checking on complex maps.
		//expectedServiceToWorkloads := map[*smh_discovery.MeshService][]*smh_discovery.MeshWorkload{
		//	services[0]: {workloads[0]},
		//	services[1]: {workloads[1]},
		//	services[2]: nil,
		//}
		//expectedWorkloadsToServices := map[*smh_discovery.MeshWorkload][]*smh_discovery.MeshService{
		//	workloads[0]: {services[0]},
		//	workloads[1]: {services[1]},
		//	workloads[2]: nil,
		//}
		servicesToWorkloads, workloadsToServices, err := dao.GetAllServiceWorkloadPairsForMesh(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(servicesToWorkloads).To(HaveKeyWithValue(services[0], []*smh_discovery.MeshWorkload{workloads[0]}))
		Expect(servicesToWorkloads).To(HaveKeyWithValue(services[1], []*smh_discovery.MeshWorkload{workloads[1]}))
		Expect(servicesToWorkloads).To(HaveKeyWithValue(services[2], BeNil()))

		Expect(workloadsToServices).To(HaveKeyWithValue(workloads[0], []*smh_discovery.MeshService{services[0]}))
		Expect(workloadsToServices).To(HaveKeyWithValue(workloads[1], []*smh_discovery.MeshService{services[1]}))
		Expect(workloadsToServices).To(HaveKeyWithValue(workloads[2], BeNil()))
	})

	It("should GetWorkloadsToAllUpstreamServices", func() {
		workloads, services := expectListMeshWorkloadsAndServicesForMesh()
		workloadsToAllUpstreamServices, err := dao.GetWorkloadsToAllUpstreamServices(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadsToAllUpstreamServices).To(HaveKey(selection.ToUniqueSingleClusterString(workloads[0].ObjectMeta)))
		Expect(workloadsToAllUpstreamServices).To(HaveKey(selection.ToUniqueSingleClusterString(workloads[1].ObjectMeta)))
		Expect(workloadsToAllUpstreamServices).To(HaveKey(selection.ToUniqueSingleClusterString(workloads[2].ObjectMeta)))

		expectedSet := smh_discovery_sets.NewMeshServiceSet(services[0], services[1], services[2])

		Expect(workloadsToAllUpstreamServices).To(HaveKeyWithValue(selection.ToUniqueSingleClusterString(workloads[0].ObjectMeta), Equal(expectedSet)))
		Expect(workloadsToAllUpstreamServices).To(HaveKeyWithValue(selection.ToUniqueSingleClusterString(workloads[1].ObjectMeta), Equal(expectedSet)))
		Expect(workloadsToAllUpstreamServices).To(HaveKeyWithValue(selection.ToUniqueSingleClusterString(workloads[2].ObjectMeta), Equal(expectedSet)))
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
			Return([]*smh_discovery.MeshService{{ObjectMeta: v1.ObjectMeta{Name: "other-service"}}, services[0]}, nil)
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, acp1.Spec.GetDestinationSelector()).
			Return([]*smh_discovery.MeshService{services[1]}, nil)
		expectedMeshServices := smh_discovery_sets.NewMeshServiceSet(services[0], services[1])
		acpServicesInMesh, err := dao.GetServicesWithACP(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(acpServicesInMesh).To(Equal(expectedMeshServices))
	})

	It("should GetWorkloadsToUpstreamServicesWithACP", func() {
		services := []*smh_discovery.MeshService{
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
			Return([]*smh_discovery.MeshWorkload{{ObjectMeta: v1.ObjectMeta{Name: "other-workload"}}, workloads[0]}, nil)
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, acp1.Spec.GetDestinationSelector()).
			Return([]*smh_discovery.MeshService{services[0]}, nil)
		mockResourceSelector.
			EXPECT().
			GetMeshWorkloadsByIdentitySelector(ctx, acp2.Spec.GetSourceSelector()).
			Return([]*smh_discovery.MeshWorkload{workloads[0], workloads[1]}, nil)
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, acp2.Spec.GetDestinationSelector()).
			Return([]*smh_discovery.MeshService{services[1], services[2]}, nil)
		expectedDeclaredWorkloads := smh_discovery_sets.NewMeshWorkloadSet(workloads[0], workloads[1])
		declaredWorkloads, workloadsToUpstreamServices, err := dao.GetWorkloadsToUpstreamServicesWithACP(ctx, mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(declaredWorkloads).To(Equal(expectedDeclaredWorkloads))
		Expect(workloadsToUpstreamServices).To(
			HaveKeyWithValue(selection.ToUniqueSingleClusterString(workloads[0].ObjectMeta),
				Equal(smh_discovery_sets.NewMeshServiceSet(services[0], services[1], services[2]))))
		Expect(workloadsToUpstreamServices).To(
			HaveKeyWithValue(selection.ToUniqueSingleClusterString(workloads[1].ObjectMeta),
				Equal(smh_discovery_sets.NewMeshServiceSet(services[1], services[2]))))
	})
})
