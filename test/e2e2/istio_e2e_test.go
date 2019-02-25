package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"

	"github.com/solo-io/supergloo/pkg2/factory"
	kube_client "github.com/solo-io/supergloo/pkg2/kube"
	"k8s.io/client-go/kubernetes"

	"github.com/solo-io/solo-kit/pkg/utils/contextutils"

	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/supergloo/pkg2/constants"

	"github.com/solo-io/supergloo/pkg2/secret"

	"github.com/solo-io/supergloo/pkg2/install/istio"
	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/supergloo/pkg2/install"

	istiosecret "github.com/solo-io/supergloo/pkg2/api/external/istio/encryption/v1"
	"github.com/solo-io/supergloo/test/util"

	gloo "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg2/api/v1"
	istioSync "github.com/solo-io/supergloo/pkg2/translator/istio"

	istiov1 "github.com/solo-io/supergloo/pkg2/api/external/istio/encryption/v1"
)

/*
End to end tests for istio install and mesh syncing.
*/
var _ = Describe("Istio Install and Encryption E2E", func() {
	defer GinkgoRecover()
	var installNamespace string
	var meshName string
	var secretName string
	path := os.Getenv("HELM_CHART_PATH")
	if path == "" {
		path = constants.IstioInstallPath
	}

	getSnapshot := func(mtls, install bool, secretRef *core.ResourceRef, secret *istiov1.IstioCacertsSecret) *v1.InstallSnapshot {
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
						Enabled: &types.BoolValue{Value: install},
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
	var installSyncer *install.InstallSyncer
	var bookinfoNamespace string
	logger := contextutils.LoggerFrom(context.TODO())

	BeforeEach(func() {
		randStr := helpers.RandString(8)
		installNamespace = "istio-install-test-" + randStr
		meshName = "istio-mesh-test-" + randStr
		secretName = "istio-secret-test-" + randStr
		bookinfoNamespace = "istio-bookinfo-test-" + randStr
		util.TryCreateNamespace("supergloo-system")
		util.TryCreateNamespace("gloo-system")
		meshClient = util.GetMeshClient()
		upstreamClient = util.GetUpstreamClient()

		secretClient = util.GetSecretClient()
		var err error
		installSyncer, err = install.NewKubeInstallSyncer(meshClient, secretClient, util.GetKubeClient(), util.GetApiExtsClient())
		Expect(err).NotTo(HaveOccurred())
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
		util.DeleteCrb(istio.CrbName)
		util.TerminateNamespaceBlocking(installNamespace) // can't install a new istio due to nodePort conflicts
		util.TerminateNamespace(bookinfoNamespace)
	})

	deployBookInfoAndWaitForAvailable := func() string {
		// create namespace for bookinfo
		ns := &kubecore.Namespace{
			ObjectMeta: kubemeta.ObjectMeta{
				Name: bookinfoNamespace,
				Labels: map[string]string{
					"istio-injection": "enabled",
				},
			},
		}
		logger.Infof("Creating namespace %s", bookinfoNamespace)
		util.GetKubeClient().CoreV1().Namespaces().Create(ns)

		bookinfo := "https://raw.githubusercontent.com/istio/istio/4c0a001b5e542d43b4c66ae75c1f41f2c1ff183e/samples/bookinfo/platform/kube/bookinfo.yaml"
		kubectlargs := []string{"apply", "-n", bookinfoNamespace, "-f", bookinfo}
		cmd := exec.Command("kubectl", kubectlargs...)
		logger.Infof("Deploying bookinfo")
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred())
		logger.Infof("Waiting until bookinfo pods are ready")
		Eventually(func() int {
			return util.WaitForAvailablePodsWithTimeout(bookinfoNamespace, "500s")
		}, 500*time.Second, time.Second).Should(BeEquivalentTo(6))
		logger.Infof("Bookinfo pods are ready")
		return bookinfoNamespace
	}

	Context("istio + encryption", func() {
		It("Can install istio with mtls enabled and custom root cert", func() {
			secret, ref := util.CreateTestRsaSecret(constants.SuperglooNamespace, secretName)
			snap := getSnapshot(true, true, ref, secret)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int {
				return util.WaitForAvailablePods(installNamespace)
			}, 500*time.Second, time.Second).Should(BeEquivalentTo(9))

			// At this point citadel has started up with self-signed to false and a mounted cacerts

			// make sure what's in cacerts is right
			util.CheckCertMatchesIstio(installNamespace)
		})

		It("Can install istio with mtls enabled and deploy custom cert later", func() {
			testSecret, ref := util.CreateTestRsaSecret(constants.SuperglooNamespace, secretName)
			snap := getSnapshot(true, true, nil, nil)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int {
				return util.WaitForAvailablePods(installNamespace)
			}, 500*time.Second, time.Second).Should(BeEquivalentTo(9))

			mesh, err := meshClient.Read(constants.SuperglooNamespace, meshName, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			mesh.Encryption.Secret = ref

			kubeClient, err := kubernetes.NewForConfig(util.GetKubeConfig())
			Expect(err).NotTo(HaveOccurred())

			secretClient := kube_client.NewKubeSecretClient(kubeClient)
			podClient := kube_client.NewKubePodClient(kubeClient)

			istioSecretClient, err := factory.GetIstioCacertsSecretClient(kubeClient)
			Expect(err).NotTo(HaveOccurred())
			istioSecretClient.Register()
			Expect(err).NotTo(HaveOccurred())

			secretSyncer := &secret.KubeSecretSyncer{
				SecretClient:      secretClient,
				PodClient:         podClient,
				IstioSecretClient: istioSecretClient,
			}

			syncer := istioSync.EncryptionSyncer{
				SecretSyncer: secretSyncer,
			}
			err = syncer.Sync(context.TODO(), getTranslatorSnapshot(mesh, testSecret))
			Expect(err).NotTo(HaveOccurred())

			// syncer will restart citadel
			Eventually(func() int {
				return util.WaitForAvailablePods(installNamespace)
			}, 500*time.Second, time.Second).Should(BeEquivalentTo(9))
			// At this point citadel has started up with self-signed to false and a mounted cacerts

			// make sure what's in cacerts is right
			util.CheckCertMatchesIstio(installNamespace)
		})

		curlSucceeds := func(plainText bool) bool {
			pod, err := util.GetPodWithSubstringInName(bookinfoNamespace, "productpage")
			Expect(err).NotTo(HaveOccurred())
			s := "s"
			if plainText {
				s = ""
			}
			curlUrl := fmt.Sprintf("http%s://%s:9080/productpage", s, pod.Status.PodIP)
			kubectlArgs := []string{
				"exec",
				"-n",
				bookinfoNamespace,
				pod.Name,
				"-c",
				"istio-proxy",
				"--",
				"curl",
				curlUrl,
			}

			if !plainText {
				kubectlArgs = append(kubectlArgs, "--key", "/etc/certs/key.pem", "--cert", "/etc/certs/cert-chain.pem", "--cacert", "/etc/certs/root-cert.pem", "-k")
			}

			argsString := strings.Join(kubectlArgs, " ")
			logger.Infof("Curling productpage with command: kubectl %s\n", argsString)
			cmd := exec.Command("kubectl", kubectlArgs...)
			err = cmd.Run()
			if err != nil {
				logger.Infof("Curling product page failed.")
			}
			return err == nil
		}

		plainTextCurlSucceeds := func() bool {
			return curlSucceeds(true)
		}

		tlsCurlSucceeds := func() bool {
			return curlSucceeds(false)
		}

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

			// Shouldn't be able to plain text curl from inside the pod
			deployBookInfoAndWaitForAvailable()
			Eventually(tlsCurlSucceeds, "120s", "1s").Should(BeTrue()) // can take awhile after pod starts up for this to work
			Expect(curlSucceeds(true)).To(BeFalse())
		})

		It("Can install istio with mtls disabled and toggle it on", func() {
			// deploy istio and bookinfo with gateway
			snap := getSnapshot(false, true, nil, nil)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() int {
				return util.WaitForAvailablePods(installNamespace)
			}, 500*time.Second, time.Second).Should(BeEquivalentTo(9))

			deployBookInfoAndWaitForAvailable()

			// plain text and tls curling should work from inside the pod
			Eventually(plainTextCurlSucceeds, "120s", "1s").Should(BeTrue()) // can take awhile after pod starts up for this to work
			Expect(tlsCurlSucceeds()).To(BeTrue())

			// turn on mtls in permissive mode
			snap = getSnapshot(true, true, nil, nil)
			err = installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() int {
				return util.WaitForAvailablePods(installNamespace)
			}, 500*time.Second, time.Second).Should(BeEquivalentTo(9))

			// permissive mode is enabled by default
			Expect(tlsCurlSucceeds()).To(BeTrue())
			Expect(plainTextCurlSucceeds()).To(BeTrue())

		})
	})

	Context("istio + policy", func() {

		AfterEach(func() {
			gexec.TerminateAndWait(2 * time.Second)
		})

		It("Should install istio and enable policy", func() {
			snap := getSnapshot(true, true, nil, nil)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())
			util.WaitForAvailablePodsWithTimeout(installNamespace, "300s")

			deployBookInfoAndWaitForAvailable()

			// start discovery
			cmd := exec.Command(PathToUds, "--namespace="+bookinfoNamespace)
			_, err = gexec.Start(cmd, os.Stdout, os.Stderr)
			Expect(err).NotTo(HaveOccurred())

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

			meshSyncer, err := istioSync.NewPolicySyncer("supergloo-system", kube.NewKubeCache(), util.GetKubeConfig())
			Expect(err).NotTo(HaveOccurred())

			getupstreamnames := func() ([]string, error) {
				return util.GetUpstreamNames(upstreamClient, bookinfoNamespace)
			}
			Eventually(getupstreamnames, "60s", "1s").ShouldNot(HaveLen(0))

			syncSnapshot := getTranslatorSnapshot(mesh, nil)
			ups, err := upstreamClient.List(bookinfoNamespace, clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			syncSnapshot.Upstreams = gloo.UpstreamsByNamespace{
				bookinfoNamespace: ups,
			}

			err = meshSyncer.Sync(context.TODO(), syncSnapshot)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
