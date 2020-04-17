package table_printing_test

import (
	"bytes"
	"io/ioutil"
	"os"

	types2 "github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/test_goldens"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_TRAFFIC_POLICY_GOLDENS = false

var _ = Describe("Traffic Policy Table Printer", func() {
	const tpGoldenDirectory = "traffic_policy"
	var runTest = func(fileName string, printMode table_printing.PrintMode, trafficPolicies []*zephyr_networking.TrafficPolicy) {
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
			[]*zephyr_networking.TrafficPolicy{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: zephyr_networking_types.TrafficPolicySpec{
						DestinationSelector: &zephyr_core_types.ServiceSelector{
							ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
								ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
									Services: []*zephyr_core_types.ResourceRef{{
										Cluster:   "management-plane-cluster",
										Name:      "reviews",
										Namespace: "default",
									}},
								},
							},
						},
						TrafficShift: &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
							Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &zephyr_core_types.ResourceRef{
										Cluster:   "target-cluster",
										Name:      "reviews",
										Namespace: "default",
									},
									Weight: 80,
								},
								{
									Destination: &zephyr_core_types.ResourceRef{
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
						HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/static",
								},
								QueryParameters: []*zephyr_networking_types.TrafficPolicySpec_QueryParameterMatcher{
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
								Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
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
						FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
							Percentage: 25,
							FaultInjectionType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
								Delay: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay{
									HttpDelayType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
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
						Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 10,
						},
						CorsPolicy: &zephyr_networking_types.TrafficPolicySpec_CorsPolicy{
							AllowHeaders: []string{"x-auth"},
							AllowMethods: []string{"get"},
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: zephyr_networking_types.TrafficPolicySpec{
						DestinationSelector: &zephyr_core_types.ServiceSelector{
							ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
								ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
									Services: []*zephyr_core_types.ResourceRef{{
										Cluster:   "management-plane-cluster",
										Name:      "reviews",
										Namespace: "default",
									}},
								},
							},
						},
						TrafficShift: &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
							Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &zephyr_core_types.ResourceRef{
										Cluster:   "target-cluster",
										Name:      "reviews",
										Namespace: "default",
									},
									Weight: 80,
								},
								{
									Destination: &zephyr_core_types.ResourceRef{
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
						HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/static",
								},
								QueryParameters: []*zephyr_networking_types.TrafficPolicySpec_QueryParameterMatcher{
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
								Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
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
						Mirror: &zephyr_networking_types.TrafficPolicySpec_Mirror{
							Percentage: 10,
							Destination: &zephyr_core_types.ResourceRef{
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
