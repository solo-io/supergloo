package version_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	"github.com/solo-io/mesh-projects/pkg/version"
)

var _ = Describe("Version", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("correctly prints the JSON with a trailing newline", func() {
		version.Version = "fake-version"

		output, err := cli_mocks.MockMeshctl{MockController: ctrl}.Invoke("version")

		Expect(output).To(Equal("{\"version\":\"fake-version\"}\n"))
		Expect(err).NotTo(HaveOccurred())
	})
})
