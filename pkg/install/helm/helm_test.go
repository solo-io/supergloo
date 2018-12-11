package helm_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/install/helm"
)

var _ = Describe("HelmTest", func() {
	It("Can get helm client", func() {
		helmClient := helm.KubeHelmClient{}
		_, err := helmClient.GetHelmClient(context.TODO())
		helmClient.Teardown()
		Expect(err).NotTo(HaveOccurred())
	})
})
