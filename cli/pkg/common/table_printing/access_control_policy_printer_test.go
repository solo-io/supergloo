package table_printing_test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/cli/pkg/common/table_printing"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if you need to update the golden files programmatically, change this to `true` to write the
// files instead of checking against them
var UPDATE_ACP_GOLDENS = false

var _ = Describe("Access Control Printer", func() {
	var runTest = func(fileName string, printMode table_printing.PrintMode, accessControlPolicies []*v1alpha1.AccessControlPolicy) {
		goldenContents, err := ioutil.ReadFile("./test_goldens/access_control_policy/" + fileName)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewAccessControlPolicyPrinter(table_printing.DefaultTableBuilder).Print(output, printMode, accessControlPolicies)

		if UPDATE_ACP_GOLDENS {
			err = ioutil.WriteFile("./test_goldens/access_control_policy/"+fileName, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_ACP_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("ACP table printer", runTest,
		Entry("can print multiple complex AC policies", "multiple_complex_access_control_policies.txt", table_printing.ServicePrintMode, []*v1alpha1.AccessControlPolicy{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "simple",
				},
				Spec: types.AccessControlPolicySpec{
					AllowedPorts: []uint32{8080, 8443},
					AllowedMethods: []core_types.HttpMethodValue{
						core_types.HttpMethodValue_GET,
						core_types.HttpMethodValue_POST,
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "simple",
				},
				Spec: types.AccessControlPolicySpec{
					SourceSelector: &core_types.IdentitySelector{
						IdentitySelectorType: &core_types.IdentitySelector_Matcher_{
							Matcher: &core_types.IdentitySelector_Matcher{
								Namespaces: []string{"ns1", "ns2"},
							},
						},
					},
				},
			},
		}),
	)
})
