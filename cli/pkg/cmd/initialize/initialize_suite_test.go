package initialize_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInitialize(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Initialize Suite")
}
