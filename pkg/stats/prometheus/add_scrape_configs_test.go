package prometheus_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/prometheus"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/pkg/stats/prometheus"
)

var _ = Describe("AddScrapeConfigs", func() {
	Context("valid configmap, new scrape configs", func() {
		It("adds any missing scrape configs to the target prometheus configmap", func() {
			client, err := v1.NewPrometheusConfigClient(&factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()})
			Expect(err).NotTo(HaveOccurred())
			input, err := client.Write(inputs.InputPrometheusConfig("name", "namespace"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			scs := inputs.InputIstioPrometheusScrapeConfigs()
			updater := NewConfigUpdater(input.Metadata.Ref(), scs, client)
			err = updater.EnsureScrapeConfigs(context.TODO())
			Expect(err).NotTo(HaveOccurred())
			updated, err := client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			promCfg, err := prometheus.ConfigFromResource(updated)
			Expect(err).NotTo(HaveOccurred())
			for _, sc := range scs {
				Expect(promCfg.ScrapeConfigs).To(ContainElement(sc))
			}
		})
	})
	Context("valid configmap, no new scrape configs", func() {
		It("performs a no-op", func() {
			client, err := v1.NewPrometheusConfigClient(&factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()})
			Expect(err).NotTo(HaveOccurred())
			input, err := client.Write(inputs.InputPrometheusConfig("name", "namespace"), clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			scs := inputs.InputIstioPrometheusScrapeConfigs()
			updater := NewConfigUpdater(input.Metadata.Ref(), scs, client)
			err = updater.EnsureScrapeConfigs(context.TODO())
			Expect(err).NotTo(HaveOccurred())
			updated, err := client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			promCfg, err := prometheus.ConfigFromResource(updated)
			Expect(err).NotTo(HaveOccurred())
			for _, sc := range scs {
				Expect(promCfg.ScrapeConfigs).To(ContainElement(sc))
			}
			// run the updater again, expect equality
			err = updater.EnsureScrapeConfigs(context.TODO())
			Expect(err).NotTo(HaveOccurred())
			secondUpdate, err := client.Read(input.Metadata.Namespace, input.Metadata.Name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(secondUpdate).To(Equal(updated))

		})
	})
	Context("invalid configmap", func() {
		It("returns an error", func() {
			client, err := v1.NewPrometheusConfigClient(&factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()})
			Expect(err).NotTo(HaveOccurred())
			scs := inputs.InputIstioPrometheusScrapeConfigs()
			updater := NewConfigUpdater(core.ResourceRef{"doesnt", "exist"}, scs, client)
			err = updater.EnsureScrapeConfigs(context.TODO())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
		})
	})
})
