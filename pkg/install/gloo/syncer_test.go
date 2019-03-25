package gloo_test

import (
	"context"

	"github.com/solo-io/supergloo/pkg/install/gloo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"
)

type mockGlooInstaller struct {
	enabledInstalls, disabledInstalls v1.InstallList
	errorOnInstall                    bool
}

func (i *mockGlooInstaller) EnsureGlooInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList, meshIngresses v1.MeshIngressList) (*v1.MeshIngress, error) {
	if i.errorOnInstall {
		return nil, errors.Errorf("i was told to error")
	}
	if install.Disabled {
		i.disabledInstalls = append(i.disabledInstalls, install)
		return nil, nil
	}
	i.enabledInstalls = append(i.enabledInstalls, install)
	return &v1.MeshIngress{Metadata: install.Metadata}, nil
}

type failingMeshIngressClient struct {
	errorOnWrite, errorOnRead, errorOnDelete bool
}

func (c *failingMeshIngressClient) BaseClient() clients.ResourceClient {
	panic("implement me")
}

func (c *failingMeshIngressClient) Register() error {
	panic("implement me")
}

func (c *failingMeshIngressClient) Read(namespace, name string, opts clients.ReadOpts) (*v1.MeshIngress, error) {
	if c.errorOnRead {
		return nil, errors.Errorf("i was told to error")
	}
	return &v1.MeshIngress{Metadata: core.Metadata{Name: name, Namespace: namespace}}, nil
}

func (c *failingMeshIngressClient) Write(resource *v1.MeshIngress, opts clients.WriteOpts) (*v1.MeshIngress, error) {
	if c.errorOnWrite {
		return nil, errors.Errorf("i was told to error")
	}
	return resource, nil
}

func (c *failingMeshIngressClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	if c.errorOnDelete {
		return errors.Errorf("i was told to error")
	}
	return nil
}

func (c *failingMeshIngressClient) List(namespace string, opts clients.ListOpts) (v1.MeshIngressList, error) {
	panic("implement me")
}

func (c *failingMeshIngressClient) Watch(namespace string, opts clients.WatchOpts) (<-chan v1.MeshIngressList, <-chan error, error) {
	panic("implement me")
}

var _ = Describe("Syncer", func() {
	var (
		installer     *mockGlooInstaller
		meshClient    v1.MeshClient
		ingressClient v1.MeshIngressClient
		installClient v1.InstallClient
		report        reporter.Reporter
	)
	Context("happy paths", func() {

		BeforeEach(func() {
			installer = &mockGlooInstaller{}
			meshClient, _ = v1.NewMeshClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			ingressClient, _ = v1.NewMeshIngressClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			installClient, _ = v1.NewInstallClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			report = reporter.NewReporter("test", installClient.BaseClient())
		})
		Context("multiple active installs", func() {
			It("it reports an error on them, does not call the installer", func() {
				installList := v1.InstallList{
					inputs.GlooIstall("a", "b", "c", "versiondoesntmatter", false),
					inputs.GlooIstall("b", "b", "c", "versiondoesntmatter", false),
				}
				snap := &v1.InstallSnapshot{Installs: map[string]v1.InstallList{"": installList}}
				installeSyncer := gloo.NewInstallSyncer(installer, meshClient, ingressClient, report)
				err := installeSyncer.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())

				Expect(installer.disabledInstalls).To(HaveLen(0))
				Expect(installer.enabledInstalls).To(HaveLen(0))

				i1, err := installClient.Read("b", "a", clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(i1.Status.State).To(Equal(core.Status_Rejected))
				Expect(i1.Status.Reason).To(ContainSubstring("multiple gloo ingress installations are not currently supported"))

				i2, err := installClient.Read("b", "b", clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(i2.Status.State).To(Equal(core.Status_Rejected))
				Expect(i2.Status.Reason).To(ContainSubstring("multiple gloo ingress installations are not currently supported"))
			})
		})
		Context("one active install, one inactive install with a previous install", func() {
			It("it reports success, calls installer, writes the created mesh", func() {
				installList := v1.InstallList{
					inputs.GlooIstall("a", "b", "c", "versiondoesntmatter", true),
					inputs.GlooIstall("b", "b", "c", "versiondoesntmatter", false),
				}
				installedIngress, _ := ingressClient.Write(&v1.MeshIngress{
					Metadata: core.Metadata{Namespace: "a", Name: "a"},
				}, clients.WriteOpts{})
				ref := installedIngress.Metadata.Ref()
				install := installList[0]
				Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Ingress{}))
				mesh := install.InstallType.(*v1.Install_Ingress)
				mesh.Ingress.InstalledIngress = &ref
				snap := &v1.InstallSnapshot{Installs: map[string]v1.InstallList{"": installList}}
				installSyncer := gloo.NewInstallSyncer(installer, meshClient, ingressClient, report)
				err := installSyncer.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())

				Expect(installer.disabledInstalls).To(HaveLen(1))
				Expect(installer.enabledInstalls).To(HaveLen(1))

				i1, err := installClient.Read("b", "a", clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(i1.Status.State).To(Equal(core.Status_Accepted))

				i2, err := installClient.Read("b", "b", clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(i2.Status.State).To(Equal(core.Status_Accepted))

				// installed mesh should have been removed
				_, err = meshClient.Read(installedIngress.Metadata.Namespace, installedIngress.Metadata.Name, clients.ReadOpts{})
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotExist(err)).To(BeTrue())
			})
		})
	})
	Context("when install fails", func() {
		BeforeEach(func() {
			installer = &mockGlooInstaller{errorOnInstall: true}
			meshClient, _ = v1.NewMeshClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			installClient, _ = v1.NewInstallClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			report = reporter.NewReporter("test", installClient.BaseClient())
		})
		It("it marks the install as rejected", func() {
			installList := v1.InstallList{
				inputs.GlooIstall("a", "b", "c", "versiondoesntmatter", true),
				inputs.GlooIstall("b", "b", "c", "versiondoesntmatter", false),
			}

			snap := &v1.InstallSnapshot{Installs: map[string]v1.InstallList{"": installList}}
			installeSyncer := gloo.NewInstallSyncer(installer, meshClient, ingressClient, report)
			err := installeSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			i1, err := installClient.Read("b", "a", clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(i1.Status.State).To(Equal(core.Status_Accepted))

			i2, err := installClient.Read("b", "b", clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(i2.Status.State).To(Equal(core.Status_Rejected))
			Expect(i2.Status.Reason).To(ContainSubstring("install failed"))
		})
	})
	Context("when ingress client fails", func() {
		BeforeEach(func() {
			installer = &mockGlooInstaller{}
			meshClient, _ = v1.NewMeshClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			ingressClient = &failingMeshIngressClient{errorOnWrite: true}
			installClient, _ = v1.NewInstallClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			report = reporter.NewReporter("test", installClient.BaseClient())
		})
		It("it marks the install as rejected", func() {
			installList := v1.InstallList{
				inputs.GlooIstall("a", "b", "c", "versiondoesntmatter", true),
				inputs.GlooIstall("b", "b", "c", "versiondoesntmatter", false),
			}

			snap := &v1.InstallSnapshot{Installs: map[string]v1.InstallList{"": installList}}
			installeSyncer := gloo.NewInstallSyncer(installer, meshClient, ingressClient, report)
			err := installeSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			i1, err := installClient.Read("b", "a", clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(i1.Status.State).To(Equal(core.Status_Accepted))

			i2, err := installClient.Read("b", "b", clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(i2.Status.State).To(Equal(core.Status_Rejected))
			Expect(i2.Status.Reason).To(ContainSubstring("writing installed mesh-ingress object {b b} failed after successful install"))
		})
	})
})
