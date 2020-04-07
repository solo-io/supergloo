package resource_printing_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestResourcePrinting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ResourcePrinting Suite")
}
