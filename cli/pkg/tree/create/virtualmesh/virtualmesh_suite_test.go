package virtualmesh_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVirtualmesh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Virtualmesh Suite")
}
