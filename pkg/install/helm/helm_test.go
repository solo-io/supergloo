package helm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/install/helm"
)

var _ = Describe("HelmTest", func() {
	It("Can get helm client", func() {
		_, err := helm.GetHelmClient()
		helm.Teardown()
		Expect(err).NotTo(HaveOccurred())
	})
})
