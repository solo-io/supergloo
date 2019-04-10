package istio

import (
	"context"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/helmchart"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("makeManifestsForInstall", func() {
	type testCase struct {
		installNs       string
		version         string
		existingInstall *v1.Mesh
		istioPrefs      *v1.IstioInstall
	}
	test := func(c testCase) (helmchart.Manifests, error) {
		install := inputs.IstioInstall("test", "mesh", c.installNs, c.version, false)
		if c.istioPrefs != nil {
			c.istioPrefs.IstioVersion = c.version
			install.GetMesh().MeshInstallType = &v1.MeshInstall_IstioMesh{IstioMesh: c.istioPrefs}
		}
		if c.existingInstall != nil {
			ref := c.existingInstall.Metadata.Ref()
			install.GetMesh().InstalledMesh = &ref
		}
		return makeManifestsForInstall(context.TODO(), install, c.existingInstall, install.GetMesh().GetIstioMesh())
	}
	Context("invalid opts", func() {
		It("errors on version", func() {
			_, err := test(testCase{
				version:   "badver",
				installNs: "ok",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a suppported istio version"))
		})
		It("errors on namespace", func() {
			_, err := test(testCase{
				version: IstioVersion103,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must provide installation namespace"))
		})
	})
	Context("valid opts", func() {
		It("renders charts with the correct manifests for the values", func() {
			manifests, err := test(testCase{
				version:   IstioVersion103,
				installNs: "ok",
				istioPrefs: &v1.IstioInstall{
					InstallJaeger: true,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(manifests).NotTo(BeEmpty())
			Expect(manifestsHaveKey(manifests, "istio/charts/tracing/templates/service-jaeger.yaml")).To(BeTrue())

			manifests, err = test(testCase{
				version:   IstioVersion103,
				installNs: "ok",
				istioPrefs: &v1.IstioInstall{
					InstallJaeger: false,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(manifests).NotTo(BeEmpty())
			Expect(manifestsHaveKey(manifests, "istio/charts/tracing/templates/service-jaeger.yaml")).To(BeFalse())
		})
		It("sets selfsigned to be false if the preinstalled mesh has a custom cert", func() {
			manifests, err := test(testCase{
				version:   IstioVersion103,
				installNs: "ok",
				istioPrefs: &v1.IstioInstall{
					InstallJaeger: true,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(manifests).NotTo(BeEmpty())
			Expect(manifestsHaveContent(manifests, "--self-signed-ca=true")).To(BeTrue())

			manifests, err = test(testCase{
				version:   IstioVersion103,
				installNs: "ok",
				istioPrefs: &v1.IstioInstall{
					InstallJaeger: true,
				},
				existingInstall: inputs.IstioMesh("ok", &core.ResourceRef{"some", "secret"}),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(manifests).NotTo(BeEmpty())
			Expect(manifestsHaveContent(manifests, "--self-signed-ca=false")).To(BeTrue())
		})
	})
})

func manifestsHaveKey(manifests helmchart.Manifests, key string) bool {
	for _, man := range manifests {
		if man.Name == key {
			return true
		}
	}
	return false
}

func manifestsHaveContent(manifests helmchart.Manifests, content string) bool {
	for _, man := range manifests {
		if strings.Contains(man.Content, content) {
			return true
		}
	}
	return false
}
