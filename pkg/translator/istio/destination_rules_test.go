package istio

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("DestinationRules", func() {
	It("creates one desintaion rule per host, with one subset per unique set of labels", func() {
		inputUpstreams := inputs.BookInfoUpstrams()
		dests, err := destinationRulesFromUpstreams("ns", inputUpstreams)
		Expect(err).NotTo(HaveOccurred())
		Expect(dests).To(HaveLen(4))
		assertDestinationRule(dests[0], "details.default.svc.cluster.local", []map[string]string{
			{
				"app": "details",
			},
			{
				"version": "v1",
				"app":     "details",
			},
		})
		assertDestinationRule(dests[1], "productpage.default.svc.cluster.local", []map[string]string{
			{
				"app": "productpage",
			},
			{
				"version": "v1",
				"app":     "productpage",
			},
		})
		assertDestinationRule(dests[2], "ratings.default.svc.cluster.local", []map[string]string{
			{
				"app": "ratings",
			},
			{
				"version": "v1",
				"app":     "ratings",
			},
		})
		assertDestinationRule(dests[3], "reviews.default.svc.cluster.local", []map[string]string{
			{
				"app": "reviews",
			},
			{
				"version": "v1",
				"app":     "reviews",
			},
			{
				"version": "v2",
				"app":     "reviews",
			},
			{
				"version": "v3",
				"app":     "reviews",
			},
		})
	})
})

func assertDestinationRule(dr *v1alpha3.DestinationRule, host string, labelSets []map[string]string) {
	Expect(dr.Metadata.Name).To(Equal(host))
	Expect(dr.Host).To(Equal(host))
	Expect(dr.Subsets).To(HaveLen(len(labelSets)))
	for i, labels := range labelSets {
		Expect(dr.Subsets[i].Name).To(Equal(subsetName(labels)))
		Expect(dr.Subsets[i].Labels).To(Equal(labels))
	}
}
