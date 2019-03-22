package istio_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/supergloo/pkg/stats/istio"
)

var _ = Describe("ScrapeConfigs", func() {
	It("renders scrape configs structs for istio", func() {
		scrapeConfigs, err := PrometheusScrapeConfigs("any-namespace-you-want")
		Expect(err).NotTo(HaveOccurred())
		Expect(scrapeConfigs).To(HaveLen(6))
		Expect(scrapeConfigs[0].JobName).To(Equal("istio-mesh"))
		Expect(scrapeConfigs[1].JobName).To(Equal("envoy-stats"))
		Expect(scrapeConfigs[2].JobName).To(Equal("istio-policy"))
		Expect(scrapeConfigs[3].JobName).To(Equal("istio-telemetry"))
		Expect(scrapeConfigs[4].JobName).To(Equal("pilot"))
		Expect(scrapeConfigs[5].JobName).To(Equal("galley"))
	})
})
