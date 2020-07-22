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
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_MESH_GOLDENS = false

var _ = Describe("Mesh Table Printer", func() {
	const meshGoldenDir = "mesh"
	var runTest = func(fileName string, meshes []*smh_discovery.Mesh) {
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
			[]*smh_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "istio-mesh-1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
							Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
								Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
									Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
										InstallationNamespace: "istio-system",
										Version:               "1.5.1",
									},
									CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{
										TrustDomain:           "cluster.local",
										CitadelNamespace:      "istio-system",
										CitadelServiceAccount: "istiod",
									},
								},
							},
						},
						Cluster: &smh_core_types.ResourceRef{
							Name: "cluster-1",
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "linkerd-mesh-1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Linkerd{
							Linkerd: &smh_discovery_types.MeshSpec_LinkerdMesh{
								Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
									InstallationNamespace: "linkerd",
									Version:               "2.7.1",
								},
							},
						},
						Cluster: &smh_core_types.ResourceRef{
							Name: "cluster-1",
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "appmesh-mesh-1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{
							AwsAppMesh: &smh_discovery_types.MeshSpec_AwsAppMesh{
								Name:         "mesh-name",
								Region:       "us-east-2",
								AwsAccountId: "1234567",
								Clusters:     []string{"cluster-1", "cluster-2"},
							},
						},
					},
				},
			},
		),
	)
})
