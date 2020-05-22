package get_virtual_mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_table_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	types3 "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/kubeconfig/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Get VirtualMesh Cmd", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		meshctl                *cli_test.MockMeshctl
		mockKubeLoader         *mock_kubeconfig.MockKubeLoader
		mockVirtualMeshPrinter *mock_table_printing.MockVirtualMeshPrinter
		mockVirtualMeshClient  *mock_zephyr_networking.MockVirtualMeshClient
		mockMeshClient         *mock_core.MockMeshClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = mock_kubeconfig.NewMockKubeLoader(ctrl)
		mockVirtualMeshPrinter = mock_table_printing.NewMockVirtualMeshPrinter(ctrl)
		mockVirtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				VirtualMeshClient: mockVirtualMeshClient,
				MeshClient:        mockMeshClient,
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

		virtualMeshes := []*zephyr_networking.VirtualMesh{
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
		meshes := []*zephyr_discovery.Mesh{
			{
				Spec: types.MeshSpec{
					MeshType: &types.MeshSpec_AwsAppMesh_{},
				},
			},
			{
				Spec: types.MeshSpec{
					MeshType: &types.MeshSpec_Istio{},
				},
			},
		}
		mockKubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(nil, nil)
		mockVirtualMeshClient.EXPECT().
			ListVirtualMesh(ctx).
			Return(&zephyr_networking.VirtualMeshList{
				Items: []zephyr_networking.VirtualMesh{*virtualMeshes[0], *virtualMeshes[1]},
			}, nil)

		for i, vm := range virtualMeshes {
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, client.ObjectKey{Name: vm.Spec.GetMeshes()[0].GetName(), Namespace: vm.Spec.GetMeshes()[0].GetNamespace()}).
				Return(meshes[i], nil)
		}

		mockVirtualMeshPrinter.EXPECT().
			Print(gomock.Any(), virtualMeshes, meshes).
			Return(nil)
		_, err := meshctl.Invoke("get virtualmeshes")
		Expect(err).NotTo(HaveOccurred())
	})
})
