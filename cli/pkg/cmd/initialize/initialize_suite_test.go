package initialize_test

import (
	"testing"

	"github.com/solo-io/supergloo/test/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInitialize(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Initialize Suite")
}

var (
	lock *testutils.TestClusterLocker
	err  error
)

var _ = BeforeSuite(func() {
	kubeClient := testutils.MustKubeClient()
	lock, err = testutils.NewTestClusterLocker(kubeClient, "default")
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
})
