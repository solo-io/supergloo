package meshservice_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMeshservice(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Meshservice Suite")
}
