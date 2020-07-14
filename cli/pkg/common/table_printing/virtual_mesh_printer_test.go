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
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_VIRTUAL_MESH_GOLDENS = false

var _ = Describe("Virtual Mesh Table Printer", func() {
	const tpGoldenDirectory = "virtual_mesh"
	var runTest = func(fileName string, virtualMeshes []*smh_networking.VirtualMesh) {
		goldenFilename := test_goldens.GoldenFilePath(tpGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewVirtualMeshPrinter(table_printing.DefaultTableBuilder()).Print(output, virtualMeshes)

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
			[]*smh_networking.VirtualMesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "vm-1",
						Namespace: "service-mesh-hub",
					},
					Spec: smh_networking_types.VirtualMeshSpec{
						DisplayName: "my favorite virtual mesh",
						Meshes: []*smh_core_types.ResourceRef{
							{Name: "mesh-1"},
							{Name: "mesh-2"},
							{Name: "mesh-3"},
						},
						TrustModel: &smh_networking_types.VirtualMeshSpec_Limited{
							Limited: &smh_networking_types.VirtualMeshSpec_LimitedTrust{},
						},
						CertificateAuthority: &smh_networking_types.VirtualMeshSpec_CertificateAuthority{
							Type: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
								Builtin: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
									RsaKeySizeBytes: 2048,
									OrgName:         "my-org",
								},
							},
						},
						EnforceAccessControl: smh_networking_types.VirtualMeshSpec_ENABLED,
					},
					Status: smh_networking_types.VirtualMeshStatus{
						FederationStatus: &smh_core_types.Status{
							State:   smh_core_types.Status_CONFLICT,
							Message: "This is a conflict",
						},
						ConfigStatus: &smh_core_types.Status{
							State:   smh_core_types.Status_ACCEPTED,
							Message: "This should not be printed",
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "vm-2",
						Namespace: "service-mesh-hub",
					},
					Spec: smh_networking_types.VirtualMeshSpec{
						DisplayName: "not as cool of a virtual mesh",
						Meshes: []*smh_core_types.ResourceRef{
							{Name: "mesh-1"},
							{Name: "mesh-2"},
							{Name: "mesh-3"},
						},
						CertificateAuthority: &smh_networking_types.VirtualMeshSpec_CertificateAuthority{
							Type: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Provided_{
								Provided: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Provided{
									Certificate: &smh_core_types.ResourceRef{
										Name:      "my-magic-cert",
										Namespace: "default",
									},
								},
							},
						},
						EnforceAccessControl: smh_networking_types.VirtualMeshSpec_MESH_DEFAULT,
					},
					Status: smh_networking_types.VirtualMeshStatus{
						CertificateStatus: &smh_core_types.Status{
							State:   smh_core_types.Status_INVALID,
							Message: "This is invalid",
						},
						AccessControlEnforcementStatus: &smh_core_types.Status{
							State:   smh_core_types.Status_PROCESSING_ERROR,
							Message: "This is a processing error",
						},
					},
				},
			},
		),
	)
})
