package setup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/supergloo/pkg/setup"
)

var _ = Describe("Setup", func() {
	It("works", func() {
		err := Main()
		Expect(err).NotTo(HaveOccurred())
	})
})
