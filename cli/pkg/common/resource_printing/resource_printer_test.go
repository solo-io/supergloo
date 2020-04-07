package resource_printing_test

import (
	"bytes"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/solo-io/mesh-projects/cli/pkg/common/resource_printing"
)

var _ = Describe("JSONPrinter", func() {
	var (
		resourcePrinter resource_printing.ResourcePrinter
	)

	BeforeEach(func() {
		resourcePrinter = resource_printing.NewResourcePrinter()
	})

	var runTest = func(printFormat string, fileName string, obj runtime.Object) {
		goldenContents, err := ioutil.ReadFile("./test_goldens/" + fileName)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = resourcePrinter.Print(output, obj, printFormat)
		Expect(err).ToNot(HaveOccurred())
		s := output.String()
		Expect(s).To(Equal(string(goldenContents)))
	}

	vm := &networking_v1alpha1.VirtualMesh{
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

	table.DescribeTable("should print in specified format",
		runTest,
		table.Entry("should print VirtualMesh as json",
			resource_printing.JSONFormat, "virtualmesh_json.txt", vm,
		),
		table.Entry("should print VirtualMesh as yaml",
			resource_printing.YAMLFormat, "virtualmesh_yaml.txt", vm,
		),
	)
})
