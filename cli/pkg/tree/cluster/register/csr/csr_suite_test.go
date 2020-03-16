package csr_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCsr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Csr Suite")
}
