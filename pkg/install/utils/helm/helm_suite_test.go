package helm_test

import (
	"testing"

	"github.com/solo-io/supergloo/test/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}

var (
	lock *testutils.TestClusterLocker
	err  error
)

var _ = BeforeSuite(func() {
	kubeClient = testutils.MustKubeClient()
	lock, err = testutils.NewTestClusterLocker(kubeClient, "default", "1")
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
})
