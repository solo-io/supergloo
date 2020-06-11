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
var UPDATE_ACP_GOLDENS = false

var _ = Describe("Access Control Printer", func() {
	const acpGoldenDirectory = "access_control_policy"
	var runTest = func(
		fileName string,
		printMode table_printing.PrintMode,
		accessControlPolicies []*smh_networking.AccessControlPolicy,
	) {
		goldenFilename := test_goldens.GoldenFilePath(acpGoldenDirectory, fileName)
		goldenContents, err := ioutil.ReadFile(goldenFilename)
		Expect(err).NotTo(HaveOccurred())

		output := &bytes.Buffer{}
		err = table_printing.NewAccessControlPolicyPrinter(table_printing.DefaultTableBuilder()).
			Print(output, printMode, accessControlPolicies)

		if UPDATE_ACP_GOLDENS || test_goldens.UpdateGoldens() {
			err = ioutil.WriteFile(goldenFilename, []byte(output.String()), os.ModeAppend)
			Expect(err).NotTo(HaveOccurred(), "Failed to update the golden file")
			Fail("Need to change UPDATE_GOLDENS back to false before committing")
		} else {
			Expect(err).NotTo(HaveOccurred())
			Expect(output.String()).To(Equal(string(goldenContents)))
		}
	}

	DescribeTable("ACP table printer", runTest,
		Entry(
			"can print multiple complex AC policies",
			"multiple_complex_access_control_policies",
			table_printing.ServicePrintMode,
			[]*smh_networking.AccessControlPolicy{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: smh_networking_types.AccessControlPolicySpec{
						AllowedPorts: []uint32{8080, 8443},
						AllowedMethods: []smh_core_types.HttpMethodValue{
							smh_core_types.HttpMethodValue_GET,
							smh_core_types.HttpMethodValue_POST,
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: smh_networking_types.AccessControlPolicySpec{
						SourceSelector: &smh_core_types.IdentitySelector{
							IdentitySelectorType: &smh_core_types.IdentitySelector_Matcher_{
								Matcher: &smh_core_types.IdentitySelector_Matcher{
									Namespaces: []string{"ns1", "ns2"},
								},
							},
						},
					},
				},
			}),
	)
})
