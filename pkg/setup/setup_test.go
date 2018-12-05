package setup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/supergloo/pkg/setup"
)

// TODO: ilackarms: make this test runnable (now it blocks forever)
var _ = Describe("Setup", func() {
	XIt("works", func() {
		err := Main(nil)
		Expect(err).NotTo(HaveOccurred())
	})
})
