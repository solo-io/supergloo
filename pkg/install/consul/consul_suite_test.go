package consul

import (
	"testing"

	"github.com/solo-io/supergloo/test/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConsulInstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consul Installer Suite")
}

var _ = BeforeSuite(func() {
	util.TryCreateNamespace("supergloo-system")
})
