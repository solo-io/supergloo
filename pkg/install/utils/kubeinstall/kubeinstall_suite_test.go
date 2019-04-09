package kubeinstall_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/supergloo/test/testutils"
)

func TestKubeinstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubeinstall Suite")
}

var (
	lock *clusterlock.TestClusterLocker
	err  error
)

var _ = BeforeSuite(func() {
	lock, err = clusterlock.NewTestClusterLocker(testutils.MustKubeClient(), clusterlock.Options{
		IdPrefix: os.ExpandEnv("supergloo-helm-{$BUILD_ID}-"),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
})
