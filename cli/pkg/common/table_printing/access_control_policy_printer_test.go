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
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
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
		accessControlPolicies []*zephyr_networking.AccessControlPolicy,
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
			[]*zephyr_networking.AccessControlPolicy{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: zephyr_networking_types.AccessControlPolicySpec{
						AllowedPorts: []uint32{8080, 8443},
						AllowedMethods: []zephyr_core_types.HttpMethodValue{
							zephyr_core_types.HttpMethodValue_GET,
							zephyr_core_types.HttpMethodValue_POST,
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "simple",
					},
					Spec: zephyr_networking_types.AccessControlPolicySpec{
						SourceSelector: &zephyr_core_types.IdentitySelector{
							IdentitySelectorType: &zephyr_core_types.IdentitySelector_Matcher_{
								Matcher: &zephyr_core_types.IdentitySelector_Matcher{
									Namespaces: []string{"ns1", "ns2"},
								},
							},
						},
					},
				},
			}),
	)
})
