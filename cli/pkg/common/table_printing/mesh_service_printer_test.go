package table_printing_test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/test_goldens"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_MESH_SERVICE_GOLDENS = false

var _ = Describe("Mesh Service Table Printer", func() {
	const serviceGoldenDirectory = "service"
	var runTest = func(fileName string, meshServices []*discovery_v1alpha1.MeshService) {
		goldenFilename := test_goldens.GoldenFilePath(serviceGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewMeshServicePrinter(table_printing.DefaultTableBuilder()).
			Print(output, meshServices)

		if UPDATE_MESH_SERVICE_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("Mesh Services printer", runTest,
		Entry(
			"can print multiple mesh services",
			"services",
			[]*discovery_v1alpha1.MeshService{
				{
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							Ref: &core_types.ResourceRef{
								Name:      "mesh-service-1",
								Namespace: "default",
								Cluster:   "cluster-1",
							},
							WorkloadSelectorLabels: map[string]string{
								"hello": "world",
							},
							Labels: map[string]string{
								"foo": "bar",
							},
							Ports: []*discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "fake",
									Protocol: "HTTP4",
								},
								{
									Port:     8080,
									Name:     "fake-2",
									Protocol: "UDP7",
								},
							},
						},
						Mesh: &core_types.ResourceRef{
							Name: "istio-mesh-1",
						},
						Subsets: map[string]*discovery_types.MeshServiceSpec_Subset{
							"subset-1": {
								Values: []string{"1", "2", "3"},
							},
						},
						Federation: &discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: "mcDNSname",
						},
					},
					Status: discovery_types.MeshServiceStatus{
						FederationStatus: &core_types.Status{
							State:   core_types.Status_INVALID,
							Message: "Should be printed",
						},
					},
				},
				{
					Status: discovery_types.MeshServiceStatus{
						FederationStatus: &core_types.Status{
							State:   core_types.Status_ACCEPTED,
							Message: "Should be ignored",
						},
					},
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							Ref: &core_types.ResourceRef{
								Name:      "mesh-service-2",
								Namespace: "bookunfo",
								Cluster:   "cluster-2",
							},
							WorkloadSelectorLabels: map[string]string{
								"we":       "have",
								"multiple": "selector",
								"labels":   "wooho",
							},
							Ports: []*discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     15443,
									Name:     "https",
									Protocol: "IDK",
								},
							},
						},
						Mesh: &core_types.ResourceRef{
							Name: "istio-mesh-1",
						},
						Subsets: map[string]*discovery_types.MeshServiceSpec_Subset{
							"subset-2": {
								Values: []string{"4", "5", "6"},
							},
						},
						Federation: &discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: "mcDNSname",
							FederatedToWorkloads: []*core_types.ResourceRef{
								{
									Name:      "service-1",
									Namespace: "namespace-1",
									Cluster:   "cluster-1",
								},
								{
									Name:      "service-2",
									Namespace: "namespace-2",
									Cluster:   "cluster-2",
								},
							},
						},
					},
				},
			},
		),
	)
})
