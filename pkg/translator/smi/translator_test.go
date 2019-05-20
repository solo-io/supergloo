package smi

import (
	accessv1alpha1 "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/apis/access/v1alpha1"
	specsv1alpha1 "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/apis/specs/v1alpha1"
	splitv1alpha1 "github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/api/external/smi/access"
	"github.com/solo-io/supergloo/api/external/smi/specs"
	"github.com/solo-io/supergloo/api/external/smi/split"
	sgaccess "github.com/solo-io/supergloo/pkg/api/external/smi/access/v1alpha1"
	sgspecs "github.com/solo-io/supergloo/pkg/api/external/smi/specs/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/smi/split/v1alpha1"
	"github.com/solo-io/supergloo/test/inputs"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					TrafficSplit: split.TrafficSplit{
						ObjectMeta: v1.ObjectMeta{
							Name:      "trafficshifting-productpage-reviews.default.svc.cluster.local",
							Namespace: "default",
						},
						Spec: splitv1alpha1.TrafficSplitSpec{
							Service: "reviews.default.svc.cluster.local",
							Backends: []splitv1alpha1.TrafficSplitBackend{
								{
									Service: "reviews.default.svc.cluster.local",
									Weight:  *resource.NewMilliQuantity(1000, resource.DecimalSI),
								},
							},
						},
					},
				},
				&v1alpha1.TrafficSplit{
					TrafficSplit: split.TrafficSplit{
						ObjectMeta: v1.ObjectMeta{
							Name:      "trafficshifting-reviews-50-50-reviews.default.svc.cluster.local",
							Namespace: "default",
						},
						Spec: splitv1alpha1.TrafficSplitSpec{
							Service: "reviews.default.svc.cluster.local",
							Backends: []splitv1alpha1.TrafficSplitBackend{
								{
									Service: "reviews.default.svc.cluster.local",
									Weight:  *resource.NewMilliQuantity(333, resource.DecimalSI),
								},
								{
									Service: "reviews.default.svc.cluster.local",
									Weight:  *resource.NewMilliQuantity(667, resource.DecimalSI),
								},
							},
						},
					},
				},
			},
		}))
	})
})

var _ = Describe("createSecurityConfig", func() {
	It("creates traffictargets and httproutegroups", func() {
		ns := "default"
		rules := inputs.BookInfoSecurityRules(ns, nil)
		upstreams := inputs.BookInfoUpstreams(ns)
		pods := inputs.BookInfoPods(ns)
		services := inputs.BookInfoServices(ns)
		resourceErrs := make(reporter.ResourceErrors)
		sc := createSecurityConfig(rules, upstreams, pods, services, resourceErrs)
		Expect(sc).To(Equal(SecurityConfig{
			TrafficTargets: sgaccess.TrafficTargetList{
				{
					TrafficTarget: access.TrafficTarget{
						ObjectMeta: v1.ObjectMeta{
							Name:      "details-v1-pod-1",
							Namespace: "default",
						},
						Destination: accessv1alpha1.IdentityBindingSubject{
							Kind:      "ServiceAccount",
							Name:      "details-v1-pod-1",
							Namespace: "default",
							Port:      "http",
						},
						Sources: []accessv1alpha1.IdentityBindingSubject{
							{
								Kind:      "ServiceAccount",
								Name:      "productpage-v1-pod-1",
								Namespace: "default",
								Port:      "",
							},
						},
						Specs: []accessv1alpha1.TrafficTargetSpec{
							{
								Kind: "HTTPRouteGroup",
								Name: "allow-productpage-to-details",
								Matches: []string{
									"/details/*",
								},
							},
						},
					},
				},
				{
					TrafficTarget: access.TrafficTarget{
						ObjectMeta: v1.ObjectMeta{
							Name:      "ratings-v1-pod-1",
							Namespace: "default",
						},
						Destination: accessv1alpha1.IdentityBindingSubject{
							Kind:      "ServiceAccount",
							Name:      "ratings-v1-pod-1",
							Namespace: "default",
							Port:      "http",
						},
						Sources: []accessv1alpha1.IdentityBindingSubject{
							{
								Kind:      "ServiceAccount",
								Name:      "productpage-v1-pod-1",
								Namespace: "default",
								Port:      "",
							},
						},
						Specs: []accessv1alpha1.TrafficTargetSpec{
							{
								Kind: "HTTPRouteGroup",
								Name: "allow-productpage-to-ratings",
								Matches: []string{
									"/ratings/*",
								},
							},
						},
					},
				},
				{
					TrafficTarget: access.TrafficTarget{
						ObjectMeta: v1.ObjectMeta{
							Name:      "reviews-v1-pod-1",
							Namespace: "default",
						},
						Destination: accessv1alpha1.IdentityBindingSubject{
							Kind:      "ServiceAccount",
							Name:      "reviews-v1-pod-1",
							Namespace: "default",
							Port:      "http",
						},
						Sources: []accessv1alpha1.IdentityBindingSubject{
							{
								Kind:      "ServiceAccount",
								Name:      "ratings-v1-pod-1",
								Namespace: "default",
								Port:      "",
							},
						},
						Specs: []accessv1alpha1.TrafficTargetSpec{
							{
								Kind: "HTTPRouteGroup",
								Name: "allow-ratings-to-reviews",
								Matches: []string{
									"/reviews/*",
								},
							},
						},
					},
				},
				{
					TrafficTarget: access.TrafficTarget{
						ObjectMeta: v1.ObjectMeta{
							Name:      "reviews-v2-pod-1",
							Namespace: "default",
						},
						Destination: accessv1alpha1.IdentityBindingSubject{
							Kind:      "ServiceAccount",
							Name:      "reviews-v2-pod-1",
							Namespace: "default",
							Port:      "http",
						},
						Sources: []accessv1alpha1.IdentityBindingSubject{
							{
								Kind:      "ServiceAccount",
								Name:      "ratings-v1-pod-1",
								Namespace: "default",
							},
						},
						Specs: []accessv1alpha1.TrafficTargetSpec{
							{
								Kind:    "HTTPRouteGroup",
								Name:    "allow-ratings-to-reviews",
								Matches: []string{"/reviews/*"},
							},
						},
					},
				},
				{
					TrafficTarget: access.TrafficTarget{
						ObjectMeta: v1.ObjectMeta{
							Name:      "reviews-v3-pod-1",
							Namespace: "default",
						},
						Destination: accessv1alpha1.IdentityBindingSubject{
							Kind:      "ServiceAccount",
							Name:      "reviews-v3-pod-1",
							Namespace: "default",
							Port:      "http",
						},
						Sources: []accessv1alpha1.IdentityBindingSubject{
							{
								Kind:      "ServiceAccount",
								Name:      "ratings-v1-pod-1",
								Namespace: "default",
							},
						},
						Specs: []accessv1alpha1.TrafficTargetSpec{
							{
								Kind:    "HTTPRouteGroup",
								Name:    "allow-ratings-to-reviews",
								Matches: []string{"/reviews/*"},
							},
						},
					},
				},
			},
			HTTPRouteGroups: sgspecs.HTTPRouteGroupList{
				{
					HTTPRouteGroup: specs.HTTPRouteGroup{
						ObjectMeta: v1.ObjectMeta{
							Name:      "allow-productpage-to-details",
							Namespace: "default",
						},
						Matches: []specsv1alpha1.HTTPMatch{
							{
								Name: "/details/*",
								Methods: []string{
									"GET",
								},
								PathRegex: "/details/*",
							},
						},
					},
				},
				{
					HTTPRouteGroup: specs.HTTPRouteGroup{
						ObjectMeta: v1.ObjectMeta{
							Name:      "allow-productpage-to-ratings",
							Namespace: "default",
						},
						Matches: []specsv1alpha1.HTTPMatch{
							{
								Name: "/ratings/*",
								Methods: []string{
									"GET",
								},
								PathRegex: "/ratings/*",
							},
						},
					},
				},
				{
					HTTPRouteGroup: specs.HTTPRouteGroup{
						ObjectMeta: v1.ObjectMeta{
							Name:      "allow-ratings-to-reviews",
							Namespace: "default",
						},
						Matches: []specsv1alpha1.HTTPMatch{
							{
								Name: "/reviews/*",
								Methods: []string{
									"GET",
								},
								PathRegex: "/reviews/*",
							},
						},
					},
				},
			},
		}))
	})
})
