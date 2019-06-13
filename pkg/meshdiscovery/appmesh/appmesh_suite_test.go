package appmesh_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	T *testing.T
)

func TestAppmesh(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Appmesh Suite")
}
