package equality_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2/types"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/virtualservice/equality"
)

var _ = Describe("MatcherEquality", func() {
	workloadSelector1a := &v1alpha2.WorkloadSelector{
		Labels:     map[string]string{"a": "b", "c": "d"},
		Namespaces: []string{"foo", "bar"},
	}
	workloadSelector1b := &v1alpha2.WorkloadSelector{
		Labels:     map[string]string{"c": "d", "a": "b"},
		Namespaces: []string{"bar", "foo"},
	}

	workloadSelector2a := &v1alpha2.WorkloadSelector{
		Labels:     map[string]string{"e": "f", "g": "h"},
		Namespaces: []string{"boo", "baz"},
	}
	workloadSelector2b := &v1alpha2.WorkloadSelector{
		Labels:     map[string]string{"g": "h", "e": "f"},
		Namespaces: []string{"baz", "boo"},
	}

	httpMatcher1 := &v1alpha2.TrafficPolicySpec_HttpMatcher{
		PathSpecifier: &v1alpha2.TrafficPolicySpec_HttpMatcher_Prefix{
			Prefix: "/foo",
		},
		Headers: []*v1alpha2.TrafficPolicySpec_HeaderMatcher{
			{
				Name:        "header-name",
				Value:       "header-value",
				Regex:       false,
				InvertMatch: false,
			},
		},
		QueryParameters: []*v1alpha2.TrafficPolicySpec_QueryParameterMatcher{
			{
				Name:  "query-param-name",
				Value: "query-param-value",
				Regex: false,
			},
		},
		Method: &v1alpha2.TrafficPolicySpec_HttpMethod{
			Method: types.HttpMethodValue_GET,
		},
	}
	httpMatcher2 := &v1alpha2.TrafficPolicySpec_HttpMatcher{
		PathSpecifier: &v1alpha2.TrafficPolicySpec_HttpMatcher_Prefix{
			Prefix: "/bar",
		},
		Headers: []*v1alpha2.TrafficPolicySpec_HeaderMatcher{
			{
				Name:        "header2-name",
				Value:       "header2-value",
				Regex:       false,
				InvertMatch: false,
			},
		},
		QueryParameters: []*v1alpha2.TrafficPolicySpec_QueryParameterMatcher{
			{
				Name:  "query-param2-name",
				Value: "query-param2-value",
				Regex: false,
			},
		},
		Method: &v1alpha2.TrafficPolicySpec_HttpMethod{
			Method: types.HttpMethodValue_HEAD,
		},
	}

	DescribeTable("should equate two TrafficPolicies with semantically equivalent request matchers",
		func(tp1 *v1alpha2.TrafficPolicySpec, tp2 *v1alpha2.TrafficPolicySpec, expected bool) {
			Expect(equality.TrafficPolicyMatchersEqual(tp1, tp2)).To(Equal(expected))
		},

		Entry(
			"no matchers",
			&v1alpha2.TrafficPolicySpec{},
			&v1alpha2.TrafficPolicySpec{},
			true,
		),

		Entry("equal workload selectors with differently ordered fields",
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector1a,
					workloadSelector2a,
				},
			},
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector2b,
					workloadSelector1b,
				},
			},
			true,
		),

		Entry("unequal workload selectors of equal length",
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector1a,
					workloadSelector1a,
				},
			},
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector2b,
					workloadSelector1b,
				},
			},
			false,
		),

		Entry("unequal workload selectors of equal length",
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector1a,
					workloadSelector2a,
				},
			},
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector2b,
					workloadSelector2b,
				},
			},
			false,
		),

		Entry("unequal workload selectors with differently ordered fields",
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector1a,
					{
						Labels:     map[string]string{"e": "f", "g": "i"}, // diff
						Namespaces: []string{"boo", "baz"},
						Clusters:   []string{"cluster3", "cluster4"},
					},
				},
			},
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector2b,
					workloadSelector1b,
				},
			},
			false,
		),

		Entry("equal http matchers",
			&v1alpha2.TrafficPolicySpec{
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					httpMatcher1,
				},
			},
			&v1alpha2.TrafficPolicySpec{
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					httpMatcher1,
				},
			},
			true,
		),

		Entry("unequal http matchers of equal length",
			&v1alpha2.TrafficPolicySpec{
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					{
						PathSpecifier: &v1alpha2.TrafficPolicySpec_HttpMatcher_Prefix{
							Prefix: "/diff", // diff
						},
						Headers: []*v1alpha2.TrafficPolicySpec_HeaderMatcher{
							{
								Name:        "header2-name",
								Value:       "header2-value",
								Regex:       false,
								InvertMatch: false,
							},
						},
						QueryParameters: []*v1alpha2.TrafficPolicySpec_QueryParameterMatcher{
							{
								Name:  "query-param2-name",
								Value: "query-param2-value",
								Regex: false,
							},
						},
						Method: &v1alpha2.TrafficPolicySpec_HttpMethod{
							Method: types.HttpMethodValue_HEAD,
						},
					},
					httpMatcher1,
				},
			},
			&v1alpha2.TrafficPolicySpec{
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					httpMatcher1,
					httpMatcher2,
				},
			},
			false,
		),

		Entry("equal multiple http matchers",
			&v1alpha2.TrafficPolicySpec{
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					httpMatcher2,
					httpMatcher1,
				},
			},
			&v1alpha2.TrafficPolicySpec{
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					httpMatcher1,
					httpMatcher2,
				},
			},
			true,
		),

		Entry("equal with both workload selectors and http matchers",
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector1a,
					workloadSelector2a,
				},
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					httpMatcher2,
					httpMatcher1,
				},
			},
			&v1alpha2.TrafficPolicySpec{
				SourceSelector: []*v1alpha2.WorkloadSelector{
					workloadSelector2b,
					workloadSelector1b,
				},
				HttpRequestMatchers: []*v1alpha2.TrafficPolicySpec_HttpMatcher{
					httpMatcher1,
					httpMatcher2,
				},
			},
			true,
		),
	)
})
