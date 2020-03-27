package istio_translator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIstioTranslator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IstioTranslator Suite")
}
