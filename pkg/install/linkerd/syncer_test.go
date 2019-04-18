package linkerd_test

import (
	"context"

	"k8s.io/client-go/kubernetes/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/kubeinstall/mocks"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/util"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/pkg/install/linkerd"
)

var _ = Describe("Syncer", func() {
	Context("happy paths", func() {
		var (
			kubeInstaller *mocks.MockKubeInstaller
			meshClient    v1.MeshClient
			installClient v1.InstallClient
			report        reporter.Reporter
		)
		BeforeEach(func() {
			kubeInstaller = &mocks.MockKubeInstaller{}
			meshClient, _ = v1.NewMeshClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			installClient, _ = v1.NewInstallClient(&factory.MemoryResourceClientFactory{
				Cache: memory.NewInMemoryResourceCache(),
			})
			report = reporter.NewReporter("test", installClient.BaseClient())
		})
		Context("one active install, one inactive install with a previous install", func() {
			It("it reports success, calls installer, writes the created mesh", func() {
				installList := v1.InstallList{
					inputs.LinkerdInstall("a", "b", "c", "versiondoesntmatter", true),
					inputs.LinkerdInstall("b", "b", "c", Version_stable230, false),
				}
				install := installList[0]
				Expect(install.InstallType).To(BeAssignableToTypeOf(&v1.Install_Mesh{}))
				snap := &v1.InstallSnapshot{Installs: map[string]v1.InstallList{"": installList}}
				installSyncer := NewInstallSyncer(kubeInstaller, fake.NewSimpleClientset(), meshClient, report)
				err := installSyncer.Sync(context.TODO(), snap)
				Expect(err).NotTo(HaveOccurred())

				Expect(kubeInstaller.PurgeCalledWith.InstallLabels).To(Equal(util.LabelsForResource(installList[0])))
				Expect(kubeInstaller.ReconcileCalledWith.InstallLabels).To(Equal(util.LabelsForResource(installList[1])))
				Expect(kubeInstaller.ReconcileCalledWith.Resources).To(HaveLen(30))
				Expect(kubeInstaller.ReconcileCalledWith.InstallNamespace).To(Equal(installList[1].InstallationNamespace))

				i1, err := installClient.Read("b", "a", clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(i1.Status.State).To(Equal(core.Status_Accepted))

				i2, err := installClient.Read("b", "b", clients.ReadOpts{})
				Expect(err).NotTo(HaveOccurred())
				Expect(i2.Status.State).To(Equal(core.Status_Accepted))
			})
		})
	})
})
