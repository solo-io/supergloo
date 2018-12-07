package cli

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

/*
End to end tests for supergloo-cli
*/

var _ = Describe("cli e2e", func() {

	It("Runs properly", func() {
		Expect("").To(Equal(""))
		fmt.Println("In test")
	})
})
