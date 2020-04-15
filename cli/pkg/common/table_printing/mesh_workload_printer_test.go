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
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_MESH_WORKLOAD_GOLDENS = false

var _ = Describe("Mesh Workload Table Printer", func() {
	const workloadGoldenDirectory = "workload"
	var runTest = func(fileName string, meshWorkloads []*zephyr_discovery.MeshWorkload) {
		goldenFilename := test_goldens.GoldenFilePath(workloadGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewMeshWorkloadPrinter(table_printing.DefaultTableBuilder()).
			Print(output, meshWorkloads)

		if UPDATE_MESH_WORKLOAD_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("Mesh Workload printer", runTest,
		Entry(
			"can print multiple mesh workloads",
			"workloads",
			[]*zephyr_discovery.MeshWorkload{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "istio-mesh-1",
					},
					Spec: zephyr_discovery_types.MeshWorkloadSpec{
						KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
							KubeControllerRef: &zephyr_core_types.ResourceRef{
								Name:      "deployment",
								Namespace: "default",
								Cluster:   "management-plane",
							},
							ServiceAccountName: "service-account-1",
						},
						Mesh: &zephyr_core_types.ResourceRef{
							Name: "istio-1",
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "linkerd-mesh-1",
					},
					Spec: zephyr_discovery_types.MeshWorkloadSpec{
						KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
							KubeControllerRef: &zephyr_core_types.ResourceRef{
								Name:      "deployment",
								Namespace: "default",
								Cluster:   "remote-cluster",
							},
							Labels: map[string]string{
								"version": "v1",
								"app":     "test",
							},
						},
						Mesh: &zephyr_core_types.ResourceRef{
							Name: "istio-1",
						}},
				},
			},
		),
	)
})
