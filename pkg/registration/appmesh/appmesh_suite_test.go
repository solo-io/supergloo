package appmesh

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestAppmesh(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "AWS App Mesh Registration Suite")
}

const testNamespace = "supergloo-system"

var podNamespace string

// The test relies on the POD_NAMESPACE variable to be set.
// To be safe store the current value and set it back after the test.
var _ = BeforeSuite(func() {
	podNamespace = os.Getenv("POD_NAMESPACE")
	Expect(os.Setenv("POD_NAMESPACE", testNamespace)).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(os.Setenv("POD_NAMESPACE", podNamespace)).NotTo(HaveOccurred())
})
