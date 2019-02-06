package helm_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/supergloo/pkg2/install/helm"
)

var _ = Describe("HelmChartInstaller", func() {
	It("works without tiller! can you belize it?!", func() {
		ns := "test"
		manifests, err := RenderManifests(
			context.TODO(),
			"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
			"",
			"yella",
			ns,
			"",
			true,
		)
		Expect(err).NotTo(HaveOccurred())
		err = ApplyManifests(context.TODO(), ns, manifests)
		Expect(err).NotTo(HaveOccurred())
		err = DeleteManifests(context.TODO(), ns, manifests)
		Expect(err).NotTo(HaveOccurred())
	})
})
