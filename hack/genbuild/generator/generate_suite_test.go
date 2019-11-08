package generator

import (
	"testing"

	"github.com/solo-io/mesh-projects/hack/genbuild/gencheck"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGenerate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Build File Generation Suite")
}

var fileGenerationObjective *gencheck.FileGenerationObjective

var _ = BeforeSuite(func() {
	var err error
	fileGenerationObjective, err = gencheck.NewFileGenerationObjective("reference")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	// Note: this fails when run in parallel because a new fileGenerationObjective is established for each node and so
	// the cumulative action of each It block is spread over multiple FileGenerationObjective objects and if more than
	// one node is used the "AllFilesVisited" statement is never true.
	// Keeping this as an optional check when adding new tests to verify that all reference files have been exercised.
	//Expect(fileGenerationObjective.AllFilesVisited()).To(BeTrue())
})
