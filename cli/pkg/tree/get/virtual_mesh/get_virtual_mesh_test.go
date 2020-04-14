package get_virtual_mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_table_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/mocks"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get VirtualMesh Cmd", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		meshctl                *cli_test.MockMeshctl
		mockKubeLoader         *cli_mocks.MockKubeLoader
		mockVirtualMeshPrinter *mock_table_printing.MockVirtualMeshPrinter
		mockVirtualMeshClient  *mock_zephyr_networking.MockVirtualMeshClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		mockVirtualMeshPrinter = mock_table_printing.NewMockVirtualMeshPrinter(ctrl)
		mockVirtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				VirtualMeshClient: mockVirtualMeshClient,
			},
			KubeLoader: mockKubeLoader,
			Printers: common.Printers{
				VirtualMeshPrinter: mockVirtualMeshPrinter,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will call the VirtualMesh Printer with the proper data", func() {

		virtualMeshes := []*v1alpha1.VirtualMesh{
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
		mockVirtualMeshClient.EXPECT().
			ListVirtualMesh(ctx).
			Return(&v1alpha1.VirtualMeshList{
				Items: []v1alpha1.VirtualMesh{*virtualMeshes[0], *virtualMeshes[1]},
			}, nil)
		mockVirtualMeshPrinter.EXPECT().
			Print(gomock.Any(), virtualMeshes).
			Return(nil)
		_, err := meshctl.Invoke("get virtualmeshes")
		Expect(err).NotTo(HaveOccurred())
	})
})
