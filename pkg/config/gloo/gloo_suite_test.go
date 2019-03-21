package gloo_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGloo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gloo Suite")
}
