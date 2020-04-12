package prompts_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPrompts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prompts Suite")
}
