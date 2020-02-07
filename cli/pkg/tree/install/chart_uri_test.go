package install_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/install"
	"github.com/solo-io/mesh-projects/pkg/version"
)

var _ = Describe("GetChartUri", func() {
	BeforeEach(func() {
		version.Version = "v1.0.0" // force the version
	})

	It("should respect version override and remove v suffix in chart uri", func() {
		versionOverride := "v2.0.0"
		chartUri, err := install.GetChartUri("", versionOverride)
		Expect(err).NotTo(HaveOccurred())
		Expect(chartUri).To(Equal(fmt.Sprintf(cliconstants.HelmRepoTemplate, strings.TrimPrefix(versionOverride, "v"))))
	})

	It("should respect chart uri override", func() {
		chartOverride := "https://chart-override.com/chart.tgz"
		chartUri, err := install.GetChartUri(chartOverride, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(chartUri).To(Equal(chartOverride))
	})

	It("should throw error when both chart and version overrides are supplied", func() {
		chartOverride := "chartOverride"
		versionOverride := "versionOverride"
		chartUri, err := install.GetChartUri(chartOverride, versionOverride)
		Expect(err).To(testutils.HaveInErrorChain(install.ChartAndReleaseFlagErr(chartOverride, versionOverride)))
		Expect(chartUri).To(BeEmpty())
	})

	It("should throw error when version is not release and no chart override provided", func() {
		version.Version = version.UndefinedVersion
		chartUri, err := install.GetChartUri("", "")
		Expect(err).To(testutils.HaveInErrorChain(install.UnreleasedWithoutOverrideErr))
		Expect(chartUri).To(BeEmpty())
	})

	It("should throw error when version is pre-release and no chart override provided", func() {
		version.Version = "v1.2.3-36-gfcf38ba"
		chartUri, err := install.GetChartUri("", "")
		Expect(err).To(testutils.HaveInErrorChain(install.UnreleasedWithoutOverrideErr))
		Expect(chartUri).To(BeEmpty())
	})

	It("should throw error when chart override doesn't have a valid Helm extension", func() {
		chartOverride := "https://chart-override.com/invalid"
		chartUri, err := install.GetChartUri(chartOverride, "")
		Expect(err).To(testutils.HaveInErrorChain(install.UnsupportedHelmFileExtErr(chartOverride)))
		Expect(chartUri).To(BeEmpty())
	})
})
