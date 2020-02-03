package cluster_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	cli_mocks "github.com/solo-io/mesh-projects/cli/pkg/mocks"
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
)

var _ = Describe("Cluster Root Cmd", func() {
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("complains if it is invoked without a subcommand", func() {
		output, err := cli_mocks.MockMeshctl{MockController: ctrl}.Invoke("cluster --kubeconfig foo")
		Expect(output).To(BeEmpty())

		nonTerminalCommandErrorBuilder := cli_util.NonTerminalCommand("cluster")
		nonTerminalErr := nonTerminalCommandErrorBuilder(nil, nil)
		Expect(err).To(testutils.HaveInErrorChain(nonTerminalErr))
	})
})
