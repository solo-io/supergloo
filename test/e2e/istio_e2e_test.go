package e2e

import (
	"context"
	"os"

	"github.com/solo-io/supergloo/pkg/install/istio"

	"github.com/solo-io/supergloo/pkg/install"

	istiosecret "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	"github.com/solo-io/supergloo/test/util"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
	istioSync "github.com/solo-io/supergloo/pkg/translator/istio"
)

/*
End to end tests for consul installs with and without mTLS enabled.
Tests assume you already have a Kubernetes environment with Helm / Tiller set up, and with a "supergloo-system" namespace.
The tests will install Consul and get it configured and validate all services up and running, then sync the mesh to set
up any other configuration, then tear down and clean up all resources created.
This will take about 80 seconds with mTLS, and 50 seconds without.
*/
var _ = Describe("Istio Install and Encryption E2E", func() {

	installNamespace := "istio-system"
	superglooNamespace := "supergloo-system" // this needs to be made before running tests
	meshName := "test-istio-mesh"
	secretName := "test-tls-secret"
	kubeCache := kube.NewKubeCache()

	path := os.Getenv("HELM_CHART_PATH")
	if path == "" {
		panic("Set environment variable HELM_CHART_PATH")
	}

	getSnapshot := func(mtls bool, secret *core.ResourceRef) *v1.InstallSnapshot {
		return &v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				installNamespace: v1.InstallList{
					&v1.Install{
						Metadata: core.Metadata{
							Namespace: superglooNamespace,
							Name:      meshName,
						},
						MeshType: &v1.Install_Istio{
							Istio: &v1.Istio{
								InstallationNamespace: installNamespace,
							},
						},
						ChartLocator: &v1.HelmChartLocator{
							Kind: &v1.HelmChartLocator_ChartPath{
								ChartPath: &v1.HelmChartPath{
									Path: path,
								},
							},
						},
						Encryption: &v1.Encryption{
							TlsEnabled: mtls,
							Secret:     secret,
						},
					},
				},
			},
		}
	}

	getTranslatorSnapshot := func(mesh *v1.Mesh, secret *istiosecret.IstioCacertsSecret) *v1.TranslatorSnapshot {
		secrets := istiosecret.IstiocertsByNamespace{}
		if secret != nil {
			secrets = istiosecret.IstiocertsByNamespace{
				superglooNamespace: istiosecret.IstioCacertsSecretList{
					secret,
				},
			}
		}
		return &v1.TranslatorSnapshot{
			Meshes: v1.MeshesByNamespace{
				superglooNamespace: v1.MeshList{
					mesh,
				},
			},
			Istiocerts: secrets,
		}
	}

	var meshClient v1.MeshClient
	var secretClient istiosecret.IstioCacertsSecretClient
	var installSyncer install.InstallSyncer

	BeforeEach(func() {
		meshClient = util.GetMeshClient(kubeCache)
		secretClient = util.GetSecretClient()
		installSyncer = install.InstallSyncer{
			Kube:       util.GetKubeClient(),
			MeshClient: meshClient,
		}
	})

	AfterEach(func() {
		util.UninstallHelmRelease(meshName)
		util.TryDeleteIstioCrds()
		util.TerminateNamespaceBlocking(installNamespace)
		util.DeleteCrb(istio.CrbName)
		if meshClient != nil {
			meshClient.Delete(superglooNamespace, meshName, clients.DeleteOpts{})
		}
		if secretClient != nil {
			secretClient.Delete(superglooNamespace, secretName, clients.DeleteOpts{})
		}
	})

	It("Can install istio with mtls enabled and custom root cert", func() {
		secret, ref := util.CreateTestSecret(superglooNamespace, secretName)
		snap := getSnapshot(true, ref)
		err := installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())

		util.WaitForAvailablePods(installNamespace)
		mesh, err := meshClient.Read(superglooNamespace, meshName, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())

		meshSyncer := istioSync.EncryptionSyncer{
			Kube:         util.GetKubeClient(),
			SecretClient: secretClient,
		}
		syncSnapshot := getTranslatorSnapshot(mesh, secret)
		err = meshSyncer.Sync(context.TODO(), syncSnapshot)
		Expect(err).NotTo(HaveOccurred())

		// Check cert matches
	})

	It("Can install istio without mtls enabled", func() {
		snap := getSnapshot(false, nil)
		err := installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		util.WaitForAvailablePods(installNamespace)

		mesh, err := meshClient.Read(superglooNamespace, meshName, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		meshSyncer := istioSync.EncryptionSyncer{}
		syncSnapshot := getTranslatorSnapshot(mesh, nil)
		err = meshSyncer.Sync(context.TODO(), syncSnapshot)
		Expect(err).NotTo(HaveOccurred())
	})

})
