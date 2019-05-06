package gloo

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	T *testing.T
)

func TestGloo(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gloo Suite")
}
