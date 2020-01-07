package common_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommonUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Utils Discovery Suite")
}
