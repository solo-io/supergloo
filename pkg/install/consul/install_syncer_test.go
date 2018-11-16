package consul_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/consul"
)

var _ = Describe("ConsulInstallSyncer", func() {
	It("Can install consul", func() {
		syncer := consul.ConsulInstallSyncer{}
		var ctx context.Context
		snap := v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				"not_used": v1.InstallList{
					&v1.Install{
						Consul: &v1.ConsulInstall{
							Path: "/Users/rick/.helm/cache/archive/v0.3.0.tar.gz",
						},
					},
				},
			},
		}
		err := syncer.Sync(ctx, &snap)
		Expect(err).NotTo(HaveOccurred())
	})
})
