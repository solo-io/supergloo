package demo_init_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_exec "github.com/solo-io/mesh-projects/cli/pkg/common/exec/mocks"
	cli_test "github.com/solo-io/mesh-projects/cli/pkg/test"
)

var _ = Describe("Demo init", func() {
	var (
		ctrl    *gomock.Controller
		ctx     context.Context
		runner  *mock_exec.MockRunner
		meshctl *cli_test.MockMeshctl
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		runner = mock_exec.NewMockRunner(ctrl)

		meshctl = &cli_test.MockMeshctl{
			Runner:         runner,
			MockController: ctrl,
			Ctx:            ctx,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will call the runner", func() {
		runner.EXPECT().Run("bash", "-c", gomock.Any()).Return(nil)
		_, err := meshctl.Invoke("demo init")
		Expect(err).NotTo(HaveOccurred())
	})
})
