package cli_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	supergloo_dir        = "github.com/solo-io/supergloo/cli/cmd"
	delete_ruoting_rules = "delete routingrules.supergloo.solo.io --all -n supergloo-system"
)

var SuperglooExec string

var _ = BeforeSuite(func() {
	var err error
	SuperglooExec, err = gexec.Build(supergloo_dir)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
	cmd := exec.Command("kubectl", strings.Split(delete_ruoting_rules, " ")...)
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred())
})

func TestCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cli Suite")
}
