package istio_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	. "github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/solo-io/supergloo/test/inputs"
)

type mockIstioInstaller struct {
	enabledInstalls, disabledInstalls v1.InstallList
}

func (i *mockIstioInstaller) EnsureIstioInstall(ctx context.Context, install *v1.Install) (*v1.Mesh, error) {
	if install.Disabled {
		i.disabledInstalls = append(i.disabledInstalls, install)
		return nil, nil
	}
	i.enabledInstalls = append(i.enabledInstalls, install)
	return &v1.Mesh{}, nil
}

var _ = Describe("Syncer", func() {
	var (
		installer     *mockIstioInstaller
		meshClient    v1.MeshClient
		installClient v1.InstallClient
		report        reporter.Reporter
	)
	BeforeEach(func() {
		installer = &mockIstioInstaller{}
		meshClient, _ = v1.NewMeshClient(&factory.MemoryResourceClientFactory{
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
				inputs.IstioInstall("a", "b", "c", "versiondoesntmatter", false),
				inputs.IstioInstall("b", "b", "c", "versiondoesntmatter", false),
			}
			snap := &v1.InstallSnapshot{Installs: map[string]v1.InstallList{"": installList}}
			installeSyncer := NewInstallSyncer(installer, meshClient, report)
			err := installeSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			Expect(installer.disabledInstalls).To(HaveLen(0))
			Expect(installer.enabledInstalls).To(HaveLen(0))

			i1, err := installClient.Read("b", "a", clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(i1.Status.State).To(Equal(core.Status_Rejected))
			Expect(i1.Status.Reason).To(ContainSubstring("multiple active istio installactions are not currently supported"))

			i2, err := installClient.Read("b", "b", clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(i2.Status.State).To(Equal(core.Status_Rejected))
			Expect(i2.Status.Reason).To(ContainSubstring("multiple active istio installactions are not currently supported"))
		})
	})
	Context("one active install, one inactive install with a previous install", func() {
		It("it reports success, calls installer, writes the created mesh", func() {
			installList := v1.InstallList{
				inputs.IstioInstall("a", "b", "c", "versiondoesntmatter", true),
				inputs.IstioInstall("b", "b", "c", "versiondoesntmatter", false),
			}
			installedMesh, _ := meshClient.Write(&v1.Mesh{
				Metadata: core.Metadata{Namespace: "a", Name: "a"},
			}, clients.WriteOpts{})
			ref := installedMesh.Metadata.Ref()
			installList[0].InstalledMesh = &ref
			snap := &v1.InstallSnapshot{Installs: map[string]v1.InstallList{"": installList}}
			installeSyncer := NewInstallSyncer(installer, meshClient, report)
			err := installeSyncer.Sync(context.TODO(), snap)
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
			_, err = meshClient.Read(installedMesh.Metadata.Namespace, installedMesh.Metadata.Name, clients.ReadOpts{})
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotExist(err)).To(BeTrue())
		})
	})
})
