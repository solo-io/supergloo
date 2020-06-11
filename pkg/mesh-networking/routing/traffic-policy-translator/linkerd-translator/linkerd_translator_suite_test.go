package linkerd_translator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLinkerdTranslator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LinkerdTranslator Suite")
}
