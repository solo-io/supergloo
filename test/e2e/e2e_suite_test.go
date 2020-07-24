package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/service-mesh-hub/test/e2e"
)

// to skip testing this package, run `make run-tests SKIP_PACKAGES=test/e2e
// to test only this package, run `make run-tests TEST_PKG=test/e2e
func TestE2e(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		runShell("./ci/print-kind-info.sh")
		Fail(message, callerSkip...)
	})
	RunSpecs(t, "E2e Suite")
}

func runShell(c string) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "[%v] command FAILED: %v", c, err)
		return
	}
	fmt.Fprintf(GinkgoWriter, "[%v] command result: \n%v", c, buf.String())
}

var _ = BeforeSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	ensureWorkingDirectory()
	/* env := */ StartEnvOnce(ctx)
})

func ensureWorkingDirectory() {
	// ensure we are in proper directory
	currentFile, err := testutils.GetCurrentFile()
	Expect(err).NotTo(HaveOccurred())
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	err = os.Chdir(projectRoot)
	Expect(err).NotTo(HaveOccurred())
}

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if os.Getenv("NO_CLEANUP") != "" {
		return
	}
	_ = ClearEnv(ctx)
})
