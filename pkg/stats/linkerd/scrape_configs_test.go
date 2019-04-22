package linkerd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/supergloo/pkg/stats/linkerd"
)

var _ = Describe("ScrapeConfigs", func() {
	It("renders scrape configs structs for linkerd", func() {
		scrapeConfigs, err := PrometheusScrapeConfigs("any-namespace-you-want")
		Expect(err).NotTo(HaveOccurred())
		Expect(scrapeConfigs).To(HaveLen(1))
		Expect(scrapeConfigs[0].JobName).To(Equal("linkerd"))
	})
})
