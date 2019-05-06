package gloo

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/registration/gloo/mocks"
)

var _ = Describe("gloo registration syncers", func() {

	var (
		ctrl   *gomock.Controller
		plugin *mocks.MockGlooIngressPlugin
		cs     *clientset.Clientset
		syncer v1.RegistrationSyncer
		ctx    context.Context

		istioMeshMetadata = core.Metadata{
			Name:      "istio",
			Namespace: "istio-system",
		}
		istioMesh = &v1.Mesh{
			MtlsConfig: &v1.MtlsConfig{
				MtlsEnabled: true,
			},
			MeshType: &v1.Mesh_Istio{
				Istio: &v1.IstioMesh{
					InstallationNamespace: "istio-system",
				},
			},
			Metadata: istioMeshMetadata,
		}

		linkerdMeshMetadata = core.Metadata{
			Name:      "linkerd",
			Namespace: "linkerd-system",
		}
		linkerdMesh = &v1.Mesh{
			MtlsConfig: &v1.MtlsConfig{
				MtlsEnabled: true,
			},
			MeshType: &v1.Mesh_Linkerd{
				Linkerd: &v1.LinkerdMesh{
					InstallationNamespace: "linkerd-system",
				},
			},
			Metadata: linkerdMeshMetadata,
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

		newMockClientset = func() *clientset.Clientset {
			return &clientset.Clientset{
				Supergloo: &clientset.SuperglooClients{
					Mesh:        clients.MustMeshClient(),
					MeshIngress: clients.MustMeshIngressClient(),
				},
			}
		}

		newSnapshot = func(meshes v1.MeshList, ingresses v1.MeshIngressList) *v1.RegistrationSnapshot {
			return &v1.RegistrationSnapshot{
				Meshes: v1.MeshesByNamespace{
					"": meshes,
				},
				Meshingresses: v1.MeshingressesByNamespace{
					"": ingresses,
				},
			}
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		plugin = mocks.NewMockGlooIngressPlugin(ctrl)
		cs = newMockClientset()
		syncer = NewGlooRegistrationSyncer(cs, plugin)
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("errors", func() {
		It("returns an error when a target mesh cannot be found", func() {
			ref := istioMeshMetadata.Ref()
			snap := newSnapshot(v1.MeshList{linkerdMesh}, v1.MeshIngressList{glooIngress(&ref)})
			plugin.EXPECT().HandleMeshes(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			err := syncer.Sync(ctx, snap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find mesh"))
		})

		It("returns an error when a plugin returns an error", func() {
			ref := istioMeshMetadata.Ref()
			ingress := glooIngress(&ref)
			snap := newSnapshot(v1.MeshList{istioMesh, linkerdMesh}, v1.MeshIngressList{ingress})
			plugin.EXPECT().HandleMeshes(gomock.Any(), ingress, gomock.Eq(v1.MeshList{istioMesh})).Times(1).
				Return(fmt.Errorf("test error"))
			err := syncer.Sync(ctx, snap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))

		})
	})

	Context("success", func() {
		It("gets called with all meshes that are pointed to by the gloo mesh ingress", func() {
			ref := istioMeshMetadata.Ref()
			ingress := glooIngress(&ref)
			snap := newSnapshot(v1.MeshList{istioMesh, linkerdMesh}, v1.MeshIngressList{ingress})
			plugin.EXPECT().HandleMeshes(gomock.Any(), ingress, gomock.Eq(v1.MeshList{istioMesh})).Times(1)
			err := syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
		})

		It("only gets called if the mesh ingress is a gloo type", func() {
			snap := newSnapshot(v1.MeshList{istioMesh, linkerdMesh}, v1.MeshIngressList{&v1.MeshIngress{}})
			plugin.EXPECT().HandleMeshes(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			err := syncer.Sync(ctx, snap)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
