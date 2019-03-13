package testutils_test

import (
	"sync"
	"time"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/test/testutils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("cluster lock test", func() {

	var kubeClient kubernetes.Interface

	var _ = BeforeSuite(func() {
		kubeClient = testutils.MustKubeClient()
	})

	var _ = AfterSuite(func() {
		kubeClient.CoreV1().ConfigMaps("default").Delete(testutils.LockResourceName, &v1.DeleteOptions{})
	})

	It("can handle a single locking scenario", func() {
		lock, err := testutils.NewTestClusterLocker(kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("can handle synchronous requests", func() {
		for idx := 0; idx < 5; idx++ {
			lock, err := testutils.NewTestClusterLocker(kubeClient, "default")
			Expect(err).NotTo(HaveOccurred())
			Expect(lock.AcquireLock()).NotTo(HaveOccurred())
			Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
		}
	})

	It("can handle concurrent requests", func() {
		x := ""
		sharedString := &x
		wg := sync.WaitGroup{}
		for idx := 0; idx < 5; idx++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()
				lock, err := testutils.NewTestClusterLocker(kubeClient, "default")
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.AcquireLock(retry.Delay(2 * time.Second))).NotTo(HaveOccurred())
				Expect(*sharedString).To(Equal(""))
				*sharedString = "hello"
				time.Sleep(2 * time.Second)
				*sharedString = ""
				Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
			}()
		}
		wg.Wait()
	})
})
