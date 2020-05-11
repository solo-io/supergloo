package get_mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_table_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/kubeconfig/mocks"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get Mesh Cmd", func() {
	var (
		ctrl            *gomock.Controller
		ctx             context.Context
		meshctl         *cli_test.MockMeshctl
		mockKubeLoader  *mock_kubeconfig.MockKubeLoader
		mockMeshPrinter *mock_table_printing.MockMeshPrinter
		mockMeshClient  *mock_zephyr_discovery.MockMeshClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = mock_kubeconfig.NewMockKubeLoader(ctrl)
		mockMeshPrinter = mock_table_printing.NewMockMeshPrinter(ctrl)
		mockMeshClient = mock_zephyr_discovery.NewMockMeshClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				MeshClient: mockMeshClient,
			},
			KubeLoader: mockKubeLoader,
			Printers: common.Printers{
				MeshPrinter: mockMeshPrinter,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will call the Mesh Printer with the proper data", func() {

		meshes := []*zephyr_discovery.Mesh{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-1",
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-2",
				},
			},
		}
		mockKubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(nil, nil)
		mockMeshClient.EXPECT().
			ListMesh(ctx).
			Return(&zephyr_discovery.MeshList{
				Items: []zephyr_discovery.Mesh{*meshes[0], *meshes[1]},
			}, nil)
		mockMeshPrinter.EXPECT().
			Print(gomock.Any(), meshes).
			Return(nil)
		_, err := meshctl.Invoke("get meshes")
		Expect(err).NotTo(HaveOccurred())
	})
})
