package testutils_test

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/test/testutils"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("cluster lock test", func() {

	var clientset kubernetes.Interface

	var _ = BeforeSuite(func() {
		clientset = testutils.MustKubeClient()
	})

	It("can handle a single locking scenario", func() {
		lock, err := testutils.NewTestClusterLocker(clientset, "default", "1")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("can handle synchronous requests", func() {
		for idx := 0; idx < 5; idx++ {
			lock, err := testutils.NewTestClusterLocker(clientset, "default", strconv.Itoa(idx))
			Expect(err).NotTo(HaveOccurred())
			Expect(lock.AcquireLock()).NotTo(HaveOccurred())
			Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
		}
	})

	It("can handle concurrent requests", func() {
		for idx := 0; idx < 5; idx++ {
			idx := idx
			go func() {
				lock, err := testutils.NewTestClusterLocker(clientset, "default", strconv.Itoa(idx))
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.AcquireLock()).NotTo(HaveOccurred())
				Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
			}()
		}
	})

})
