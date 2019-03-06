package surveyutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/cliutil/testutil"
	. "github.com/solo-io/supergloo/cli/pkg/surveyutils"
)

var _ = Describe("MapStringString", func() {
	// TODO: unexclude this test when c.ExpectEOF() is fixed
	// relevant issue: https://github.com/solo-io/gloo/issues/387
	XIt("creates the expected map[string]string", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish: ")
			c.SendLine("foo=bar")
			c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish: ")
			c.SendLine("baz=qux")
			c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish: ")
			c.SendLine("")
			c.ExpectEOF()
		}, func() {
			in := make(map[string]string)
			err := SurveyMapStringString(&in)
			Expect(err).NotTo(HaveOccurred())
			Expect(in).To(Equal(map[string]string{"foo": "bar", "baz": "qux"}))
		})
	})

	It("errors on invalid input", func() {
		testutil.ExpectInteractive(func(c *testutil.Console) {
			c.ExpectString("enter a key-value pair in the format KEY=VAL. leave empty to finish: ")
			c.SendLine("foobar")
			c.ExpectString("")
			c.ExpectEOF()
		}, func() {
			in := make(map[string]string)
			err := SurveyMapStringString(&in)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("key-value pair must be in the format KEY=VAL"))
		})
	})
})
