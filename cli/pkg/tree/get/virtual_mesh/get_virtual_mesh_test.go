package get_virtual_mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_table_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	types3 "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig/mocks"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get VirtualMesh Cmd", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		meshctl                *cli_test.MockMeshctl
		mockKubeLoader         *mock_kubeconfig.MockKubeLoader
		mockVirtualMeshPrinter *mock_table_printing.MockVirtualMeshPrinter
		mockVirtualMeshClient  *mock_smh_networking.MockVirtualMeshClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = mock_kubeconfig.NewMockKubeLoader(ctrl)
		mockVirtualMeshPrinter = mock_table_printing.NewMockVirtualMeshPrinter(ctrl)
		mockVirtualMeshClient = mock_smh_networking.NewMockVirtualMeshClient(ctrl)
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

		virtualMeshes := []*smh_networking.VirtualMesh{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "virtualmesh-1",
				},
				Spec: types2.VirtualMeshSpec{
					Meshes: []*types3.ResourceRef{
						{Name: "mesh-1", Namespace: "mesh-namespace-1"},
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "virtualmesh-2",
				},
				Spec: types2.VirtualMeshSpec{
					Meshes: []*types3.ResourceRef{
						{Name: "mesh-2", Namespace: "mesh-namespace-2"},
					},
				},
			},
		}
		mockKubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(nil, nil)
		mockVirtualMeshClient.EXPECT().
			ListVirtualMesh(ctx).
			Return(&smh_networking.VirtualMeshList{
				Items: []smh_networking.VirtualMesh{*virtualMeshes[0], *virtualMeshes[1]},
			}, nil)

		mockVirtualMeshPrinter.EXPECT().
			Print(gomock.Any(), virtualMeshes).
			Return(nil)
		_, err := meshctl.Invoke("get virtualmeshes")
		Expect(err).NotTo(HaveOccurred())
	})
})
