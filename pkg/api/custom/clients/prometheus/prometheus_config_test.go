package prometheus

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("PrometheusConfig", func() {
	It("adds scrape configs", func() {
		cfg := &Config{}
		scs := inputs.InputIstioPrometheusScrapeConfigs()
		added := cfg.AddScrapeConfigs(scs)
		Expect(added).To(Equal(len(scs)))
		sortConfigs(scs)
		Expect(cfg.ScrapeConfigs).To(Equal(scs))
	})
	It("removes scrape configs by prefix", func() {
		scs := inputs.InputIstioPrometheusScrapeConfigs()
		cfg := &Config{ScrapeConfigs: scs}
		removed := cfg.RemoveScrapeConfigs("istio")
		var scsWithoutIstio []*config.ScrapeConfig
		for _, sc := range scs {
			if strings.HasPrefix(sc.JobName, "istio") {
				continue
			}
			scsWithoutIstio = append(scsWithoutIstio, sc)
		}
		Expect(removed).To(Equal(len(scsWithoutIstio)))
		sortConfigs(scsWithoutIstio)
		Expect(cfg.ScrapeConfigs).To(Equal(scsWithoutIstio))
	})
	It("removes scrape configs by name", func() {
		scs := inputs.InputIstioPrometheusScrapeConfigs()
		cfg := &Config{ScrapeConfigs: scs}
		for _, sc := range scs {
			removed := cfg.RemoveScrapeConfigs(sc.JobName)
			Expect(removed).To(Equal(1))
			Expect(cfg.ScrapeConfigs).NotTo(ContainElement(sc))

		}
		Expect(cfg.ScrapeConfigs).To(BeEmpty())
	})
})

var _ = Describe("AddPrefix", func() {
	It("adds a prefix to each job name without modifying the input object", func() {
		scs := inputs.InputIstioPrometheusScrapeConfigs()
		prefixed := AddPrefix(scs, "test")
		Expect(scs).To(Equal(inputs.InputIstioPrometheusScrapeConfigs()))
		Expect(prefixed).To(HaveLen(len(scs)))
		for i, sc := range scs {
			Expect(prefixed[i].JobName).To(Equal("test" + sc.JobName))
		}
	})
})
