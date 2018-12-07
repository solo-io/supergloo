package cli_test

import (
	"testing"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	supergloo_dir = "github.com/solo-io/supergloo/cli/cmd"
)

var SuperglooExec string

var _ = BeforeSuite(func() {
	var err error
	SuperglooExec, err = gexec.Build(supergloo_dir)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {

	gexec.CleanupBuildArtifacts()
})

func TestCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cli Suite")
}
