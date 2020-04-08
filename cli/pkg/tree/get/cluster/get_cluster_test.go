package get_cluster_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	mock_table_printing "github.com/solo-io/mesh-projects/cli/pkg/common/table_printing/mocks"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	cli_test "github.com/solo-io/mesh-projects/cli/pkg/test"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get Cluster Cmd", func() {
	var (
		ctrl               *gomock.Controller
		ctx                context.Context
		meshctl            *cli_test.MockMeshctl
		mockKubeLoader     *cli_mocks.MockKubeLoader
		mockClusterPrinter *mock_table_printing.MockKubernetesClusterPrinter
		mockClusterClient  *mock_zephyr_discovery.MockKubernetesClusterClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		mockClusterPrinter = mock_table_printing.NewMockKubernetesClusterPrinter(ctrl)
		mockClusterClient = mock_zephyr_discovery.NewMockKubernetesClusterClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				KubeClusterClient: mockClusterClient,
			},
			KubeLoader: mockKubeLoader,
			Printers: common.Printers{
				KubeClusterPrinter: mockClusterPrinter,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will call the kube cluster printer with the proper data", func() {

		clusters := []*discovery_v1alpha1.KubernetesCluster{
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
		mockClusterClient.EXPECT().
			List(ctx).
			Return(&discovery_v1alpha1.KubernetesClusterList{
				Items: []discovery_v1alpha1.KubernetesCluster{*clusters[0], *clusters[1]},
			}, nil)
		mockClusterPrinter.EXPECT().
			Print(gomock.Any(), clusters).
			Return(nil)
		_, err := meshctl.Invoke("get clusters")
		Expect(err).NotTo(HaveOccurred())
	})
})
