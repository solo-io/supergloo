package registration_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRegistration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Registration Suite")
}
