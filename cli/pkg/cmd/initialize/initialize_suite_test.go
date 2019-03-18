package initialize_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/supergloo/test/testutils"
)

func TestInitialize(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Initialize Suite")
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
	testutils.TeardownSuperGloo(testutils.MustKubeClient())
	testutils.WaitForNamespaceTeardown("supergloo-system")
	Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
})
