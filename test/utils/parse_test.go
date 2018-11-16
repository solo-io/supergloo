package utils_test

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/supergloo/test/utils"
)

var _ = Describe("Parse", func() {
	It("works", func(){
		out, err := ParseKubeManifest(context.TODO(), Linkerd1Yaml)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).NotTo(HaveOccurred())
	})
})
