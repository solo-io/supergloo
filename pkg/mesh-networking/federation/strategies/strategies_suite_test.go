package strategies_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStrategies(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Strategies Suite")
}
