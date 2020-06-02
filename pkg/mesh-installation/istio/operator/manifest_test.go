package operator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator"
)

var _ = Describe("Manifest", func() {
	It("outputs a manifest at 1.5.x when Istio 1.5 is requested", func() {
		manifestBuilder := operator.NewInstallerManifestBuilder()
		manifest, err := manifestBuilder.BuildOperatorDeploymentManifest(operator.IstioVersion1_5, "test-ns", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest).To(ContainSubstring("image: docker.io/istio/operator:1.5"))
	})

	It("outputs a manifest at 1.6.x when Istio 1.6 is requested", func() {
		manifestBuilder := operator.NewInstallerManifestBuilder()
		manifest, err := manifestBuilder.BuildOperatorDeploymentManifest(operator.IstioVersion1_6, "test-ns", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest).To(ContainSubstring("image: docker.io/istio/operator:1.6"))
	})

	It("outputs a Namespace when requested", func() {
		manifestBuilder := operator.NewInstallerManifestBuilder()
		manifest, err := manifestBuilder.BuildOperatorDeploymentManifest(operator.IstioVersion1_6, "test-ns", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest).NotTo(ContainSubstring("kind: Namespace"))

		manifest, err = manifestBuilder.BuildOperatorDeploymentManifest(operator.IstioVersion1_6, "test-ns", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest).To(ContainSubstring("kind: Namespace"))
	})
})
