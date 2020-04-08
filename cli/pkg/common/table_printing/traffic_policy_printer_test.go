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
	"github.com/solo-io/mesh-projects/cli/pkg/common/table_printing/test_goldens"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_TRAFFIC_POLICY_GOLDENS = false

var _ = Describe("Traffic Policy Table Printer", func() {
	const tpGoldenDirectory = "traffic_policy"
	var runTest = func(fileName string, printMode table_printing.PrintMode, trafficPolicies []*v1alpha1.TrafficPolicy) {
		goldenFilename := test_goldens.GoldenFilePath(tpGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewTrafficPolicyPrinter(table_printing.DefaultTableBuilder()).Print(output, printMode, trafficPolicies)

		if UPDATE_TRAFFIC_POLICY_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("Traffic Policy printer", runTest,
		Entry(
			"can print multiple complex traffic policies",
			"multiple_complex_traffic_policies",
			table_printing.ServicePrintMode,
			[]*v1alpha1.TrafficPolicy{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "simple",
					},
					Spec: types.TrafficPolicySpec{
						DestinationSelector: &core_types.ServiceSelector{
							ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
								ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
									Services: []*core_types.ResourceRef{{
										Cluster:   "management-plane-cluster",
										Name:      "reviews",
										Namespace: "default",
									}},
								},
							},
						},
						TrafficShift: &types.TrafficPolicySpec_MultiDestination{
							Destinations: []*types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &core_types.ResourceRef{
										Cluster:   "target-cluster",
										Name:      "reviews",
										Namespace: "default",
									},
									Weight: 80,
								},
								{
									Destination: &core_types.ResourceRef{
										Cluster:   "management-plane-cluster",
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
						HttpRequestMatchers: []*types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/static",
								},
								QueryParameters: []*types.TrafficPolicySpec_QueryParameterMatcher{
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
								Headers: []*types.TrafficPolicySpec_HeaderMatcher{
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
						FaultInjection: &types.TrafficPolicySpec_FaultInjection{
							Percentage: 25,
							FaultInjectionType: &types.TrafficPolicySpec_FaultInjection_Delay_{
								Delay: &types.TrafficPolicySpec_FaultInjection_Delay{
									HttpDelayType: &types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
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
						Retries: &types.TrafficPolicySpec_RetryPolicy{
							Attempts: 10,
						},
						CorsPolicy: &types.TrafficPolicySpec_CorsPolicy{
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
						DestinationSelector: &core_types.ServiceSelector{
							ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
								ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
									Services: []*core_types.ResourceRef{{
										Cluster:   "management-plane-cluster",
										Name:      "reviews",
										Namespace: "default",
									}},
								},
							},
						},
						TrafficShift: &types.TrafficPolicySpec_MultiDestination{
							Destinations: []*types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &core_types.ResourceRef{
										Cluster:   "target-cluster",
										Name:      "reviews",
										Namespace: "default",
									},
									Weight: 80,
								},
								{
									Destination: &core_types.ResourceRef{
										Cluster:   "management-plane-cluster",
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
						HttpRequestMatchers: []*types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/static",
								},
								QueryParameters: []*types.TrafficPolicySpec_QueryParameterMatcher{
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
								Headers: []*types.TrafficPolicySpec_HeaderMatcher{
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
						Mirror: &types.TrafficPolicySpec_Mirror{
							Percentage: 10,
							Destination: &core_types.ResourceRef{
								Name:      "other-svc",
								Namespace: "other-ns",
								Cluster:   "other-cluster",
							},
						},
					},
				},
			}),
	)
})
