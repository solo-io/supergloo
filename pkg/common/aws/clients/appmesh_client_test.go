package clients_test

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	appmesh2 "github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	mock_matcher "github.com/solo-io/service-mesh-hub/pkg/common/aws/matcher/mocks"
	mock_appmesh_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/aws/appmesh"
)

var _ = Describe("AppmeshClient", func() {
	var (
		ctrl                 *gomock.Controller
		ctx                  context.Context
		mockAppmeshMatcher   *mock_matcher.MockAppmeshMatcher
		mockAppmeshRawClient *mock_appmesh_clients.MockAppMeshAPI
		appmeshClient        appmesh2.AppmeshClient
		meshName             = aws2.String("mesh-name")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockAppmeshMatcher = mock_matcher.NewMockAppmeshMatcher(ctrl)
		mockAppmeshRawClient = mock_appmesh_clients.NewMockAppMeshAPI(ctrl)
		appmeshClient = appmesh2.NewAppmeshClient(mockAppmeshRawClient, mockAppmeshMatcher)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var awsNotFoundErr = func() error {
		return awserr.New(appmesh.ErrCodeNotFoundException, "", nil)
	}

	var expectVirtualServiceCreate = func(vs *appmesh.VirtualServiceData) {
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
				MeshName:           vs.MeshName,
				VirtualServiceName: vs.VirtualServiceName,
			}).
			Return(nil, awsNotFoundErr())
		mockAppmeshRawClient.
			EXPECT().
			CreateVirtualService(&appmesh.CreateVirtualServiceInput{
				MeshName:           vs.MeshName,
				VirtualServiceName: vs.VirtualServiceName,
				Spec:               vs.Spec,
			}).
			Return(nil, nil)
	}

	var expectVirtualRouterCreate = func(vr *appmesh.VirtualRouterData) {
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
				MeshName:          vr.MeshName,
				VirtualRouterName: vr.VirtualRouterName,
			}).
			Return(nil, awsNotFoundErr())
		mockAppmeshRawClient.
			EXPECT().
			CreateVirtualRouter(&appmesh.CreateVirtualRouterInput{
				MeshName:          vr.MeshName,
				VirtualRouterName: vr.VirtualRouterName,
				Spec:              vr.Spec,
			}).
			Return(nil, nil)
	}

	var expectRouteCreate = func(r *appmesh.RouteData) {
		mockAppmeshRawClient.
			EXPECT().
			DescribeRoute(&appmesh.DescribeRouteInput{
				MeshName:          r.MeshName,
				VirtualRouterName: r.VirtualRouterName,
				RouteName:         r.RouteName,
			}).
			Return(nil, awsNotFoundErr())
		mockAppmeshRawClient.
			EXPECT().
			CreateRoute(&appmesh.CreateRouteInput{
				MeshName:          r.MeshName,
				VirtualRouterName: r.VirtualRouterName,
				RouteName:         r.RouteName,
				Spec:              r.Spec,
			}).
			Return(nil, nil)
	}

	var expectVirtualNodeCreate = func(r *appmesh.VirtualNodeData) {
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
				MeshName:        r.MeshName,
				VirtualNodeName: r.VirtualNodeName,
			}).
			Return(nil, awsNotFoundErr())
		mockAppmeshRawClient.
			EXPECT().
			CreateVirtualNode(&appmesh.CreateVirtualNodeInput{
				MeshName:        r.MeshName,
				VirtualNodeName: r.VirtualNodeName,
				Spec:            r.Spec,
			}).
			Return(nil, nil)
	}

	It("EnsureVirtualService should create if not exist", func() {
		vs := &appmesh.VirtualServiceData{
			MeshName:           meshName,
			VirtualServiceName: aws2.String("vs-name"),
			Spec: &appmesh.VirtualServiceSpec{
				Provider: &appmesh.VirtualServiceProvider{},
			},
		}
		expectVirtualServiceCreate(vs)
		err := appmeshClient.EnsureVirtualService(vs)
		Expect(err).ToNot(HaveOccurred())
	})

	It("EnsureVirtualService should return error", func() {
		vs := &appmesh.VirtualServiceData{
			MeshName:           meshName,
			VirtualServiceName: aws2.String("vs-name"),
			Spec: &appmesh.VirtualServiceSpec{
				Provider: &appmesh.VirtualServiceProvider{},
			},
		}
		testErr := eris.New("test-err")
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
				MeshName:           vs.MeshName,
				VirtualServiceName: vs.VirtualServiceName,
			}).
			Return(nil, testErr)
		err := appmeshClient.EnsureVirtualService(vs)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})

	It("EnsureVirtualService should update if exists but not equal", func() {
		vs := &appmesh.VirtualServiceData{
			MeshName:           meshName,
			VirtualServiceName: aws2.String("vs-name"),
			Spec: &appmesh.VirtualServiceSpec{
				Provider: &appmesh.VirtualServiceProvider{},
			},
		}
		existingVs := &appmesh.VirtualServiceData{}
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
				MeshName:           vs.MeshName,
				VirtualServiceName: vs.VirtualServiceName,
			}).
			Return(&appmesh.DescribeVirtualServiceOutput{
				VirtualService: existingVs,
			}, nil)
		mockAppmeshMatcher.EXPECT().AreVirtualServicesEqual(existingVs, vs).Return(false)
		mockAppmeshRawClient.
			EXPECT().
			UpdateVirtualService(&appmesh.UpdateVirtualServiceInput{
				MeshName:           vs.MeshName,
				VirtualServiceName: vs.VirtualServiceName,
				Spec:               vs.Spec,
			}).
			Return(nil, nil)
		err := appmeshClient.EnsureVirtualService(vs)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should ReconcileVirtualRoutersAndRoutesAndVirtualServices", func() {
		declaredVr := []*appmesh.VirtualRouterData{
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-1"),
			}, {
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-2"),
			}, {
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-3"),
			},
			{
				MeshName:          aws2.String("some-other-mesh"),
				VirtualRouterName: aws2.String("vr-name-3"),
			},
		}
		for _, vr := range declaredVr[:3] {
			expectVirtualRouterCreate(vr)
		}
		declaredRoutes := []*appmesh.RouteData{
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-1"),
				RouteName:         aws2.String("r-name-1"),
			}, {
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-2"),
				RouteName:         aws2.String("r-name-2"),
			}, {
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-3"),
				RouteName:         aws2.String("r-name-3a"),
			},
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-3"),
				RouteName:         aws2.String("r-name-3b"),
			},
			{
				MeshName:          aws2.String("some-other-mesh"),
				VirtualRouterName: aws2.String("vr-name-3"),
				RouteName:         aws2.String("r-name-3"),
			},
		}
		for _, route := range declaredRoutes[:4] {
			expectRouteCreate(route)
		}
		declaredVs := []*appmesh.VirtualServiceData{
			{
				MeshName:           meshName,
				VirtualServiceName: aws2.String("vs-name-1"),
			}, {
				MeshName:           meshName,
				VirtualServiceName: aws2.String("vs-name-2"),
			}, {
				MeshName:           meshName,
				VirtualServiceName: aws2.String("vs-name-3"),
			},
			{
				MeshName:           aws2.String("some-other-mesh"),
				VirtualServiceName: aws2.String("vs-name-3"),
			},
		}
		for _, vs := range declaredVs[:3] {
			expectVirtualServiceCreate(vs)
		}
		existingVs := []*appmesh.VirtualServiceRef{
			{
				MeshName:           meshName,
				VirtualServiceName: aws2.String("vs-name-3"),
			},
			{
				MeshName:           meshName,
				VirtualServiceName: aws2.String("vs-name-4"),
			},
		}
		mockAppmeshRawClient.
			EXPECT().
			ListVirtualServicesPagesWithContext(ctx, &appmesh.ListVirtualServicesInput{
				MeshName: meshName,
			}, gomock.Any()).
			DoAndReturn(func(_ aws2.Context, vsReq *appmesh.ListVirtualServicesInput, callback func(*appmesh.ListVirtualServicesOutput, bool) bool) error {
				callback(&appmesh.ListVirtualServicesOutput{
					VirtualServices: existingVs,
				}, true)
				return nil
			})
		mockAppmeshRawClient.
			EXPECT().
			DeleteVirtualService(&appmesh.DeleteVirtualServiceInput{
				MeshName:           meshName,
				VirtualServiceName: existingVs[1].VirtualServiceName,
			}).
			Return(nil, nil)
		existingVr := []*appmesh.VirtualRouterRef{
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-3"),
			},
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-4"),
			},
		}
		mockAppmeshRawClient.
			EXPECT().
			ListVirtualRoutersPagesWithContext(ctx, &appmesh.ListVirtualRoutersInput{
				MeshName: meshName,
			}, gomock.Any()).
			DoAndReturn(func(_ aws2.Context, vsReq *appmesh.ListVirtualRoutersInput, callback func(*appmesh.ListVirtualRoutersOutput, bool) bool) error {
				callback(&appmesh.ListVirtualRoutersOutput{
					VirtualRouters: existingVr,
				}, true)
				return nil
			})
		existingRoutesForVr3 := []*appmesh.RouteRef{
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-3"),
				RouteName:         aws2.String("r-name-3a"),
			},
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-3"),
				RouteName:         aws2.String("r-name-4"),
			},
		}
		existingRoutesForVr4 := []*appmesh.RouteRef{
			{
				MeshName:          meshName,
				VirtualRouterName: aws2.String("vr-name-4"),
				RouteName:         aws2.String("r-name-4a"),
			},
		}
		mockAppmeshRawClient.
			EXPECT().
			ListRoutesPagesWithContext(ctx, &appmesh.ListRoutesInput{
				VirtualRouterName: existingVr[0].VirtualRouterName,
				MeshName:          meshName,
			}, gomock.Any()).
			DoAndReturn(func(_ aws2.Context, vsReq *appmesh.ListRoutesInput, callback func(*appmesh.ListRoutesOutput, bool) bool) error {
				callback(&appmesh.ListRoutesOutput{
					Routes: existingRoutesForVr3,
				}, true)
				return nil
			})
		mockAppmeshRawClient.
			EXPECT().
			ListRoutesPagesWithContext(ctx, &appmesh.ListRoutesInput{
				VirtualRouterName: existingVr[1].VirtualRouterName,
				MeshName:          meshName,
			}, gomock.Any()).
			DoAndReturn(func(_ aws2.Context, vsReq *appmesh.ListRoutesInput, callback func(*appmesh.ListRoutesOutput, bool) bool) error {
				callback(&appmesh.ListRoutesOutput{
					Routes: existingRoutesForVr4,
				}, true)
				return nil
			})
		mockAppmeshRawClient.
			EXPECT().
			DeleteRoute(&appmesh.DeleteRouteInput{
				MeshName:          meshName,
				VirtualRouterName: existingRoutesForVr3[1].VirtualRouterName,
				RouteName:         existingRoutesForVr3[1].RouteName,
			}).
			Return(nil, nil)
		mockAppmeshRawClient.
			EXPECT().
			DeleteRoute(&appmesh.DeleteRouteInput{
				MeshName:          meshName,
				VirtualRouterName: existingRoutesForVr4[0].VirtualRouterName,
				RouteName:         existingRoutesForVr4[0].RouteName,
			}).
			Return(nil, nil)
		mockAppmeshRawClient.
			EXPECT().
			DeleteVirtualRouter(&appmesh.DeleteVirtualRouterInput{
				MeshName:          meshName,
				VirtualRouterName: existingVr[1].VirtualRouterName,
			}).
			Return(nil, nil)
		err := appmeshClient.ReconcileVirtualRoutersAndRoutesAndVirtualServices(ctx, meshName, declaredVr, declaredRoutes, declaredVs)
		Expect(err).ToNot(HaveOccurred())
	})

	It("EnsureVirtualRouter should create if not exist", func() {
		vr := &appmesh.VirtualRouterData{
			MeshName:          meshName,
			VirtualRouterName: aws2.String("vr-name"),
			Spec:              &appmesh.VirtualRouterSpec{},
		}
		expectVirtualRouterCreate(vr)
		err := appmeshClient.EnsureVirtualRouter(vr)
		Expect(err).ToNot(HaveOccurred())
	})

	It("EnsureVirtualRouter should return error", func() {
		vr := &appmesh.VirtualRouterData{
			MeshName:          meshName,
			VirtualRouterName: aws2.String("vr-name"),
			Spec:              &appmesh.VirtualRouterSpec{},
		}
		testErr := eris.New("test-err")
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
				MeshName:          vr.MeshName,
				VirtualRouterName: vr.VirtualRouterName,
			}).
			Return(nil, testErr)
		err := appmeshClient.EnsureVirtualRouter(vr)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})

	It("EnsureVirtualRouter should update if exists but not equal", func() {
		vr := &appmesh.VirtualRouterData{
			MeshName:          meshName,
			VirtualRouterName: aws2.String("vr-name"),
			Spec:              &appmesh.VirtualRouterSpec{},
		}
		existingVr := &appmesh.VirtualRouterData{}
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
				MeshName:          vr.MeshName,
				VirtualRouterName: vr.VirtualRouterName,
			}).
			Return(&appmesh.DescribeVirtualRouterOutput{
				VirtualRouter: existingVr,
			}, nil)
		mockAppmeshMatcher.EXPECT().AreVirtualRoutersEqual(existingVr, vr).Return(false)
		mockAppmeshRawClient.
			EXPECT().
			UpdateVirtualRouter(&appmesh.UpdateVirtualRouterInput{
				MeshName:          vr.MeshName,
				VirtualRouterName: vr.VirtualRouterName,
				Spec:              vr.Spec,
			}).
			Return(nil, nil)
		err := appmeshClient.EnsureVirtualRouter(vr)
		Expect(err).ToNot(HaveOccurred())
	})

	It("EnsureRoute should create if not exist", func() {
		route := &appmesh.RouteData{
			MeshName:  meshName,
			RouteName: aws2.String("route-name"),
			Spec:      &appmesh.RouteSpec{},
		}
		expectRouteCreate(route)
		err := appmeshClient.EnsureRoute(route)
		Expect(err).ToNot(HaveOccurred())
	})

	It("EnsureRoute should return error", func() {
		routes := &appmesh.RouteData{
			MeshName:  meshName,
			RouteName: aws2.String("routes-name"),
			Spec:      &appmesh.RouteSpec{},
		}
		testErr := eris.New("test-err")
		mockAppmeshRawClient.
			EXPECT().
			DescribeRoute(&appmesh.DescribeRouteInput{
				MeshName:  routes.MeshName,
				RouteName: routes.RouteName,
			}).
			Return(nil, testErr)
		err := appmeshClient.EnsureRoute(routes)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})

	It("EnsureRoute should update if exists but not equal", func() {
		route := &appmesh.RouteData{
			MeshName:  meshName,
			RouteName: aws2.String("route-name"),
			Spec:      &appmesh.RouteSpec{},
		}
		existingVs := &appmesh.RouteData{}
		mockAppmeshRawClient.
			EXPECT().
			DescribeRoute(&appmesh.DescribeRouteInput{
				MeshName:  route.MeshName,
				RouteName: route.RouteName,
			}).
			Return(&appmesh.DescribeRouteOutput{
				Route: existingVs,
			}, nil)
		mockAppmeshMatcher.EXPECT().AreRoutesEqual(existingVs, route).Return(false)
		mockAppmeshRawClient.
			EXPECT().
			UpdateRoute(&appmesh.UpdateRouteInput{
				MeshName:  route.MeshName,
				RouteName: route.RouteName,
				Spec:      route.Spec,
			}).
			Return(nil, nil)
		err := appmeshClient.EnsureRoute(route)
		Expect(err).ToNot(HaveOccurred())
	})

	It("EnsureVirtualNode should create if not exist", func() {
		vn := &appmesh.VirtualNodeData{
			MeshName:        meshName,
			VirtualNodeName: aws2.String("vn-name"),
			Spec:            &appmesh.VirtualNodeSpec{},
		}
		expectVirtualNodeCreate(vn)
		err := appmeshClient.EnsureVirtualNode(vn)
		Expect(err).ToNot(HaveOccurred())
	})

	It("EnsureVirtualNode should return error", func() {
		vn := &appmesh.VirtualNodeData{
			MeshName:        meshName,
			VirtualNodeName: aws2.String("vn-name"),
			Spec:            &appmesh.VirtualNodeSpec{},
		}
		testErr := eris.New("test-err")
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
				MeshName:        vn.MeshName,
				VirtualNodeName: vn.VirtualNodeName,
			}).
			Return(nil, testErr)
		err := appmeshClient.EnsureVirtualNode(vn)
		Expect(err).To(testutils.HaveInErrorChain(testErr))
	})

	It("EnsureVirtualNode should update if exists but not equal", func() {
		vn := &appmesh.VirtualNodeData{
			MeshName:        meshName,
			VirtualNodeName: aws2.String("vn-name"),
			Spec:            &appmesh.VirtualNodeSpec{},
		}
		existingVs := &appmesh.VirtualNodeData{}
		mockAppmeshRawClient.
			EXPECT().
			DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
				MeshName:        vn.MeshName,
				VirtualNodeName: vn.VirtualNodeName,
			}).
			Return(&appmesh.DescribeVirtualNodeOutput{
				VirtualNode: existingVs,
			}, nil)
		mockAppmeshMatcher.EXPECT().AreVirtualNodesEqual(existingVs, vn).Return(false)
		mockAppmeshRawClient.
			EXPECT().
			UpdateVirtualNode(&appmesh.UpdateVirtualNodeInput{
				MeshName:        vn.MeshName,
				VirtualNodeName: vn.VirtualNodeName,
				Spec:            vn.Spec,
			}).
			Return(nil, nil)
		err := appmeshClient.EnsureVirtualNode(vn)
		Expect(err).ToNot(HaveOccurred())
	})

	It("ReconcileVirtualNodes", func() {
		declaredVs := []*appmesh.VirtualNodeData{
			{
				MeshName:        meshName,
				VirtualNodeName: aws2.String("vn-name-1"),
			}, {
				MeshName:        meshName,
				VirtualNodeName: aws2.String("vn-name-2"),
			}, {
				MeshName:        meshName,
				VirtualNodeName: aws2.String("vn-name-3"),
			},
			{
				MeshName:        aws2.String("some-other-mesh"),
				VirtualNodeName: aws2.String("vn-name-3"),
			},
		}
		for _, vn := range declaredVs[:3] {
			expectVirtualNodeCreate(vn)
		}
		existingVn := []*appmesh.VirtualNodeRef{
			{
				MeshName:        meshName,
				VirtualNodeName: aws2.String("vn-name-3"),
			},
			{
				MeshName:        meshName,
				VirtualNodeName: aws2.String("vn-name-4"),
			},
		}
		mockAppmeshRawClient.
			EXPECT().
			ListVirtualNodesPagesWithContext(ctx, &appmesh.ListVirtualNodesInput{
				MeshName: meshName,
			}, gomock.Any()).
			DoAndReturn(func(_ aws2.Context, vsReq *appmesh.ListVirtualNodesInput, callback func(*appmesh.ListVirtualNodesOutput, bool) bool) error {
				callback(&appmesh.ListVirtualNodesOutput{
					VirtualNodes: existingVn,
				}, true)
				return nil
			})
		mockAppmeshRawClient.
			EXPECT().
			DeleteVirtualNode(&appmesh.DeleteVirtualNodeInput{
				MeshName:        meshName,
				VirtualNodeName: existingVn[1].VirtualNodeName,
			}).
			Return(nil, nil)
		err := appmeshClient.ReconcileVirtualNodes(ctx, meshName, declaredVs)
		Expect(err).ToNot(HaveOccurred())
	})
})
