package kubernetes

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

var _ = Describe("Namespace base client", func() {

	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace string
		cfg       *rest.Config
		client    *configMapResourceClient
		kube      kubernetes.Interface
		kubeCache cache.KubeCoreCache
		cmObj     *kubev1.ConfigMap
	)

	BeforeEach(func() {
		namespace = helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace)
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kube, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		kubeCache, err = cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		client = newResourceClient(kube, kubeCache)
		Expect(err).NotTo(HaveOccurred())
		cmObj, err = kube.CoreV1().ConfigMaps(namespace).Create(&kubev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      namespace,
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})
	It("converts a kubernetes pod to solo-kit resource", func() {

		Eventually(func() bool {
			configmaps, err := client.List("", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			foundCm := false
			for _, v := range configmaps {
				if v.GetMetadata().Name == cmObj.Name {
					foundCm = true
				}
			}
			return foundCm
		}, time.Minute, time.Second*15).Should(BeTrue())

	})
})
