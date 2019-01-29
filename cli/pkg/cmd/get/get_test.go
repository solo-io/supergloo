package get_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/cli/test/utils"
)

var _ = Describe("Get", func() {
	It("does not panic", func() {
		err := utils.Supergloo("get routingrule  -o yaml")
		Expect(err).NotTo(HaveOccurred())
	})
})
