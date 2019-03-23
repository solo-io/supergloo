package prometheus_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	. "github.com/solo-io/supergloo/pkg/stats/prometheus"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("ensure/remove scrape configs from a target prom conig", func() {
	var client v1.PrometheusConfigClient
	BeforeEach(func() {
		var err error
		client, err = v1.NewPrometheusConfigClient(&factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()})
		Expect(err).NotTo(HaveOccurred())
	})
	meshId := "@MESHID="

	var _ = Describe("AddScrapeConfigs", func() {
		Context("valid configmap, new scrape configs", func() {
			It("adds any missing scrape configs to the target prometheus configmap", func() {
				input, err := client.Write(inputs.PrometheusConfig("name", "namespace"), clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())

				scs := inputs.InputIstioPrometheusScrapeConfigs()

				// add the scrape configs
				err = EnsureScrapeConfigs(context.TODO(), meshId, input.Metadata.Ref(), scs, client)
				Expect(err).NotTo(HaveOccurred())

				updated, err := client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				promCfg, err := prometheus.ConfigFromResource(updated)
				Expect(err).NotTo(HaveOccurred())
				for _, sc := range scs {
					sc.JobName = meshId + "-" + sc.JobName
					Expect(promCfg.ScrapeConfigs).To(ContainElement(sc))
				}
			})
		})
		Context("valid configmap, no new scrape configs", func() {
			It("performs a no-op", func() {
				input, err := client.Write(inputs.PrometheusConfig("name", "namespace"), clients.WriteOpts{})
				Expect(err).NotTo(HaveOccurred())
				scs := inputs.InputIstioPrometheusScrapeConfigs()
				err = EnsureScrapeConfigs(context.TODO(), meshId, input.Metadata.Ref(), scs, client)
				Expect(err).NotTo(HaveOccurred())
				updated, err := client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				promCfg, err := prometheus.ConfigFromResource(updated)
				Expect(err).NotTo(HaveOccurred())
				for _, sc := range scs {
					expected := *sc
					expected.JobName = meshId + "-" + expected.JobName
					Expect(promCfg.ScrapeConfigs).To(ContainElement(&expected))
				}
				// run the updater again, expect equality
				err = EnsureScrapeConfigs(context.TODO(), meshId, input.Metadata.Ref(), scs, client)
				Expect(err).NotTo(HaveOccurred())
				secondUpdate, err := client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(secondUpdate).To(Equal(updated))

			})
		})
	})

	var _ = Describe("RemoveScrapeConfigs", func() {
		It("removes all the scrape configs with the given prefix from the target prom cfg", func() {
			input, err := client.Write(inputs.PrometheusConfig("name", "namespace"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			scs := inputs.InputIstioPrometheusScrapeConfigs()

			// add our custom scrape configs
			err = EnsureScrapeConfigs(context.TODO(), meshId, input.Metadata.Ref(), scs, client)
			Expect(err).NotTo(HaveOccurred())
			updated, err := client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			promCfg, err := prometheus.ConfigFromResource(updated)
			Expect(err).NotTo(HaveOccurred())

			// configs are present
			for _, sc := range scs {
				sc.JobName = meshId + "-" + sc.JobName
				Expect(promCfg.ScrapeConfigs).To(ContainElement(sc))
			}

			// remove configs
			err = RemoveScrapeConfigs(context.TODO(), meshId, input.Metadata.Ref(), client)
			Expect(err).NotTo(HaveOccurred())

			// configs are gone
			updated, err = client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			promCfg, err = prometheus.ConfigFromResource(updated)
			Expect(err).NotTo(HaveOccurred())
			// configs are present
			for _, sc := range scs {
				sc.JobName = meshId + "-" + sc.JobName
				Expect(promCfg.ScrapeConfigs).NotTo(ContainElement(sc))
			}

		})
	})
})
