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
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_VMCSR_GOLDENS = false

var _ = Describe("VMCSR Table Printer", func() {
	const tpGoldenDirectory = "vmcsr"

	var runTest = func(fileName string, vmcsrs []*security_v1alpha1.VirtualMeshCertificateSigningRequest) {
		goldenFilename := test_goldens.GoldenFilePath(tpGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewVirtualMeshMCSRPrinter(table_printing.DefaultTableBuilder()).Print(output, vmcsrs)

		if UPDATE_VMCSR_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("VMCSR printer", runTest,
		Entry(
			"can print different kinds of virtual mesh certificate signing requestss",
			"vmcsr",
			[]*security_v1alpha1.VirtualMeshCertificateSigningRequest{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vm-1",
						Namespace: "service-mesh-hub",
					},
					Spec: security_types.VirtualMeshCertificateSigningRequestSpec{
						CsrData: []byte("test-csr"),
						CertConfig: &security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
							Hosts:    []string{"host1", "host2"},
							Org:      "my-org",
							MeshType: core_types.MeshType_ISTIO,
						},
						VirtualMeshRef: &core_types.ResourceRef{
							Name:      "name-1",
							Namespace: "namespace-1",
						},
					},
					Status: security_types.VirtualMeshCertificateSigningRequestStatus{
						ThirdPartyApproval: &security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow{
							ApprovalStatus: security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_APPROVED,
						},
						ComputedStatus: &core_types.Status{
							State: core_types.Status_ACCEPTED,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "vm-2",
						Namespace: "service-mesh-hub",
					},
					Spec: security_types.VirtualMeshCertificateSigningRequestSpec{
						CsrData: nil,
						CertConfig: &security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
							Hosts:    []string{"host2", "host4"},
							Org:      "linkerd-org",
							MeshType: core_types.MeshType_LINKERD,
						},
						VirtualMeshRef: &core_types.ResourceRef{
							Name:      "name-2",
							Namespace: "namespace-2",
						},
					},
					Status: security_types.VirtualMeshCertificateSigningRequestStatus{
						ThirdPartyApproval: &security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow{
							ApprovalStatus: security_types.VirtualMeshCertificateSigningRequestStatus_ThirdPartyApprovalWorkflow_DENIED,
						},
						ComputedStatus: &core_types.Status{
							State:   core_types.Status_CONFLICT,
							Message: "there was a conflict",
						},
						Response: &security_types.VirtualMeshCertificateSigningRequestStatus_Response{},
					},
				},
			},
		),
	)
})
