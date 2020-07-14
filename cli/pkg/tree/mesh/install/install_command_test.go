package mesh_install_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
)

var _ = Describe("Install Root Cmd", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will fail if an invalid mesh install is requested", func() {
		output, err := cli_test.MockMeshctl{MockController: ctrl, Ctx: context.TODO()}.Invoke("mesh install fake")
		Expect(output).To(BeEmpty())

		Expect(err).To(HaveOccurred())
	})
})
