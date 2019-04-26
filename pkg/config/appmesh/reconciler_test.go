package appmesh_test

import (
	"context"

	api "github.com/aws/aws-sdk-go/service/appmesh"
	gm "github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	mocks "github.com/solo-io/supergloo/pkg/config/appmesh/mocks"
	"github.com/solo-io/supergloo/pkg/translator/appmesh"

	. "github.com/solo-io/supergloo/pkg/config/appmesh"
)

var _ = Describe("Reconciler", func() {

	var (
		ctrl       *gm.Controller
		client     *mocks.MockClient
		reconciler Reconciler
		mesh       = &v1.Mesh{
			Metadata: core.Metadata{
				Name:      "test-mesh",
				Namespace: "test-ns",
			},
			MeshType: &v1.Mesh_AwsAppMesh{
				AwsAppMesh: &v1.AwsAppMesh{},
			},
		}
		meshName = mesh.Metadata.Name
		md       = &api.MeshData{}
		vnd      = &api.VirtualNodeData{}
		vsd      = &api.VirtualServiceData{}
		vrd      = &api.VirtualRouterData{}
		rd       = &api.RouteData{}
	)

	BeforeEach(func() {
		ctrl = gm.NewController(T)
	})

	AfterEach(func() {
		defer ctrl.Finish()
	})

	JustBeforeEach(func() {
		builder := mocks.NewMockClientBuilder(ctrl)
		builder.EXPECT().GetClientInstance(gm.Any(), gm.Any()).Return(client, nil)
		reconciler = NewReconciler(builder)
	})

	Context("mesh does not exist", func() {

		BeforeEach(func() {
			client = mocks.NewMockClient(ctrl)
			client.EXPECT().GetMesh(gm.Any(), gm.Any()).Return(nil, nil).Times(1)

			client.EXPECT().CreateMesh(gm.Any(), meshName).Return(md, nil).Times(1)
			client.EXPECT().CreateVirtualNode(gm.Any(), gm.Any()).Return(vnd, nil).Times(3)
			client.EXPECT().CreateVirtualRouter(gm.Any(), gm.Any()).Return(vrd, nil).Times(2)
			client.EXPECT().CreateRoute(gm.Any(), gm.Any()).Return(rd, nil).Times(2)
			client.EXPECT().CreateVirtualService(gm.Any(), gm.Any()).Return(vsd, nil).Times(2)
		})

		It("creates all the resources", func() {
			snapshot := appmesh.ResourceSnapshot{
				MeshName:        meshName,
				VirtualNodes:    map[string]*api.VirtualNodeData{"vn-1": {}, "vn-2": {}, "vn-3": {}},
				VirtualServices: map[string]*api.VirtualServiceData{"vs-1": {}, "vs-2": {}},
				VirtualRouters:  map[string]*api.VirtualRouterData{"vr-1": {}, "vr-2": {}},
				Routes:          map[string]*api.RouteData{"route-1": {}, "route-2": {}},
			}

			err := reconciler.Reconcile(context.TODO(), mesh, &snapshot)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("mesh exists", func() {

		BeforeEach(func() {
			client = mocks.NewMockClient(ctrl)
			client.EXPECT().GetMesh(gm.Any(), gm.Any()).Return(md, nil).Times(1)

			client.EXPECT().ListVirtualNodes(gm.Any(), meshName).Return([]string{"vn-1", "vn-2", "vn-3"}, nil).Times(1)
			client.EXPECT().ListVirtualServices(gm.Any(), meshName).Return([]string{"vs-1", "vs-2", "vs-3"}, nil).Times(1)
			client.EXPECT().ListVirtualRouters(gm.Any(), meshName).Return([]string{"vr-1", "vr-2"}, nil).Times(1)
			client.EXPECT().ListRoutes(gm.Any(), meshName, "vr-1").Return([]string{"vr1-route-1"}, nil).Times(1)
			client.EXPECT().ListRoutes(gm.Any(), meshName, "vr-2").Return([]string{"vr2-route-1", "vr2-route-2"}, nil).Times(1)

			// VN: 3 updated
			client.EXPECT().UpdateVirtualNode(gm.Any(), gm.Any()).Return(vnd, nil).Times(3)

			// VS: 1 created, 2 updated, 1 deleted
			client.EXPECT().CreateVirtualService(gm.Any(), gm.Any()).Return(vsd, nil).Times(1)
			client.EXPECT().UpdateVirtualService(gm.Any(), gm.Any()).Return(vsd, nil).Times(2)
			client.EXPECT().DeleteVirtualService(gm.Any(), gm.Any(), "vs-3").Return(nil).Times(1)

			// VR: 1 created, 1 updated, 1 deleted
			client.EXPECT().CreateVirtualRouter(gm.Any(), gm.Any()).Return(vrd, nil).Times(1)
			client.EXPECT().UpdateVirtualRouter(gm.Any(), gm.Any()).Return(vrd, nil).Times(1)
			client.EXPECT().DeleteVirtualRouter(gm.Any(), gm.Any(), "vr-1").Return(nil).Times(1)

			// Routes: 2 updated, 1 deleted
			client.EXPECT().UpdateRoute(gm.Any(), gm.Any()).Return(rd, nil).Times(2)
			client.EXPECT().DeleteRoute(gm.Any(), gm.Any(), "vr-2", "vr2-route-2").Return(nil).Times(1)
		})

		It("correctly reconciles the resources", func() {
			snapshot := appmesh.ResourceSnapshot{
				MeshName:        meshName,
				VirtualNodes:    map[string]*api.VirtualNodeData{"vn-1": {}, "vn-2": {}, "vn-3": {}},
				VirtualServices: map[string]*api.VirtualServiceData{"vs-1": {}, "vs-2": {}, "vs-4": {}},
				VirtualRouters:  map[string]*api.VirtualRouterData{"vr-2": {}, "vr-3": {}},
				Routes:          map[string]*api.RouteData{"vr1-route-1": {}, "vr2-route-1": {}},
			}

			err := reconciler.Reconcile(context.TODO(), mesh, &snapshot)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
