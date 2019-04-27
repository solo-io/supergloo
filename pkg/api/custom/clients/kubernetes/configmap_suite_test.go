package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNamespace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Namespace Suite")
}
