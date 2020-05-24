package errutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestErrutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errutils Suite")
}
