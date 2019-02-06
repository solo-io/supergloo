package helm_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/supergloo/pkg2/install/helm"
)

var _ = Describe("HelmChartInstaller", func() {
	It("works without tiller! can you belize it?!", func() {
		manifests, err := RenderManifests(
			context.TODO(),
			"/home/ilackarms/go/src/github.com/solo-io/supergloo/istio-1.0.3.tgz",
			"",
			"yella",
			"istio-system",
			"",
			true,
		)
		Expect(err).NotTo(HaveOccurred())
		err = ApplyManifests(context.TODO(), "istio-system", manifests)
		Expect(err).NotTo(HaveOccurred())
	})
})
