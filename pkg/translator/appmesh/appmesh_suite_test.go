package istio_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAppmesh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Appmesh Suite")
}
