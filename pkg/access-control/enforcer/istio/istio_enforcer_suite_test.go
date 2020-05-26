package istio_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIstioEnforcer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IstioEnforcer Suite")
}
