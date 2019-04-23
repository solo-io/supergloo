package kubernetes_test

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	. "github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	"github.com/solo-io/supergloo/test/testutils"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodBaseClient", func() {
	var namespace string
	BeforeEach(func() {
		namespace = "podclient-" + helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})
	It("converts a kubernetes pod to solo-kit resource", func() {
		kube := testutils.MustKubeClient()
		kcache, err := cache.NewKubeCoreCache(context.TODO(), kube)
		Expect(err).NotTo(HaveOccurred())
		rc := NewPodResourceClient(kube, kcache)

		pod, err := kube.CoreV1().Pods(namespace).Create(&kubev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "happy",
				Namespace: namespace,
			},
			Spec: kubev1.PodSpec{
				Containers: []kubev1.Container{
					{
						Name:  "nginx",
						Image: "nginx:latest",
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		var pods resources.ResourceList
		Eventually(func() (resources.ResourceList, error) {
			pods, err = rc.List(namespace, clients.ListOpts{})
			return pods, err
		}).Should(HaveLen(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(pods).To(HaveLen(1))
		Expect(pods[0].GetMetadata().Name).To(Equal(pod.Name))
		Expect(pods[0].GetMetadata().Namespace).To(Equal(pod.Namespace))
		kubePod, err := ToKubePod(pods[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(kubePod.Spec.Containers).To(Equal(pod.Spec.Containers))
	})
})
