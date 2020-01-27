package version_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	mock_server "github.com/solo-io/mesh-projects/cli/pkg/tree/version/server/mocks"
	"github.com/solo-io/mesh-projects/pkg/version"
	"github.com/spf13/cobra"
)

var _ = Describe("Version", func() {
	var (
		ctrl                    *gomock.Controller
		meshctl                 *cli_mocks.MockMeshctl
		mockServerVersionClient *mock_server.MockServerVersionClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockServerVersionClient = mock_server.NewMockServerVersionClient(ctrl)
		meshctl = &cli_mocks.MockMeshctl{
			MockController:     ctrl,
			MasterVerification: func(cmd *cobra.Command, args []string) (err error) { return },
			Clients: &common.Clients{
				ServerVersionClient: mockServerVersionClient,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("handles the case where master kube config is undefined", func() {
		version.Version = "fake-version"
		mockServerVersionClient.EXPECT().GetServerVersion().Return(nil, nil)
		output, err := meshctl.Invoke("version --master-cluster-config foo --master-write-namespace bar")

		Expect(output).To(Equal("Client: {\"version\":\"fake-version\"}\nServer: version undefined, could not find any version of service mesh hub running\n"))
		Expect(err).NotTo(HaveOccurred())
	})

	It("correctly prints the JSON with a trailing newline", func() {
		version.Version = "fake-version"

		mockServerVersionClient.
			EXPECT().
			GetServerVersion().
			Return(&server.ServerVersion{
				Namespace: "namespace",
				Containers: []*server.ImageMeta{
					{
						Tag:      "latest",
						Name:     "mesh-discovery",
						Registry: "gcr.io/service-mesh-hub/foo/bar",
					},
				},
			}, nil)

		output, err := meshctl.Invoke("version --master-cluster-config foo --master-write-namespace bar")

		Expect(output).To(Equal("Client: {\"version\":\"fake-version\"}\nServer: {\"Namespace\":\"namespace\",\"Containers\":[{\"Tag\":\"latest\",\"Name\":\"mesh-discovery\",\"Registry\":\"gcr.io/service-mesh-hub/foo/bar\"}]}\n"))
		Expect(err).NotTo(HaveOccurred())
	})
})
