package istio_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
)

var _ = Describe("Istio Root Cmd", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("complains if it is invoked without a subcommand", func() {
		output, err := cli_mocks.MockMeshctl{MockController: ctrl, Ctx: context.TODO()}.Invoke("istio --kubeconfig foo")
		Expect(output).To(BeEmpty())

		nonTerminalCommandErrorBuilder := cli_util.NonTerminalCommand("istio")
		nonTerminalErr := nonTerminalCommandErrorBuilder(nil, nil)
		Expect(err).To(testutils.HaveInErrorChain(nonTerminalErr))
	})
})
