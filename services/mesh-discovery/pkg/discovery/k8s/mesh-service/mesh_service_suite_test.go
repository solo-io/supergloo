package mesh_service_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMeshService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MeshService Suite")
}
