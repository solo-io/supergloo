package e2e

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/solo-io/supergloo/pkg/secret"

	"github.com/solo-io/supergloo/pkg/install/istio"
	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/supergloo/pkg/install"

	istiosecret "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	"github.com/solo-io/supergloo/test/util"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	gloo "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
	istioSync "github.com/solo-io/supergloo/pkg/translator/istio"

	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
)

/*
End to end tests for istio install and mesh syncing.
*/
var _ = Describe("Istio Install and Encryption E2E", func() {
	defer GinkgoRecover()
	var installNamespace string
	var meshName string
	var secretName string
	kubeCache := kube.NewKubeCache()
	path := os.Getenv("HELM_CHART_PATH")
	if path == "" {
		path = constants.IstioInstallPath
	}

	getSnapshot := func(mtls bool, install bool, secretRef *core.ResourceRef, secret *istiov1.IstioCacertsSecret) *v1.InstallSnapshot {
		secrets := istiosecret.IstiocertsByNamespace{}
		if secret != nil {
			secrets = istiosecret.IstiocertsByNamespace{
				constants.SuperglooNamespace: istiosecret.IstioCacertsSecretList{
					secret,
				},
			}
		}
		return &v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				installNamespace: v1.InstallList{
					&v1.Install{
						Metadata: core.Metadata{
							Namespace: constants.SuperglooNamespace,
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
							Secret:     secretRef,
						},
					},
				},
			},
			Istiocerts: secrets,
		}
	}

	getTranslatorSnapshot := func(mesh *v1.Mesh, secret *istiosecret.IstioCacertsSecret) *v1.TranslatorSnapshot {
		secrets := istiosecret.IstiocertsByNamespace{}
		if secret != nil {
			secrets = istiosecret.IstiocertsByNamespace{
				constants.SuperglooNamespace: istiosecret.IstioCacertsSecretList{
					secret,
				},
			}
		}
		return &v1.TranslatorSnapshot{
			Meshes: v1.MeshesByNamespace{
				constants.SuperglooNamespace: v1.MeshList{
					mesh,
				},
			},
			Istiocerts: secrets,
		}
	}

	var meshClient v1.MeshClient
	var upstreamClient gloo.UpstreamClient

	var secretClient istiosecret.IstioCacertsSecretClient
	var installSyncer install.InstallSyncer

	BeforeEach(func() {
		randStr := helpers.RandString(8)
		installNamespace = "istio-install-test-" + randStr
		meshName = "istio-mesh-test-" + randStr
		secretName = "istio-secret-test-" + randStr
		util.TryCreateNamespace("supergloo-system")
		util.TryCreateNamespace("gloo-system")
		meshClient = util.GetMeshClient(kubeCache)
		upstreamClient = util.GetUpstreamClient(kubeCache)

		secretClient = util.GetSecretClient()
		installSyncer = install.InstallSyncer{
			Kube:         util.GetKubeClient(),
			MeshClient:   meshClient,
			ApiExts:      util.GetApiExtsClient(),
			SecretClient: util.GetSecretClient(),
		}
	})

	AfterEach(func() {
		if meshClient != nil {
			meshClient.Delete(constants.SuperglooNamespace, meshName, clients.DeleteOpts{})
		}
		if secretClient != nil {
			secretClient.Delete(constants.SuperglooNamespace, secretName, clients.DeleteOpts{})
			secretClient.Delete(installNamespace, secret.CustomRootCertificateSecretName, clients.DeleteOpts{})
		}
		util.TerminateNamespaceBlocking("supergloo-system")
		// delete gloo system to remove gloo resources like upstreams
		util.TerminateNamespaceBlocking("gloo-system")

		util.UninstallHelmRelease(meshName)
		util.TryDeleteIstioCrds()
		util.TerminateNamespaceBlocking(installNamespace)
		util.DeleteCrb(istio.CrbName)
	})

	Describe("istio + encryption", func() {
		It("Can install istio with mtls enabled and custom root cert", func() {
			secret, ref := util.CreateTestRsaSecret(constants.SuperglooNamespace, secretName)
			snap := getSnapshot(true, true, ref, secret)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			Expect(util.WaitForAvailablePods(installNamespace)).To(BeEquivalentTo(9))
			// At this point citadel has started up with self-signed to false and a mounted cacerts

			// make sure what's in cacerts is right
			util.CheckCertMatchesIstio(installNamespace)
		})

		It("Can install istio with mtls enabled and deploy custom cert later", func() {
			secret, ref := util.CreateTestRsaSecret(constants.SuperglooNamespace, secretName)
			snap := getSnapshot(true, true, nil, nil)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			Expect(util.WaitForAvailablePods(installNamespace)).To(BeEquivalentTo(9))

			mesh, err := meshClient.Read(constants.SuperglooNamespace, meshName, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			mesh.Encryption.Secret = ref

			syncer := istioSync.EncryptionSyncer{
				SecretClient:   util.GetSecretClient(),
				Kube:           util.GetKubeClient(),
				IstioNamespace: installNamespace,
			}
			err = syncer.Sync(context.TODO(), getTranslatorSnapshot(mesh, secret))
			Expect(err).NotTo(HaveOccurred())

			// syncer will restart citadel
			Expect(util.WaitForAvailablePods(installNamespace)).To(BeEquivalentTo(9))
			// At this point citadel has started up with self-signed to false and a mounted cacerts

			// make sure what's in cacerts is right
			util.CheckCertMatchesIstio(installNamespace)
		})

		It("Can install istio with mtls enabled and self-signing", func() {
			snap := getSnapshot(true, true, nil, nil)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			util.WaitForAvailablePods(installNamespace)
			_, err = meshClient.Read(constants.SuperglooNamespace, meshName, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())

			// make sure default mesh policy and destination rule are created, meaning overrides for security were applied
			cmd := exec.Command("kubectl", "get", "meshpolicy", "default")
			Expect(cmd.Run()).To(BeNil())
			cmd = exec.Command("kubectl", "get", "destinationrule", "default", "-n", installNamespace)
			Expect(cmd.Run()).To(BeNil())

			// TODO: deploy sample app and do more checking
		})

		FIt("Can install istio with mtls disabled and toggle it on", func() {
			snap := getSnapshot(false, true, nil, nil)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(util.WaitForAvailablePods(installNamespace)).To(BeEquivalentTo(9))

			snap = getSnapshot(true, true, nil, nil)
			err = installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			Expect(util.WaitForAvailablePods(installNamespace)).To(BeEquivalentTo(9))

			// TODO: currently no check that mesh policy actually changed from mtls {mode: PERMISSIVE} to mtls: {}
		})
	})

	Describe("istio + policy", func() {
		var (
			bookinfons string
		)

		BeforeEach(func() {
			// TODO: change this to something random once we fix discovery
			// to work with labeled namespaces
			bookinfons = "bookinfo"
		})

		AfterEach(func() {
			gexec.TerminateAndWait(2 * time.Second)
			if bookinfons != "default" {
				util.TerminateNamespaceBlocking(bookinfons)
			}
		})

		deployBookInfo := func() string {
			// create namespace for bookinfo
			ns := &kubecore.Namespace{
				ObjectMeta: kubemeta.ObjectMeta{
					Name: bookinfons,
					Labels: map[string]string{
						"istio-injection": "enabled",
					},
				},
			}
			util.GetKubeClient().CoreV1().Namespaces().Create(ns)

			bookinfo := "https://raw.githubusercontent.com/istio/istio/4c0a001b5e542d43b4c66ae75c1f41f2c1ff183e/samples/bookinfo/platform/kube/bookinfo.yaml"
			kubectlargs := []string{"apply", "-n", bookinfons, "-f", bookinfo}
			cmd := exec.Command("kubectl", kubectlargs...)
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			return bookinfons
		}

		It("Should install istio and enable policy", func() {

			// start discovery
			cmd := exec.Command(PathToUds, "-discover", bookinfons)
			_, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

			snap := getSnapshot(true, true, nil, nil)
			err = installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			util.WaitForAvailablePodsWithTimeout(installNamespace, "300s")

			deployBookInfo()
			util.WaitForAvailablePodsWithTimeout(bookinfons, "500s")

			mesh, err := meshClient.Read(constants.SuperglooNamespace, meshName, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())

			mesh.Policy = &v1.Policy{
				Rules: []*v1.Rule{
					{
						Source: &core.ResourceRef{
							Name:      "default-reviews-9080",
							Namespace: "gloo-system",
						},
						Destination: &core.ResourceRef{
							Name:      "default-ratings-9080",
							Namespace: "gloo-system",
						},
					},
				},
			}

			meshSyncer, err := istioSync.NewPolicySyncer("supergloo-system", kubeCache, util.GetKubeConfig())
			Expect(err).NotTo(HaveOccurred())

			getupstreamnames := func() ([]string, error) {
				return util.GetUpstreamNames(upstreamClient)
			}
			Eventually(getupstreamnames, "60s", "1s").ShouldNot(HaveLen(0))

			syncSnapshot := getTranslatorSnapshot(mesh, nil)
			ups, err := upstreamClient.List("gloo-system", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			syncSnapshot.Upstreams = gloo.UpstreamsByNamespace{
				"gloo-system": ups,
			}

			err = meshSyncer.Sync(context.TODO(), syncSnapshot)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
