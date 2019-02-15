package istio

import (
	"context"

	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("Translator", func() {
	It("gahagafaga", func() {
		t := NewTranslator("hi", nil)
		meshConfig, resourceErrs, err := t.Translate(context.TODO(), &v1.ConfigSnapshot{
			Upstreams:    map[string]gloov1.UpstreamList{"": inputs.BookInfoUpstrams()},
			Routingrules: map[string]v1.RoutingRuleList{"": inputs.BookInfoRoutingRules("")},
		})
		Expect(meshConfig).NotTo(HaveOccurred())
		Expect(resourceErrs).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("appliesToDestination", func() {
	Context("upstream selector match", func() {
		It("returns true", func() {
			applies, err := appliesToDestination("details.default.svc.cluster.local", &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "gloo-system"},
						},
					},
				},
			}, inputs.BookInfoUpstrams())
			Expect(err).NotTo(HaveOccurred())
			Expect(applies).To(BeTrue())
		})
	})
	Context("namespace selector match", func() {
		It("returns true", func() {
			applies, err := appliesToDestination("details.default.svc.cluster.local", &v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default"},
					},
				},
			}, inputs.BookInfoUpstrams())
			Expect(err).NotTo(HaveOccurred())
			Expect(applies).To(BeTrue())
		})
	})
	Context("label selector match", func() {
		It("returns true", func() {
			applies, err := appliesToDestination("details.default.svc.cluster.local", &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{"version": "v1", "app": "details"},
					},
				},
			}, inputs.BookInfoUpstrams())
			Expect(err).NotTo(HaveOccurred())
			Expect(applies).To(BeTrue())
		})
	})
})

var _ = Describe("labelSetsForSelector", func() {
	Context("PodSelector_UpstreamSelector", func() {
		It("returns labels for each upstream found", func() {
			labelSets, err := labelSetsFromSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "gloo-system"},
							{Name: "default-reviews-v2-9080", Namespace: "gloo-system"},
							{Name: "default-reviews-9080", Namespace: "gloo-system"},
						},
					},
				},
			}, inputs.BookInfoUpstrams())
			Expect(err).NotTo(HaveOccurred())
			Expect(labelSets).To(Equal([]map[string]string{
				{"version": "v1", "app": "details"},
				{"app": "reviews", "version": "v2"},
				{"app": "reviews"},
			}))
		})
	})
	Context("PodSelector_NamespaceSelector", func() {
		It("returns labels for each upstream in the namespace", func() {
			labelSets, err := labelSetsFromSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default"},
					},
				},
			}, inputs.BookInfoUpstrams())
			Expect(err).NotTo(HaveOccurred())
			Expect(labelSets).To(Equal([]map[string]string{
				{"app": "details"},
				{"version": "v1", "app": "details"},
				{"app": "productpage"},
				{"app": "productpage", "version": "v1"},
				{"app": "ratings"},
				{"app": "ratings", "version": "v1"},
				{"app": "reviews"},
				{"version": "v1", "app": "reviews"},
				{"app": "reviews", "version": "v2"},
				{"version": "v3", "app": "reviews"},
			}))
		})
	})
})

var _ = Describe("convertMatcher", func() {
	It("converts a gloo match to an istio match", func() {
		istioMatch := convertMatcher(map[string]string{"app": "details", "version": "v1"}, 1234, &gloov1.Matcher{
			PathSpecifier: &gloov1.Matcher_Exact{
				Exact: "hi",
			},
			Methods: []string{"GET", "ME", "OUTTA", "HERE"},
			Headers: []*gloov1.HeaderMatcher{
				{Name: "k", Value: "v", Regex: true},
				{Name: "a", Value: "z", Regex: false},
			},
		})
		Expect(istioMatch).To(Equal(&v1alpha3.HTTPMatchRequest{
			Uri: &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Exact{Exact: "hi"},
			},
			Method: &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Regex{Regex: "GET|ME|OUTTA|HERE"},
			},
			Headers: map[string]*v1alpha3.StringMatch{
				"a": {MatchType: &v1alpha3.StringMatch_Exact{Exact: "z"}},
				"k": {MatchType: &v1alpha3.StringMatch_Regex{Regex: "v"}},
			},
			Port:         1234,
			SourceLabels: map[string]string{"app": "details", "version": "v1"},
		}))
	})
})
