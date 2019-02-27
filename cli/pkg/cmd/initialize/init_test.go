package initialize_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/test/utils"
	"github.com/solo-io/supergloo/test/testutils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("init", func() {
	It("successfully installs supergloo to the cluster", func() {
		err := utils.Supergloo("init")
		Expect(err).NotTo(HaveOccurred())

		kube := testutils.MustKubeClient()
		Eventually(func() error {
			_, err := kube.CoreV1().Namespaces().Get("supergloo-system", v1.GetOptions{})
			return err
		}).ShouldNot(HaveOccurred())
	})
})
