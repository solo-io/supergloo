package appmesh_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

func TestAppmesh(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "Appmesh Suite")
}
