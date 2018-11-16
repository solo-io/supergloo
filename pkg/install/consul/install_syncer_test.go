package consul_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/install/consul"
)

var _ = Describe("ConsulInstallSyncer", func() {
	FIt("Can get helm client", func() {
		_, err := consul.GetHelmClient()
		Expect(err).NotTo(HaveOccurred())
	})
})
