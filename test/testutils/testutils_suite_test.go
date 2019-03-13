package testutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testutils Suite")
}
