package initialize_test

import (
	"testing"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

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
	kubeClient := clients.MustKubeClient()
	lock, err = clusterlock.NewTestClusterLocker(kubeClient, "default")
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	defer lock.ReleaseLock()
	testutils.TeardownSuperGloo(clients.MustKubeClient())
	testutils.WaitForNamespaceTeardown("supergloo-system")
})
