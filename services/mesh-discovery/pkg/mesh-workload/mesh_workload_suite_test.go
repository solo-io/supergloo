package mesh_workload_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	T *testing.T
)

func TestMeshWorkload(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "MeshWorkload Suite")
}
