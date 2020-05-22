package table_printing_test

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/test_goldens"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_VIRTUAL_MESH_GOLDENS = false

var _ = Describe("Virtual Mesh Table Printer", func() {
	const tpGoldenDirectory = "virtual_mesh"
	var runTest = func(fileName string, virtualMeshes []*zephyr_networking.VirtualMesh, meshes []*zephyr_discovery.Mesh) {
		goldenFilename := test_goldens.GoldenFilePath(tpGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewVirtualMeshPrinter(table_printing.DefaultTableBuilder()).Print(output, virtualMeshes, meshes)

		if UPDATE_VIRTUAL_MESH_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("Virtual Mesh printer", runTest,
		Entry(
			"can print different kinds of virtual meshes",
			"virtual_mesh",
			[]*zephyr_networking.VirtualMesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "vm-1",
						Namespace: "service-mesh-hub",
					},
					Spec: zephyr_networking_types.VirtualMeshSpec{
						DisplayName: "my favorite virtual mesh",
						Meshes: []*zephyr_core_types.ResourceRef{
							{Name: "mesh-1"},
							{Name: "mesh-2"},
							{Name: "mesh-3"},
						},
						TrustModel: &zephyr_networking_types.VirtualMeshSpec_Limited{
							Limited: &zephyr_networking_types.VirtualMeshSpec_LimitedTrust{},
						},
						CertificateAuthority: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority{
							Type: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
								Builtin: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
									RsaKeySizeBytes: 2048,
									OrgName:         "my-org",
								},
							},
						},
						EnforceAccessControl: &types.BoolValue{Value: true},
					},
					Status: zephyr_networking_types.VirtualMeshStatus{
						FederationStatus: &zephyr_core_types.Status{
							State:   zephyr_core_types.Status_CONFLICT,
							Message: "This is a conflict",
						},
						ConfigStatus: &zephyr_core_types.Status{
							State:   zephyr_core_types.Status_ACCEPTED,
							Message: "This should not be printed",
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "vm-2",
						Namespace: "service-mesh-hub",
					},
					Spec: zephyr_networking_types.VirtualMeshSpec{
						DisplayName: "not as cool of a virtual mesh",
						Meshes: []*zephyr_core_types.ResourceRef{
							{Name: "mesh-1"},
							{Name: "mesh-2"},
							{Name: "mesh-3"},
						},
						CertificateAuthority: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority{
							Type: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Provided_{
								Provided: &zephyr_networking_types.VirtualMeshSpec_CertificateAuthority_Provided{
									Certificate: &zephyr_core_types.ResourceRef{
										Name:      "my-magic-cert",
										Namespace: "default",
									},
								},
							},
						},
						EnforceAccessControl: nil,
					},
					Status: zephyr_networking_types.VirtualMeshStatus{
						CertificateStatus: &zephyr_core_types.Status{
							State:   zephyr_core_types.Status_INVALID,
							Message: "This is invalid",
						},
						AccessControlEnforcementStatus: &zephyr_core_types.Status{
							State:   zephyr_core_types.Status_PROCESSING_ERROR,
							Message: "This is a processing error",
						},
					},
				},
			},
			[]*zephyr_discovery.Mesh{
				{},
				{
					Spec: types2.MeshSpec{
						MeshType: &types2.MeshSpec_Istio{},
					},
				},
			},
		),
	)
})
