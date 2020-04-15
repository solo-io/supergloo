package resource_printing_test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/table_printing/test_goldens"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
)

var _ = Describe("JSONPrinter", func() {
	var (
		resourcePrinter resource_printing.ResourcePrinter
	)

	BeforeEach(func() {
		resourcePrinter = resource_printing.NewResourcePrinter()
	})

	var runTest = func(printFormat resource_printing.OutputFormat, fileName string, obj runtime.Object) {
		goldenFilename := test_goldens.GoldenFilePath("", fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = resourcePrinter.Print(output, obj, printFormat)
		Expect(err).ToNot(HaveOccurred())

		if test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).ToNot(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	vm := &zephyr_networking.VirtualMesh{
		TypeMeta: metav1.TypeMeta{Kind: "VirtualMesh"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm",
			Namespace: "service-mesh-hub",
		},
		Spec: networking_types.VirtualMeshSpec{
			DisplayName: "test-vm",
			Meshes: []*types.ResourceRef{
				{
					Name:      "istio-istio-system-management-plane-cluster",
					Namespace: "service-mesh-hub",
				},
				{
					Name:      "istio-istio-system-target-cluster",
					Namespace: "service-mesh-hub",
				},
			},
			CertificateAuthority: &networking_types.VirtualMeshSpec_CertificateAuthority{
				Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
					Builtin: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
						TtlDays:         365,
						RsaKeySizeBytes: 4096,
						OrgName:         "solo.io",
					},
				},
			},
			Federation: &networking_types.VirtualMeshSpec_Federation{
				Mode: networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
			},
			TrustModel: &networking_types.VirtualMeshSpec_Shared{
				Shared: &networking_types.VirtualMeshSpec_SharedTrust{},
			},
		},
	}

	DescribeTable("Resource Printer", runTest,
		Entry("should print VirtualMesh as json", resource_printing.JSONFormat, "virtualmesh_json", vm),
		Entry("should print VirtualMesh as yaml", resource_printing.YAMLFormat, "virtualmesh_yaml", vm),
	)

})
