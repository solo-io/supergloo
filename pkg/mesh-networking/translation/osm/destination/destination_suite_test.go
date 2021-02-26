package destination_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDestination(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Destination Suite")
}
