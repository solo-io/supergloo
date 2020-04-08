package validate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/validate"
)

var _ = Describe("Validate", func() {
	It("should validate k8s DNS1123 name", func() {
		err := validate.K8sName("valid-name")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should invalidate bad k8s DNS1123 name", func() {
		err := validate.K8sName(".invalid-name")
		Expect(err).To(HaveOccurred())
	})

	It("should validate labels from string", func() {
		err := validate.Labels("k1=v1, k2=v2")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should invalidate bad label string", func() {
		err := validate.Labels("k1=v1 k2=v2")
		Expect(err).To(HaveOccurred())
	})

	It("should validate namespaces as comma delimited string", func() {
		err := validate.Namespaces("namespace1, namespace2, namespace3")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should invalidate bad namespaces string", func() {
		err := validate.Namespaces("namespace1 namespace2 namespace3")
		Expect(err).To(HaveOccurred())
	})

	It("should validate positive integer", func() {
		err := validate.PositiveInteger("123")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should invalidate unparsable integer", func() {
		err := validate.PositiveInteger("12a3")
		Expect(err).To(HaveOccurred())
	})

	It("should invalidate non-positive integer", func() {
		err := validate.PositiveInteger("-1")
		Expect(err).To(HaveOccurred())
	})

	It("should allow empty input", func() {
		err := validate.AllowEmpty(func(userInput string) error {
			return nil
		})("")
		Expect(err).ToNot(HaveOccurred())
	})
})
