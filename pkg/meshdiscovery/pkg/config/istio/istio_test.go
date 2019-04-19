package istio

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
)

var _ = Describe("istio discovery config", func() {

	var (
		cs  *clientset.Clientset
		ctx context.Context
	)

	BeforeEach(func() {
		var err error
		ctx = context.TODO()
		cs, err = clientset.ClientsetFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		clients.UseMemoryClients()
	})

	Context("plugin creation", func() {
		It("can be initialized without an error", func() {
			_, err := NewIstioConfigDiscoveryRunner(ctx, cs)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("full mesh", func() {

		var (
			mesh       *v1.Mesh
			install    *v1.Install
			meshPolicy *v1alpha1.MeshPolicy
		)
		BeforeEach(func() {
			mesh = &v1.Mesh{
				MeshType: &v1.Mesh_Istio{
					Istio: &v1.IstioMesh{
						InstallationNamespace: "hello",
					},
				},
				MtlsConfig: &v1.MtlsConfig{},
				DiscoveryMetadata: &v1.DiscoveryMetadata{
					InstallationNamespace: "hello",
				},
			}
			meshPolicy = &v1alpha1.MeshPolicy{
				Peers: []*v1alpha1.PeerAuthenticationMethod{
					{
						Params: &v1alpha1.PeerAuthenticationMethod_Mtls{
							Mtls: &v1alpha1.MutualTls{
								Mode: v1alpha1.MutualTls_STRICT,
							},
						},
					},
				},
			}
			install = &v1.Install{
				InstallationNamespace: "world",
				InstallType: &v1.Install_Mesh{
					Mesh: &v1.MeshInstall{
						MeshInstallType: &v1.MeshInstall_IstioMesh{
							IstioMesh: &v1.IstioInstall{
								EnableMtls:   true,
								IstioVersion: "1.0.9",
								CustomRootCert: &core.ResourceRef{
									Name: "one",
								},
								EnableAutoInject: true,
							},
						},
					},
				},
			}
		})

		It("Can merge properly with no install or mesh policy", func() {
			fm := &meshResources{
				Mesh: mesh,
			}
			Expect(fm.merge()).To(BeEquivalentTo(fm.Mesh))
		})
		It("can merge properly with a mesh policy", func() {
			fm := &meshResources{
				Mesh:       mesh,
				MeshPolicy: meshPolicy,
			}
			Expect(fm.merge().DiscoveryMetadata.MtlsConfig).To(BeEquivalentTo(&v1.MtlsConfig{
				RootCertificate: nil,
				MtlsEnabled:     true,
			}))
		})
		It("can merge properly with install and mesh policy", func() {
			fm := &meshResources{
				Mesh:       mesh,
				Install:    install,
				MeshPolicy: meshPolicy,
			}
			merge := fm.merge()
			Expect(merge.DiscoveryMetadata.MtlsConfig).To(BeEquivalentTo(&v1.MtlsConfig{
				MtlsEnabled: true,
				RootCertificate: &core.ResourceRef{
					Name: "one",
				},
			}))
			Expect(merge.DiscoveryMetadata).To(BeEquivalentTo(&v1.DiscoveryMetadata{
				MtlsConfig: &v1.MtlsConfig{
					MtlsEnabled: true,
					RootCertificate: &core.ResourceRef{
						Name: "one",
					},
				},
				InstallationNamespace:  "world",
				MeshVersion:            "1.0.9",
				InjectedNamespaceLabel: injectionLabel,
				EnableAutoInject:       true,
			}))
		})

	})
})
