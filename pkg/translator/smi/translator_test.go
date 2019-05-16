package smi

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/smi/split/v1alpha1"
	"github.com/solo-io/supergloo/test/inputs"
)

var _ = Describe("createRoutingConfig", func() {
	It("creates traffic splits", func() {
		ns := "default"
		rules := inputs.AdvancedBookInfoRoutingRules(ns, nil)
		upstreams := inputs.BookInfoUpstreams(ns)
		services := inputs.BookInfoServices(ns)
		resourceErrs := make(reporter.ResourceErrors)
		rc := createRoutingConfig(rules, upstreams, services, resourceErrs)
		Expect(rc).To(Equal(RoutingConfig{
			TrafficSplits: v1alpha1.TrafficSplitList{
				&v1alpha1.TrafficSplit{
					Metadata: core.Metadata{
						Name:      "trafficshifting-productpage-reviews.default.svc.cluster.local",
						Namespace: "default",
					},
					Spec: &v1alpha1.TrafficSplitSpec{
						Service: "reviews.default.svc.cluster.local",
						Backends: []*v1alpha1.TrafficSplitBackend{
							{
								Service: "reviews.default.svc.cluster.local",
								Weight:  "1000m",
							},
						},
					},
				},
				&v1alpha1.TrafficSplit{
					Metadata: core.Metadata{
						Name:      "trafficshifting-reviews-50-50-reviews.default.svc.cluster.local",
						Namespace: "default",
					},
					Spec: &v1alpha1.TrafficSplitSpec{
						Service: "reviews.default.svc.cluster.local",
						Backends: []*v1alpha1.TrafficSplitBackend{
							{
								Service: "reviews.default.svc.cluster.local",
								Weight:  "333m",
							},
							{
								Service: "reviews.default.svc.cluster.local",
								Weight:  "667m",
							},
						},
					},
				},
			},
		}))
	})
})
