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
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_VIRTUAL_MESH_GOLDENS = false

var _ = Describe("Virtual Mesh Table Printer", func() {
	const tpGoldenDirectory = "virtual_mesh"
	var runTest = func(fileName string, virtualMeshes []*zephyr_networking.VirtualMesh) {
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
			[]*zephyr_networking.VirtualMesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vm-1",
						Namespace: "service-mesh-hub",
					},
					Spec: networking_types.VirtualMeshSpec{
						DisplayName: "my favorite virtual mesh",
						Meshes: []*core_types.ResourceRef{
							{Name: "mesh-1"},
							{Name: "mesh-2"},
							{Name: "mesh-3"},
						},
						TrustModel: &networking_types.VirtualMeshSpec_Limited{
							Limited: &networking_types.VirtualMeshSpec_LimitedTrust{},
						},
						CertificateAuthority: &networking_types.VirtualMeshSpec_CertificateAuthority{
							Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
								Builtin: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
									RsaKeySizeBytes: 2048,
									OrgName:         "my-org",
								},
							},
						},
						EnforceAccessControl: true,
					},
					Status: networking_types.VirtualMeshStatus{
						FederationStatus: &core_types.Status{
							State:   core_types.Status_CONFLICT,
							Message: "This is a conflict",
						},
						ConfigStatus: &core_types.Status{
							State:   core_types.Status_ACCEPTED,
							Message: "This should not be printed",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vm-2",
						Namespace: "service-mesh-hub",
					},
					Spec: networking_types.VirtualMeshSpec{
						DisplayName: "not as cool of a virtual mesh",
						Meshes: []*core_types.ResourceRef{
							{Name: "mesh-1"},
							{Name: "mesh-2"},
							{Name: "mesh-3"},
						},
						CertificateAuthority: &networking_types.VirtualMeshSpec_CertificateAuthority{
							Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Provided_{
								Provided: &networking_types.VirtualMeshSpec_CertificateAuthority_Provided{
									Certificate: &core_types.ResourceRef{
										Name:      "my-magic-cert",
										Namespace: "default",
									},
								},
							},
						},
						EnforceAccessControl: false,
					},
					Status: networking_types.VirtualMeshStatus{
						CertificateStatus: &core_types.Status{
							State:   core_types.Status_INVALID,
							Message: "This is invalid",
						},
						AccessControlEnforcementStatus: &core_types.Status{
							State:   core_types.Status_PROCESSING_ERROR,
							Message: "This is a processing error",
						},
					},
				},
			},
		),
	)
})
