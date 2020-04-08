package install_istio_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIstioInstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mesh Install Suite")
}
