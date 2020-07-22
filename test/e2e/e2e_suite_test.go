package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/service-mesh-hub/test/e2e"
)

// to skip testing this package, run `make run-tests SKIP_PACKAGES=test/e2e
// to test only this package, run `make run-tests TEST_PKG=test/e2e
func TestE2e(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		runShell("kubectl logs -n=service-mesh-hub -l app=discovery")
		runShell("kubectl logs -n=service-mesh-hub -l app=networking")
		runShell("kubectl get pod -A")
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
	/* env := */ StartEnvOnce(ctx)
})

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if os.Getenv("NO_CLEANUP") != "" {
		return
	}
	_ = ClearEnv(ctx)
})
