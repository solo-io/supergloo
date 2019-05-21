package smi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSmi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Smi Suite")
}
