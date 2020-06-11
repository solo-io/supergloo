package get_vmcsr_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_table_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	mock_kubeconfig "github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig/mocks"
	mock_security_config "github.com/solo-io/service-mesh-hub/test/mocks/clients/security.smh.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get VirtualMesh Cmd", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		meshctl                   *cli_test.MockMeshctl
		mockKubeLoader            *mock_kubeconfig.MockKubeLoader
		mockVirtualMeshCSRPrinter *mock_table_printing.MockVirtualMeshCSRPrinter
		mockVirtualMeshCSRClient  *mock_security_config.MockVirtualMeshCertificateSigningRequestClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = mock_kubeconfig.NewMockKubeLoader(ctrl)
		mockVirtualMeshCSRPrinter = mock_table_printing.NewMockVirtualMeshCSRPrinter(ctrl)
		mockVirtualMeshCSRClient = mock_security_config.NewMockVirtualMeshCertificateSigningRequestClient(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Ctx:            ctx,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				VirtualMeshCSRClient: mockVirtualMeshCSRClient,
			},
			KubeLoader: mockKubeLoader,
			Printers: common.Printers{
				VirtualMeshCSRPrinter: mockVirtualMeshCSRPrinter,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will call the VirtualMeshCSR Printer with the proper data", func() {

		virtualMeshes := []*smh_security.VirtualMeshCertificateSigningRequest{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-csr-1",
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "mesh-csr-2",
				},
			},
		}
		mockKubeLoader.EXPECT().
			GetRestConfigForContext("", "").
			Return(nil, nil)
		mockVirtualMeshCSRClient.EXPECT().
			ListVirtualMeshCertificateSigningRequest(ctx).
			Return(&smh_security.VirtualMeshCertificateSigningRequestList{
				Items: []smh_security.VirtualMeshCertificateSigningRequest{*virtualMeshes[0], *virtualMeshes[1]},
			}, nil)
		mockVirtualMeshCSRPrinter.EXPECT().
			Print(gomock.Any(), virtualMeshes).
			Return(nil)
		_, err := meshctl.Invoke("get vmcsr")
		Expect(err).NotTo(HaveOccurred())
	})
})
