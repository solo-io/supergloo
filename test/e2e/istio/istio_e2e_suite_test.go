package istio_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/solo-io/gloo-mesh/test/e2e/istio/pkg/tests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/test/utils"
	"github.com/solo-io/go-utils/testutils"
)

// to skip testing this package, run `make run-tests SKIP_PACKAGES=test/e2e/istio
// to test only this package, run `make run-tests TEST_PKG=test/e2e/istio
func TestIstio(t *testing.T) {
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
	ensureWorkingDirectory()
	tests.SetupClustersAndFederation(nil)
})

func ensureWorkingDirectory() {
	// ensure we are in proper directory
	currentFile, err := testutils.GetCurrentFile()
	Expect(err).NotTo(HaveOccurred())
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..", "..")
	err = os.Chdir(projectRoot)
	Expect(err).NotTo(HaveOccurred())
}

var _ = AfterSuite(tests.TeardownFederationAndClusters)

var _ = tests.InitializeTests()
