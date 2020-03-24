package crd_uninstall_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCrd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Crd Uninstaller Suite")
}
