package initialize_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/testutils"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("init", func() {
	It("successfully installs supergloo to the cluster", func() {
		err := utils.Supergloo("init --release 0.0.0")
		Expect(err).NotTo(HaveOccurred())

		kube := testutils.MustKubeClient()
		Eventually(func() error {
			_, err := kube.CoreV1().Namespaces().Get("supergloo-system", v1.GetOptions{})
			return err
		}).ShouldNot(HaveOccurred())
		Eventually(func() (v12.PodPhase, error) {
			p, err := kube.CoreV1().Pods("supergloo-system").Get("supergloo", v1.GetOptions{})
			if err != nil {
				return v12.PodPending, err
			}
			return p.Status.Phase, nil
		}, time.Minute).Should(Equal(v12.PodRunning))
	})
})
