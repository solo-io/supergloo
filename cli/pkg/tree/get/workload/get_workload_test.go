package get_workload_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_table_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig/mocks"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get Workload Cmd", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		meshctl             *cli_test.MockMeshctl
		mockKubeLoader      *mock_kubeconfig.MockKubeLoader
		mockWorkloadPrinter *mock_table_printing.MockMeshWorkloadPrinter
		mockWorkloadClient  *mock_zephyr_discovery.MockMeshWorkloadClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = mock_kubeconfig.NewMockKubeLoader(ctrl)
		mockWorkloadPrinter = mock_table_printing.NewMockMeshWorkloadPrinter(ctrl)
		mockWorkloadClient = mock_zephyr_discovery.NewMockMeshWorkloadClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				MeshWorkloadClient: mockWorkloadClient,
			},
			KubeLoader: mockKubeLoader,
			Printers: common.Printers{
				MeshWorkloadPrinter: mockWorkloadPrinter,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will call the mesh workload printer with the proper data", func() {

		meshWorkloads := []*zephyr_discovery.MeshWorkload{
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
		mockWorkloadClient.EXPECT().
			ListMeshWorkload(ctx).
			Return(&zephyr_discovery.MeshWorkloadList{
				Items: []zephyr_discovery.MeshWorkload{*meshWorkloads[0], *meshWorkloads[1]},
			}, nil)
		mockWorkloadPrinter.EXPECT().
			Print(gomock.Any(), meshWorkloads).
			Return(nil)
		_, err := meshctl.Invoke("get workloads")
		Expect(err).NotTo(HaveOccurred())
	})
})
