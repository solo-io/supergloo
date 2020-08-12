package cors_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cors Suite")
}
