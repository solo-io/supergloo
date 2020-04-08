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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_MESH_GOLDENS = false

var _ = Describe("Mesh Table Printer", func() {
	const meshGoldenDir = "mesh"
	var runTest = func(fileName string, meshes []*discovery_v1alpha1.Mesh) {
		goldenFilename := test_goldens.GoldenFilePath(meshGoldenDir, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewMeshPrinter(table_printing.DefaultTableBuilder()).Print(output, meshes)

		if UPDATE_MESH_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("Mesh printer", runTest,
		Entry(
			"can print multiple meshes of different types",
			"multi_mesh",
			[]*discovery_v1alpha1.Mesh{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "istio-mesh-1",
					},
					Spec: discovery_types.MeshSpec{
						MeshType: &discovery_types.MeshSpec_Istio{
							Istio: &discovery_types.MeshSpec_IstioMesh{
								Installation: &discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: "istio-system",
									Version:               "1.5.1",
								},
								CitadelInfo: &discovery_types.MeshSpec_IstioMesh_CitadelInfo{
									TrustDomain:           "cluster.local",
									CitadelNamespace:      "istio-system",
									CitadelServiceAccount: "istiod",
								},
							},
						},
						Cluster: &core_types.ResourceRef{
							Name: "cluster-1",
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "linkerd-mesh-1",
					},
					Spec: discovery_types.MeshSpec{
						MeshType: &discovery_types.MeshSpec_Linkerd{
							Linkerd: &discovery_types.MeshSpec_LinkerdMesh{
								Installation: &discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: "linkerd",
									Version:               "2.7.1",
								},
							},
						},
						Cluster: &core_types.ResourceRef{
							Name: "cluster-1",
						},
					},
				},
			},
		),
	)
})
