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
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_TRAFFIC_POLICY_GOLDENS = false

var _ = Describe("Traffic Policy Table Printer", func() {
	const tpGoldenDirectory = "traffic_policy"
	var runTest = func(fileName string, printMode table_printing.PrintMode, trafficPolicies []*smh_networking.TrafficPolicy) {
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
			[]*smh_networking.TrafficPolicy{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: smh_networking_types.TrafficPolicySpec{
						DestinationSelector: &smh_core_types.ServiceSelector{
							ServiceSelectorType: &smh_core_types.ServiceSelector_ServiceRefs_{
								ServiceRefs: &smh_core_types.ServiceSelector_ServiceRefs{
									Services: []*smh_core_types.ResourceRef{{
										Cluster:   "management-plane-cluster",
										Name:      "reviews",
										Namespace: "default",
									}},
								},
							},
						},
						TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
							Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &smh_core_types.ResourceRef{
										Cluster:   "target-cluster",
										Name:      "reviews",
										Namespace: "default",
									},
									Weight: 80,
								},
								{
									Destination: &smh_core_types.ResourceRef{
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
						HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/static",
								},
								QueryParameters: []*smh_networking_types.TrafficPolicySpec_QueryParameterMatcher{
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
								Headers: []*smh_networking_types.TrafficPolicySpec_HeaderMatcher{
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
						FaultInjection: &smh_networking_types.TrafficPolicySpec_FaultInjection{
							Percentage: 25,
							FaultInjectionType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
								Delay: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay{
									HttpDelayType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
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
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts: 10,
						},
						CorsPolicy: &smh_networking_types.TrafficPolicySpec_CorsPolicy{
							AllowHeaders: []string{"x-auth"},
							AllowMethods: []string{"get"},
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: smh_networking_types.TrafficPolicySpec{
						DestinationSelector: &smh_core_types.ServiceSelector{
							ServiceSelectorType: &smh_core_types.ServiceSelector_ServiceRefs_{
								ServiceRefs: &smh_core_types.ServiceSelector_ServiceRefs{
									Services: []*smh_core_types.ResourceRef{{
										Cluster:   "management-plane-cluster",
										Name:      "reviews",
										Namespace: "default",
									}},
								},
							},
						},
						TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
							Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &smh_core_types.ResourceRef{
										Cluster:   "target-cluster",
										Name:      "reviews",
										Namespace: "default",
									},
									Weight: 80,
								},
								{
									Destination: &smh_core_types.ResourceRef{
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
						HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/static",
								},
								QueryParameters: []*smh_networking_types.TrafficPolicySpec_QueryParameterMatcher{
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
								Headers: []*smh_networking_types.TrafficPolicySpec_HeaderMatcher{
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
						Mirror: &smh_networking_types.TrafficPolicySpec_Mirror{
							Percentage: 10,
							Destination: &smh_core_types.ResourceRef{
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
