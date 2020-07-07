package fieldutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFieldutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fieldutils Suite")
}
