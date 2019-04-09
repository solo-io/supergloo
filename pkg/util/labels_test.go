package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/pkg/util"
)

var _ = Describe("Labels", func() {
	It("produces labels that include the type name and the resource ref", func() {
		labels := LabelsForResource(inputs.IstioMesh("test", nil))
		Expect(labels).To(Equal(map[string]string{
			"*v1.Mesh": "test.fancy-istio",
		}))
	})
})
