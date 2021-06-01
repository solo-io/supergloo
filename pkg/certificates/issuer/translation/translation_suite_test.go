package translation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTranslation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Translation Suite")
}
