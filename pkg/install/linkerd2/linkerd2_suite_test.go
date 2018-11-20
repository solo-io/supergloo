package linkerd2

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConsulInstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linkerd2 Installer Suite")
}
