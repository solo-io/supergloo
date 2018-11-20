package e2e

import (
	"context"

	"github.com/solo-io/supergloo/pkg/install"

	gloo "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/test/util"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/consul"
	consulSync "github.com/solo-io/supergloo/pkg/translator/consul"

	helmkube "k8s.io/helm/pkg/kube"
)

/*
End to end tests for consul installs with and without mTLS enabled.
Tests assume you already have a Kubernetes environment with Helm / Tiller set up, and with a "supergloo-system" namespace.
The tests will install Consul and get it configured and validate all services up and running, then sync the mesh to set
up any other configuration, then tear down and clean up all resources created.
This will take about 80 seconds with mTLS, and 50 seconds without.
*/
var _ = Describe("Consul Install and Encryption E2E", func() {

	installNamespace := "consul"
	superglooNamespace := "supergloo-system" // this needs to be made before running tests
	meshName := "test-consul-mesh"
	secretName := "test-tls-secret"
	consulPort := 8500
	kubeCache := kube.NewKubeCache()

	getSnapshot := func(mtls bool, secret *core.ResourceRef) *v1.InstallSnapshot {
		return &v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				installNamespace: v1.InstallList{
					&v1.Install{
						Metadata: core.Metadata{
							Namespace: superglooNamespace,
							Name:      meshName,
						},
						MeshType: &v1.Install_Consul{
							Consul: &v1.Consul{
								InstallationNamespace: installNamespace,
							},
						},
						ChartLocator: &v1.HelmChartLocator{
							Kind: &v1.HelmChartLocator_ChartPath{
								ChartPath: &v1.HelmChartPath{
									Path: "https://github.com/hashicorp/consul-helm/archive/v0.3.0.tar.gz",
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

	getTranslatorSnapshot := func(mesh *v1.Mesh, secret *gloo.Secret) *v1.TranslatorSnapshot {
		secrets := gloo.SecretsByNamespace{}
		if secret != nil {
			secrets = gloo.SecretsByNamespace{
				superglooNamespace: gloo.SecretList{
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
			Secrets: secrets,
		}
	}

	var tunnel *helmkube.Tunnel
	var meshClient v1.MeshClient
	var secretClient gloo.SecretClient
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
		// Delete secret
		if tunnel != nil {
			tunnel.Close()
			tunnel = nil
		}
		if meshClient != nil {
			meshClient.Delete(superglooNamespace, meshName, clients.DeleteOpts{})
		}
		if secretClient != nil {
			secretClient.Delete(superglooNamespace, secretName, clients.DeleteOpts{})
		}

		util.DeleteWebhookConfigIfExists(consul.WebhookCfg)
		util.DeleteCrb(consul.CrbName)
		util.TerminateNamespaceBlocking(installNamespace)
	})

	It("Can install consul with mtls enabled and custom root cert", func() {
		secret, ref := util.CreateTestSecret(superglooNamespace, secretName)
		snap := getSnapshot(true, ref)
		err := installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())

		util.WaitForAvailablePods(installNamespace)
		mesh, err := meshClient.Read(superglooNamespace, meshName, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		tunnel, err = util.CreateConsulTunnel(installNamespace, consulPort)
		Expect(err).NotTo(HaveOccurred())

		meshSyncer := consulSync.ConsulSyncer{
			LocalPort: tunnel.Local,
		}
		syncSnapshot := getTranslatorSnapshot(mesh, secret)
		err = meshSyncer.Sync(context.TODO(), syncSnapshot)
		Expect(err).NotTo(HaveOccurred())

		util.CheckCertMatches(tunnel.Local, util.TestRoot)
	})

	It("Can install consul without mtls enabled", func() {
		snap := getSnapshot(false, nil)
		err := installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		util.WaitForAvailablePods(installNamespace)

		mesh, err := meshClient.Read(superglooNamespace, meshName, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		meshSyncer := consulSync.ConsulSyncer{}
		syncSnapshot := getTranslatorSnapshot(mesh, nil)
		err = meshSyncer.Sync(context.TODO(), syncSnapshot)
		Expect(err).NotTo(HaveOccurred())
	})

})
