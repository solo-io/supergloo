package headermanipulation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHeadermanipulation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Headermanipulation Suite")
}
