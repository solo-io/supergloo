package decider_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFederationStrategy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Federation Decider Suite")
}
