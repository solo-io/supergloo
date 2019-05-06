package linkerd

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	clients2 "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("linkerd mtls plugin", func() {
	var (
		// kube   kubernetes.Interface
		plugin      *glooLinkerdMtlsPlugin
		ctx         context.Context
		cs          *clientset.Clientset
		linkerdMesh = &v1.Mesh{
			MtlsConfig: &v1.MtlsConfig{
				MtlsEnabled: true,
			},
			MeshType: &v1.Mesh_Linkerd{
				Linkerd: &v1.LinkerdMesh{
					InstallationNamespace: "linkerd-system",
				},
			},
			Metadata: core.Metadata{
				Name:      "linkerd",
				Namespace: "linkerd-system",
			},
		}
		glooIngress = func(meshes ...*core.ResourceRef) *v1.MeshIngress {
			return &v1.MeshIngress{
				InstallationNamespace: "gloo-system",
				MeshIngressType: &v1.MeshIngress_Gloo{
					Gloo: &v1.GlooMeshIngress{},
				},
				Meshes: meshes,
			}
		}
		settings = &gloov1.Settings{
			Metadata: core.Metadata{
				Name:      "settings",
				Namespace: "gloo-system",
			},
			Linkerd: false,
		}
	)

	var mockClientset = func() *clientset.Clientset {
		return &clientset.Clientset{
			Supergloo: &clientset.SuperglooClients{
				Settings: clients.MustSettingsClient(),
			},
		}
	}

	BeforeEach(func() {
		clients.UseMemoryClients()
		// kube = clients.MustKubeClient()
		ctx = context.TODO()
		cs = mockClientset()
		plugin = NewGlooLinkerdMtlsPlugin(cs)
		_, err := cs.Supergloo.Settings.Write(settings, clients2.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := cs.Supergloo.Settings.Delete(settings.Metadata.Namespace, settings.Metadata.Name, clients2.DeleteOpts{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns nil if ingress is undefined", func() {
		err := plugin.HandleMeshes(ctx, nil, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does nothing if no linkerd mesh exists", func() {
		err := plugin.HandleMeshes(ctx, glooIngress(nil), nil)
		Expect(err).NotTo(HaveOccurred())
		result, err := cs.Supergloo.Settings.Read(settings.Metadata.Namespace, settings.Metadata.Name, clients2.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Linkerd).To(BeFalse())
	})

	It("changes linkerd to true if linkerd mesh exists", func() {
		err := plugin.HandleMeshes(ctx, glooIngress(nil), v1.MeshList{linkerdMesh})
		Expect(err).NotTo(HaveOccurred())
		result, err := cs.Supergloo.Settings.Read(settings.Metadata.Namespace, settings.Metadata.Name, clients2.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Linkerd).To(BeTrue())
	})

	It("changes linkerd to false if it was true", func() {
		settings.Linkerd = true
		err := plugin.HandleMeshes(ctx, glooIngress(nil), v1.MeshList{})
		Expect(err).NotTo(HaveOccurred())
		result, err := cs.Supergloo.Settings.Read(settings.Metadata.Namespace, settings.Metadata.Name, clients2.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Linkerd).To(BeFalse())
		settings.Linkerd = false
	})
})
