package preprocess_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPreprocess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Preprocess Suite")
}
