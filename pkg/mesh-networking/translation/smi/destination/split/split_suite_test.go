package split_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSplit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Split Suite")
}
