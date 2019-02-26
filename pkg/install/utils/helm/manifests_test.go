package helm_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/supergloo/pkg/install/utils/helm"
)

var _ = Describe("Manifests", func() {
	Context("gzip", func() {
		It("gzips the entire set of manifests down to a gzipped base-64 encoded string", func() {
			manifests, err := RenderManifests(
				context.TODO(),
				"https://s3.amazonaws.com/supergloo.solo.io/istio-1.0.3.tgz",
				"",
				"yella",
				"anything",
				"",
				true,
			)
			Expect(err).NotTo(HaveOccurred())
			compressed, err := manifests.Gzipped()
			Expect(err).NotTo(HaveOccurred())
			Expect(compressed).To(HaveLen(23768))

			newManifests, err := NewManifestsFromGzippedString(compressed)
			Expect(err).NotTo(HaveOccurred())
			Expect(newManifests).To(Equal(manifests))
		})
	})
})
