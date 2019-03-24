package gloo_test

import (
	"testing"

	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/supergloo/test/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGloo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Meshingress Suite")
}

var (
	lock *clusterlock.TestClusterLocker
	err  error
)

var _ = BeforeSuite(func() {
	kubeClient := testutils.MustKubeClient()
	lock, err = clusterlock.NewTestClusterLocker(kubeClient, "default")
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	defer lock.ReleaseLock()
	kubeClient := testutils.MustKubeClient()
	kubeClient.CoreV1().Namespaces().Delete("gloo-system", nil)
	testutils.WaitForNamespaceTeardown("gloo-system")
})
