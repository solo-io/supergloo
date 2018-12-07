package cli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cli Suite")
}

const (
	supergloo_dir  = "$GOPATH/src/github.com/solo-io/supergloo"
	cli_builder    = "make-cli.sh"
	test_suite_dir = "test/cli"
)

var _ = BeforeSuite(func() {
	err := os.Chdir(os.ExpandEnv(supergloo_dir))
	Expect(err).NotTo(HaveOccurred())
	make := exec.Command("/bin/sh", path.Join(test_suite_dir, cli_builder))
	err = make.Run()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(fmt.Sprintf("%s/_output", test_suite_dir))
	Expect(err).NotTo(HaveOccurred())
})
