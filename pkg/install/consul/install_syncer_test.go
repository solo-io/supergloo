package consul_test

import (
	"context"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/consul"
)

var _ = Describe("ConsulInstallSyncer", func() {
	It("Can install consul with mtls enabled", func() {
		syncer := consul.ConsulInstallSyncer{}
		var ctx context.Context
		snap := v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				"not_used": v1.InstallList{
					&v1.Install{
						Consul: &v1.ConsulInstall{
							Path: "/Users/rick/.helm/cache/archive/v0.3.0.tar.gz",
						},
						Encryption: &v1.Encryption{
							TlsEnabled: true,
						},
					},
				},
			},
		}
		err := syncer.Sync(ctx, &snap)
		Expect(err).NotTo(HaveOccurred())

		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())
		_, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
	})

	FIt("Can install consul without mtls enabled", func() {
		syncer := consul.ConsulInstallSyncer{}
		var ctx context.Context
		snap := v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				"not_used": v1.InstallList{
					&v1.Install{
						Consul: &v1.ConsulInstall{
							Path: "/Users/rick/.helm/cache/archive/v0.3.0.tar.gz",
						},
						Encryption: &v1.Encryption{
							TlsEnabled: false,
						},
					},
				},
			},
		}
		err := syncer.Sync(ctx, &snap)
		Expect(err).NotTo(HaveOccurred())

		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())
		_, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
	})
})
