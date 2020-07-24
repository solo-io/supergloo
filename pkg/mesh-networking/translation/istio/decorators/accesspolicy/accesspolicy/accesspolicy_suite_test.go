package accesspolicy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAccesspolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Accesspolicy Suite")
}
