package mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
)

var _ = Describe("Mesh Root Cmd", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("complains if it is invoked without a subcommand", func() {
		output, err := cli_test.MockMeshctl{MockController: ctrl, Ctx: context.TODO()}.Invoke("mesh")
		Expect(output).To(BeEmpty())

		nonTerminalCommandErrorBuilder := common.NonTerminalCommand("mesh")
		nonTerminalErr := nonTerminalCommandErrorBuilder(nil, nil)
		Expect(err).To(testutils.HaveInErrorChain(nonTerminalErr))
	})
})
