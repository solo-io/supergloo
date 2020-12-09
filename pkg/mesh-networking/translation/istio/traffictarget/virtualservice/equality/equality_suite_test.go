package equality_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEquality(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Equality Suite")
}
