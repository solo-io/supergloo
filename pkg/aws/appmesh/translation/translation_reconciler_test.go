package translation_test

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws/appmesh/translation"
	mock_translation "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh/translation/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("TranslationReconciler", func() {
	var (
		ctrl                         *gomock.Controller
		ctx                          context.Context
		mockAppmeshTranslator        *mock_translation.MockAppmeshTranslator
		mockDao                      *mock_translation.MockAppmeshTranslationDao
		appmeshTranslationReconciler translation.AppmeshTranslationReconciler
		mesh                         = &zephyr_discovery.Mesh{
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
		mockAppmeshTranslator = mock_translation.NewMockAppmeshTranslator(ctrl)
		mockDao = mock_translation.NewMockAppmeshTranslationDao(ctrl)
		appmeshTranslationReconciler = translation.NewAppmeshTranslationReconciler(mockAppmeshTranslator, mockDao)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should reconcile with global access control enabled", func() {
		meshService1 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service1"}}
		meshService2 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service2"}}
		meshService3 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service3"}}
		meshService4 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service4"}}
		meshService5 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service5"}}
		meshService6 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service6"}}

		meshWorkload1 := &zephyr_discovery.MeshWorkload{ObjectMeta: v1.ObjectMeta{Name: "workload1"}}
		meshWorkload2 := &zephyr_discovery.MeshWorkload{ObjectMeta: v1.ObjectMeta{Name: "workload2"}}
		meshWorkload3 := &zephyr_discovery.MeshWorkload{ObjectMeta: v1.ObjectMeta{Name: "workload3"}}
		servicesToBackingWorkloads := map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload{
			meshService1: {
				{ObjectMeta: v1.ObjectMeta{Name: "service1-workload1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "service1-workload2"}},
			},
			meshService2: {
				{ObjectMeta: v1.ObjectMeta{Name: "service2-workload2"}},
			},
			meshService3: {
				{ObjectMeta: v1.ObjectMeta{Name: "service3-workload1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "service3-workload3"}},
			},
		}
		workloadsToBackingServices := map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService{
			meshWorkload1: {
				{ObjectMeta: v1.ObjectMeta{Name: "workload1-service1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "workload1-service3"}},
			},
			meshWorkload2: {
				{ObjectMeta: v1.ObjectMeta{Name: "workload2-service1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "workload2-service2"}},
			},
			meshWorkload3: {
				{ObjectMeta: v1.ObjectMeta{Name: "workload3-service3"}},
			},
		}
		servicesWithACP := zephyr_discovery_sets.NewMeshServiceSet(meshService1, meshService2)
		workloadsWithACP := zephyr_discovery_sets.NewMeshWorkloadSet(meshWorkload1, meshWorkload3)
		workloadsToUpstreamServices := map[string]zephyr_discovery_sets.MeshServiceSet{
			selection.ToUniqueSingleClusterString(meshWorkload1.ObjectMeta): zephyr_discovery_sets.NewMeshServiceSet(meshService4),
			selection.ToUniqueSingleClusterString(meshWorkload3.ObjectMeta): zephyr_discovery_sets.NewMeshServiceSet(meshService5, meshService6),
			selection.ToUniqueSingleClusterString(meshWorkload2.ObjectMeta): zephyr_discovery_sets.NewMeshServiceSet(meshService1), // excluded
		}
		mockDao.EXPECT().GetAllServiceWorkloadPairsForMesh(ctx, mesh).Return(servicesToBackingWorkloads, workloadsToBackingServices, nil)
		mockDao.EXPECT().GetServicesWithACP(ctx, mesh).Return(servicesWithACP, nil)
		mockDao.EXPECT().GetWorkloadsToUpstreamServicesWithACP(ctx, mesh).Return(workloadsWithACP, workloadsToUpstreamServices, nil)

		appmeshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
		vs1 := &appmesh2.VirtualServiceData{}
		vs2 := &appmesh2.VirtualServiceData{}
		vr1 := &appmesh2.VirtualRouterData{}
		vr2 := &appmesh2.VirtualRouterData{}
		r1 := &appmesh2.RouteData{}
		r2 := &appmesh2.RouteData{}
		vn1 := &appmesh2.VirtualNodeData{}
		vn3 := &appmesh2.VirtualNodeData{}
		mockAppmeshTranslator.EXPECT().BuildVirtualService(appmeshName, meshService1).Return(vs1)
		mockAppmeshTranslator.EXPECT().BuildVirtualService(appmeshName, meshService2).Return(vs2)
		mockAppmeshTranslator.EXPECT().BuildVirtualRouter(appmeshName, meshService1).Return(vr1)
		mockAppmeshTranslator.EXPECT().BuildVirtualRouter(appmeshName, meshService2).Return(vr2)
		mockAppmeshTranslator.EXPECT().
			BuildRoute(appmeshName,
				translation.DefaultRouteName,
				translation.DefaultRoutePriority,
				meshService1,
				servicesToBackingWorkloads[meshService1]).
			Return(r1, nil)
		mockAppmeshTranslator.EXPECT().
			BuildRoute(appmeshName,
				translation.DefaultRouteName,
				translation.DefaultRoutePriority,
				meshService2,
				servicesToBackingWorkloads[meshService2]).
			Return(r2, nil)
		mockAppmeshTranslator.
			EXPECT().
			BuildVirtualNode(
				appmeshName,
				meshWorkload1,
				workloadsToBackingServices[meshWorkload1][0],
				workloadsToUpstreamServices[selection.ToUniqueSingleClusterString(meshWorkload1.ObjectMeta)].List()).
			Return(vn1)
		mockAppmeshTranslator.
			EXPECT().
			BuildVirtualNode(
				appmeshName,
				meshWorkload3,
				workloadsToBackingServices[meshWorkload3][0],
				workloadsToUpstreamServices[selection.ToUniqueSingleClusterString(meshWorkload3.ObjectMeta)].List()).
			Return(vn3)
		mockDao.EXPECT().ReconcileVirtualRouters(ctx, mesh, []*appmesh2.VirtualRouterData{vr1, vr2}).Return(nil)
		mockDao.EXPECT().ReconcileVirtualServices(ctx, mesh, []*appmesh2.VirtualServiceData{vs1, vs2}).Return(nil)
		mockDao.EXPECT().ReconcileRoutes(ctx, mesh, []*appmesh2.RouteData{r1, r2}).Return(nil)
		mockDao.EXPECT().ReconcileVirtualNodes(ctx, mesh, []*appmesh2.VirtualNodeData{vn1, vn3}).Return(nil)

		err := appmeshTranslationReconciler.Reconcile(ctx, mesh, &v1alpha1.VirtualMesh{
			Spec: types2.VirtualMeshSpec{EnforceAccessControl: types2.VirtualMeshSpec_ENABLED}})
		Expect(err).To(BeNil())
	})

	It("should reconcile with global access control disabled", func() {
		meshService1 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service1"}}
		meshService2 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service2"}}
		meshService3 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service3"}}
		meshService4 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service4"}}
		meshService5 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service5"}}
		meshService6 := &zephyr_discovery.MeshService{ObjectMeta: v1.ObjectMeta{Name: "service6"}}

		meshWorkload1 := &zephyr_discovery.MeshWorkload{ObjectMeta: v1.ObjectMeta{Name: "workload1"}}
		meshWorkload2 := &zephyr_discovery.MeshWorkload{ObjectMeta: v1.ObjectMeta{Name: "workload2"}}
		meshWorkload3 := &zephyr_discovery.MeshWorkload{ObjectMeta: v1.ObjectMeta{Name: "workload3"}}
		servicesToBackingWorkloads := map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload{
			meshService1: {
				{ObjectMeta: v1.ObjectMeta{Name: "service1-workload1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "service1-workload2"}},
			},
			meshService2: {
				{ObjectMeta: v1.ObjectMeta{Name: "service2-workload2"}},
			},
			meshService3: {
				{ObjectMeta: v1.ObjectMeta{Name: "service3-workload1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "service3-workload3"}},
			},
		}
		workloadsToBackingServices := map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService{
			meshWorkload1: {
				{ObjectMeta: v1.ObjectMeta{Name: "workload1-service1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "workload1-service3"}},
			},
			meshWorkload2: {
				{ObjectMeta: v1.ObjectMeta{Name: "workload2-service1"}},
				{ObjectMeta: v1.ObjectMeta{Name: "workload2-service2"}},
			},
			meshWorkload3: {
				{ObjectMeta: v1.ObjectMeta{Name: "workload3-service3"}},
			},
		}
		workloadsToUpstreamServices := map[string]zephyr_discovery_sets.MeshServiceSet{
			selection.ToUniqueSingleClusterString(meshWorkload1.ObjectMeta): zephyr_discovery_sets.NewMeshServiceSet(meshService4),
			selection.ToUniqueSingleClusterString(meshWorkload3.ObjectMeta): zephyr_discovery_sets.NewMeshServiceSet(meshService5, meshService6),
		}
		mockDao.EXPECT().GetAllServiceWorkloadPairsForMesh(ctx, mesh).Return(servicesToBackingWorkloads, workloadsToBackingServices, nil)
		mockDao.EXPECT().GetWorkloadsToAllUpstreamServices(ctx, mesh).Return(workloadsToUpstreamServices, nil)

		appmeshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
		vs1 := &appmesh2.VirtualServiceData{}
		vs2 := &appmesh2.VirtualServiceData{}
		vs3 := &appmesh2.VirtualServiceData{}
		vr1 := &appmesh2.VirtualRouterData{}
		vr2 := &appmesh2.VirtualRouterData{}
		vr3 := &appmesh2.VirtualRouterData{}
		r1 := &appmesh2.RouteData{}
		r2 := &appmesh2.RouteData{}
		r3 := &appmesh2.RouteData{}
		vn1 := &appmesh2.VirtualNodeData{}
		vn2 := &appmesh2.VirtualNodeData{}
		vn3 := &appmesh2.VirtualNodeData{}
		mockAppmeshTranslator.EXPECT().BuildVirtualService(appmeshName, meshService1).Return(vs1)
		mockAppmeshTranslator.EXPECT().BuildVirtualService(appmeshName, meshService2).Return(vs2)
		mockAppmeshTranslator.EXPECT().BuildVirtualService(appmeshName, meshService3).Return(vs3)
		mockAppmeshTranslator.EXPECT().BuildVirtualRouter(appmeshName, meshService1).Return(vr1)
		mockAppmeshTranslator.EXPECT().BuildVirtualRouter(appmeshName, meshService2).Return(vr2)
		mockAppmeshTranslator.EXPECT().BuildVirtualRouter(appmeshName, meshService3).Return(vr3)
		mockAppmeshTranslator.EXPECT().
			BuildRoute(appmeshName,
				translation.DefaultRouteName,
				translation.DefaultRoutePriority,
				meshService1,
				servicesToBackingWorkloads[meshService1]).
			Return(r1, nil)
		mockAppmeshTranslator.EXPECT().
			BuildRoute(appmeshName,
				translation.DefaultRouteName,
				translation.DefaultRoutePriority,
				meshService2,
				servicesToBackingWorkloads[meshService2]).
			Return(r2, nil)
		mockAppmeshTranslator.EXPECT().
			BuildRoute(appmeshName,
				translation.DefaultRouteName,
				translation.DefaultRoutePriority,
				meshService3,
				servicesToBackingWorkloads[meshService3]).
			Return(r3, nil)
		mockAppmeshTranslator.
			EXPECT().
			BuildVirtualNode(
				appmeshName,
				meshWorkload1,
				workloadsToBackingServices[meshWorkload1][0],
				workloadsToUpstreamServices[selection.ToUniqueSingleClusterString(meshWorkload1.ObjectMeta)].List()).
			Return(vn1)
		mockAppmeshTranslator.
			EXPECT().
			BuildVirtualNode(
				appmeshName,
				meshWorkload2,
				workloadsToBackingServices[meshWorkload2][0],
				nil).
			Return(vn2)
		mockAppmeshTranslator.
			EXPECT().
			BuildVirtualNode(
				appmeshName,
				meshWorkload3,
				workloadsToBackingServices[meshWorkload3][0],
				workloadsToUpstreamServices[selection.ToUniqueSingleClusterString(meshWorkload3.ObjectMeta)].List()).
			Return(vn3)
		mockDao.EXPECT().ReconcileVirtualRouters(ctx, mesh, []*appmesh2.VirtualRouterData{vr1, vr2, vr3}).Return(nil)
		mockDao.EXPECT().ReconcileVirtualServices(ctx, mesh, []*appmesh2.VirtualServiceData{vs1, vs2, vs3}).Return(nil)
		mockDao.EXPECT().ReconcileRoutes(ctx, mesh, []*appmesh2.RouteData{r1, r2, r3}).Return(nil)
		mockDao.EXPECT().ReconcileVirtualNodes(ctx, mesh, []*appmesh2.VirtualNodeData{vn1, vn2, vn3}).Return(nil)

		err := appmeshTranslationReconciler.Reconcile(ctx, mesh, &v1alpha1.VirtualMesh{
			Spec: types2.VirtualMeshSpec{EnforceAccessControl: types2.VirtualMeshSpec_DISABLED}})
		Expect(err).To(BeNil())
	})
})
