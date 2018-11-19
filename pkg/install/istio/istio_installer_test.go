package istio_test

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/solo-io/supergloo/test/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

/*
End to end tests for istio installs with and without mTLS enabled.
Tests assume you already have a Kubernetes environment with Helm / Tiller set up, and with a "supergloo-system" namespace.
The tests will install Istio and get it configured and validate all services up and running, then tear down and
clean up all resources created. This will take about 45 seconds with mTLS, and 20 seconds without.
*/
var _ = Describe("Istio Installer", func() {

	installNamespace := "istio-system"
	superglooNamespace := "supergloo-system" // this needs to be made before running tests
	meshName := "istio"

	getSnapshot := func(mtls bool) *v1.InstallSnapshot {
		return &v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				superglooNamespace: v1.InstallList{
					&v1.Install{
						Metadata: core.Metadata{
							Namespace: superglooNamespace,
							Name:      meshName,
						},
						InstallNamespace: installNamespace,
						MeshType:         v1.MeshType_ISTIO,
						ChartLocator: &v1.HelmChartLocator{
							Kind: &v1.HelmChartLocator_ChartPath{
								ChartPath: &v1.HelmChartPath{
									// Use a local path for testing to save a lot of time...
									Path: "https://storage.googleapis.com/istio-prerelease/daily-build/master-latest-daily/charts/istio-1.1.0.tgz",
								},
							},
						},
						Encryption: &v1.Encryption{
							TlsEnabled: mtls,
						},
					},
				},
			},
		}
	}

	kubeCache := kube.NewKubeCache()

	var meshClient v1.MeshClient
	var syncer install.InstallSyncer

	BeforeEach(func() {
		// This shouldn't be necessary, but helm will fail to install if there are CRDs already defined
		// Rather than fail later, let's just try deleting them before the test
		util.TryDeleteIstioCrds()
		meshClient = util.GetMeshClient(kubeCache)
		syncer = install.InstallSyncer{
			Kube:           util.GetKubeClient(),
			MeshClient:     meshClient,
			SecurityClient: util.GetSecurityClient(),
		}
	})

	AfterEach(func() {
		util.TryDeleteIstioCrds()
		util.UninstallHelmRelease(meshName)
		util.DeleteCrb(istio.CrbName)
		util.TerminateNamespaceBlocking(installNamespace)
		meshClient.Delete(superglooNamespace, meshName, clients.DeleteOpts{})
	})

	It("Can install istio with mtls enabled", func() {
		snap := getSnapshot(true)
		err := syncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		util.WaitForAvailablePods(installNamespace)
	})

	It("Can install istio without mtls enabled", func() {
		snap := getSnapshot(false)
		err := syncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		util.WaitForAvailablePods(installNamespace)
	})
})
