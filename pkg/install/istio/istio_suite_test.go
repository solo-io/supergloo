package istio

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/supergloo/test/util"
)

func TestIstioInstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Istio Installer Suite")
}

var _ = BeforeSuite(func() {
	util.TryCreateNamespace("supergloo-system")
})
