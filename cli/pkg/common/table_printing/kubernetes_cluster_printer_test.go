package table_printing_test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/cli/pkg/common/table_printing"
	"github.com/solo-io/mesh-projects/cli/pkg/common/table_printing/test_goldens"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_CLUSTER_GOLDENS = false

var _ = Describe("KubernetesCluster Table Printer", func() {
	const clusterGoldenDirectory = "cluster"
	var runTest = func(fileName string, clusters []*discovery_v1alpha1.KubernetesCluster) {
		goldenFilename := test_goldens.GoldenFilePath(clusterGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewKubernetesClusterPrinter(table_printing.DefaultTableBuilder()).Print(output, clusters)

		if UPDATE_CLUSTER_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("Kubernetes Cluster printer", runTest,
		Entry(
			"can print multiple kuberenetes clusters",
			"multi_cluster",
			[]*discovery_v1alpha1.KubernetesCluster{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "management-plane",
					},
					Spec: discovery_types.KubernetesClusterSpec{
						SecretRef: &core_types.ResourceRef{
							Name:      "management-plane",
							Namespace: "service-mesh-hub",
						},
						Version:        "1.16.1",
						WriteNamespace: "service-mesh-hub",
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "remote-cluster",
					},
					Spec: discovery_types.KubernetesClusterSpec{
						SecretRef: &core_types.ResourceRef{
							Name:      "remote-cluster",
							Namespace: "service-mesh-hub",
						},
						Version:        "1.15.1",
						WriteNamespace: "service-mesh-hub",
					},
				},
			},
		),
	)
})
