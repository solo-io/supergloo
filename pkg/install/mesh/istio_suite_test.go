package mesh

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestIstio(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Istio Suite")
}
