package version_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version/server"
	mock_server "github.com/solo-io/mesh-projects/cli/pkg/tree/version/server/mocks"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/solo-io/mesh-projects/pkg/version"
)

var _ = Describe("Version", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		meshctl                 *cli_mocks.MockMeshctl
		mockServerVersionClient *mock_server.MockServerVersionClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockServerVersionClient = mock_server.NewMockServerVersionClient(ctrl)
		meshctl = &cli_mocks.MockMeshctl{
			MockController: ctrl,
			Clients: common.Clients{
				ServerVersionClient: mockServerVersionClient,
			},
			Ctx: ctx,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("handles the case where master kube config is undefined", func() {
		version.Version = "fake-version"
		mockServerVersionClient.EXPECT().GetServerVersion().Return(nil, nil)
		output, err := meshctl.Invoke("version")

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
				Containers: []*docker.Image{
					{
						Domain: "quay.io",
						Path:   "solo-io/service-mesh-hub/mesh-discovery",
						Tag:    "latest",
					},
				},
			}, nil)

		output, err := meshctl.Invoke("version")

		Expect(output).To(Equal(`Client: {"version":"fake-version"}
Server: {"Namespace":"namespace","Containers":[{"domain":"quay.io","path":"solo-io/service-mesh-hub/mesh-discovery","tag":"latest"}]}
`))
		Expect(err).NotTo(HaveOccurred())
	})
})
