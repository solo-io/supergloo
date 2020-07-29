package enforcement_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEnforcement(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Enforcement Suite")
}
