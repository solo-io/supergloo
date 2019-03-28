package helm_test

import (
	"os"
	"testing"

	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/supergloo/test/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}

var (
	lock *clusterlock.TestClusterLocker
	err  error
)

var _ = BeforeSuite(func() {
	kubeClient = testutils.MustKubeClient()
	lock, err = clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
		IdPrefix: os.ExpandEnv("supergloo-helm-{$BUILD_ID}-"),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
})
