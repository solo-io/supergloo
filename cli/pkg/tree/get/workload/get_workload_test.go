package get_workload_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_table_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/mocks"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get Workload Cmd", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		meshctl             *cli_test.MockMeshctl
		mockKubeLoader      *cli_mocks.MockKubeLoader
		mockWorkloadPrinter *mock_table_printing.MockMeshWorkloadPrinter
		mockWorkloadClient  *mock_zephyr_discovery.MockMeshWorkloadClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
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

		meshWorkloads := []*discovery_v1alpha1.MeshWorkload{
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
			List(ctx).
			Return(&discovery_v1alpha1.MeshWorkloadList{
				Items: []discovery_v1alpha1.MeshWorkload{*meshWorkloads[0], *meshWorkloads[1]},
			}, nil)
		mockWorkloadPrinter.EXPECT().
			Print(gomock.Any(), meshWorkloads).
			Return(nil)
		_, err := meshctl.Invoke("get workloads")
		Expect(err).NotTo(HaveOccurred())
	})
})
