package install_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIstioInstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istio Install Suite")
}
