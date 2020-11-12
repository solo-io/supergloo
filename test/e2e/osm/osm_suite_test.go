package osm_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/gloo-mesh/test/e2e"
	"github.com/solo-io/gloo-mesh/test/utils"
	"github.com/solo-io/go-utils/testutils"
)

// to skip testing this package, run `make run-tests SKIP_PACKAGES=test/e2e
// to test only this package, run `make run-tests TEST_PKG=test/e2e
func TestOsm(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		fmt.Println("skipping E2E tests")
		return
	}
	RegisterFailHandler(func(message string, callerSkip ...int) {
		utils.RunShell("./ci/print-kind-info.sh")
		Fail(message, callerSkip...)
	})
	RunSpecs(t, "E2e Suite")
}

// Before running tests, federate the two clusters by creating a VirtualMesh with mTLS enabled.
var _ = BeforeSuite(func() {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	ensureWorkingDirectory()
	/* env := */ StartSingleClusterEnvOnce(ctx)
})

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if os.Getenv("NO_CLEANUP") != "" {
		return
	}
	_ = ClearSingleClusterEnv(ctx)
})

func ensureWorkingDirectory() {
	// ensure we are in proper directory
	currentFile, err := testutils.GetCurrentFile()
	Expect(err).NotTo(HaveOccurred())
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..", "..")
	err = os.Chdir(projectRoot)
	Expect(err).NotTo(HaveOccurred())
}
