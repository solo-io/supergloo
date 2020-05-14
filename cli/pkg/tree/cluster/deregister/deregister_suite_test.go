package deregister_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDeregister(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deregister Suite")
}
