package istio_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/test/data"
	. "github.com/solo-io/service-mesh-hub/test/e2e"
	"github.com/solo-io/service-mesh-hub/test/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	err                 error
	VirtualMesh         *networkingv1alpha2.VirtualMesh
	VirtualMeshManifest utils.Manifest
)

// to skip testing this package, run `make run-tests SKIP_PACKAGES=test/e2e
// to test only this package, run `make run-tests TEST_PKG=test/e2e
func TestE2e(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		fmt.Println("skipping E2E tests")
		return
	}
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

// Before running tests, federate the two clusters by creating a VirtualMesh with mTLS enabled.
var _ = BeforeSuite(func() {
	VirtualMeshManifest, err = utils.NewManifest("virtualmesh.yaml")
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	ensureWorkingDirectory()
	/* env := */ StartEnvOnce(ctx)

	var err error
	dynamicClient, err = client.New(GetEnv().Management.Config, client.Options{})
	Expect(err).NotTo(HaveOccurred())

	federateClusters(dynamicClient)
})

func ensureWorkingDirectory() {
	// ensure we are in proper directory
	currentFile, err := testutils.GetCurrentFile()
	Expect(err).NotTo(HaveOccurred())
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	err = os.Chdir(projectRoot)
	Expect(err).NotTo(HaveOccurred())
}

func federateClusters(dynamicClient client.Client) {
	VirtualMesh = data.SelfSignedVirtualMesh(
		"bookinfo-federation",
		BookinfoNamespace,
		[]*v1.ObjectRef{
			masterMesh,
			remoteMesh,
		})

	err = VirtualMeshManifest.AppendResources(VirtualMesh)
	Expect(err).NotTo(HaveOccurred())
	err = VirtualMeshManifest.KubeApply(BookinfoNamespace)
	Expect(err).NotTo(HaveOccurred())

	// ensure status is updated
	utils.AssertVirtualMeshStatuses(dynamicClient, BookinfoNamespace)

	// check we can hit the remote service
	// give 5 minutes because the workflow depends on restarting pods
	// which can take several minutes
	Eventually(curlRemoteReviews, "5m", "2s").Should(ContainSubstring("200 OK"))
}

var _ = AfterSuite(func() {
	err = VirtualMeshManifest.KubeDelete(BookinfoNamespace)
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if os.Getenv("NO_CLEANUP") != "" {
		return
	}
	_ = ClearEnv(ctx)
})
