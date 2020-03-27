package table_printing_test

import (
	"bytes"
	"io/ioutil"
	"os"

	types2 "github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/cli/pkg/common/table_printing"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_TRAFFIC_POLICY_GOLDENS = false

var _ = Describe("Traffic Policy Table Printer", func() {
	var runTest = func(fileName string, printMode table_printing.PrintMode, trafficPolicies []*v1alpha1.TrafficPolicy) {
		goldenContents, err := ioutil.ReadFile("./test_goldens/traffic_policy/" + fileName)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewTrafficPolicyPrinter(table_printing.DefaultTableBuilder).Print(output, printMode, trafficPolicies)

		if UPDATE_TRAFFIC_POLICY_GOLDENS {
			err = ioutil.WriteFile("./test_goldens/traffic_policy/"+fileName, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("Traffic Policy printer", runTest,
		Entry("can print multiple complex traffic policies", "multiple_complex_traffic_policies.txt", table_printing.ServicePrintMode, []*v1alpha1.TrafficPolicy{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "simple",
				},
				Spec: types.TrafficPolicySpec{
					DestinationSelector: &core_types.Selector{
						Refs: []*core_types.ResourceRef{{
							Cluster:   &types2.StringValue{Value: "management-plane-cluster"},
							Name:      "reviews",
							Namespace: "default",
						}},
					},
					TrafficShift: &types.MultiDestination{
						Destinations: []*types.MultiDestination_WeightedDestination{
							{
								Destination: &core_types.ResourceRef{
									Cluster:   &types2.StringValue{Value: "target-cluster"},
									Name:      "reviews",
									Namespace: "default",
								},
								Weight: 80,
							},
							{
								Destination: &core_types.ResourceRef{
									Cluster:   &types2.StringValue{Value: "management-plane-cluster"},
									Name:      "reviews",
									Namespace: "default",
								},
								Weight: 20,
								Subset: map[string]string{
									"version": "v3",
								},
							},
						},
					},
					HttpRequestMatchers: []*types.HttpMatcher{
						{
							PathSpecifier: &types.HttpMatcher_Prefix{
								Prefix: "/static",
							},
							QueryParameters: []*types.QueryParameterMatcher{
								{
									Name:  "param1",
									Value: "value1",
								},
								{
									Name:  "param2",
									Value: "^[^a-z]",
									Regex: true,
								},
							},
							Headers: []*types.HeaderMatcher{
								{
									Name:        "header1",
									Value:       "value1",
									Regex:       false,
									InvertMatch: true,
								},
								{
									Name:  "header2",
									Value: "value2*",
									Regex: true,
								},
							},
						},
					},
					FaultInjection: &types.FaultInjection{
						Percentage: 25,
						FaultInjectionType: &types.FaultInjection_Delay_{
							Delay: &types.FaultInjection_Delay{
								HttpDelayType: &types.FaultInjection_Delay_FixedDelay{
									FixedDelay: &types2.Duration{
										Seconds: 3,
									},
								},
							},
						},
					},
					RequestTimeout: &types2.Duration{
						Seconds: 4,
					},
					Retries: &types.RetryPolicy{
						Attempts: 10,
					},
					CorsPolicy: &types.CorsPolicy{
						AllowHeaders: []string{"x-auth"},
						AllowMethods: []string{"get"},
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "simple",
				},
				Spec: types.TrafficPolicySpec{
					DestinationSelector: &core_types.Selector{
						Refs: []*core_types.ResourceRef{{
							Cluster:   &types2.StringValue{Value: "management-plane-cluster"},
							Name:      "reviews",
							Namespace: "default",
						}},
					},
					TrafficShift: &types.MultiDestination{
						Destinations: []*types.MultiDestination_WeightedDestination{
							{
								Destination: &core_types.ResourceRef{
									Cluster:   &types2.StringValue{Value: "target-cluster"},
									Name:      "reviews",
									Namespace: "default",
								},
								Weight: 80,
							},
							{
								Destination: &core_types.ResourceRef{
									Cluster:   &types2.StringValue{Value: "management-plane-cluster"},
									Name:      "reviews",
									Namespace: "default",
								},
								Weight: 20,
								Subset: map[string]string{
									"version": "v3",
								},
							},
						},
					},
					HttpRequestMatchers: []*types.HttpMatcher{
						{
							PathSpecifier: &types.HttpMatcher_Prefix{
								Prefix: "/static",
							},
							QueryParameters: []*types.QueryParameterMatcher{
								{
									Name:  "param1",
									Value: "value1",
								},
								{
									Name:  "param2",
									Value: "^[^a-z]",
									Regex: true,
								},
							},
							Headers: []*types.HeaderMatcher{
								{
									Name:        "header1",
									Value:       "value1",
									Regex:       false,
									InvertMatch: true,
								},
								{
									Name:  "header2",
									Value: "value2*",
									Regex: true,
								},
							},
						},
					},
					Mirror: &types.Mirror{
						Percentage: 10,
						Destination: &core_types.ResourceRef{
							Name:      "other-svc",
							Namespace: "other-ns",
							Cluster:   &types2.StringValue{Value: "other-cluster"},
						},
					},
				},
			},
		}),
	)
})
