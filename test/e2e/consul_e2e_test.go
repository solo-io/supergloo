package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"

	istiosecret "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"

	"github.com/hashicorp/consul/api"

	"github.com/solo-io/supergloo/pkg/install"

	gloo "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/test/util"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/install/consul"
	consulSync "github.com/solo-io/supergloo/pkg/translator/consul"

	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	helmkube "k8s.io/helm/pkg/kube"
)

/*
End to end tests for consul workflows
*/
var _ = Describe("Consul E2E", func() {

	var namespace = helpers.RandString(6)
	const (
		meshName   = "test-consul-mesh"
		secretName = "test-tls-secret"
		consulPort = 8500
	)

	kubeCache := kube.NewKubeCache()

	var (
		tunnel         *helmkube.Tunnel
		meshClient     v1.MeshClient
		secretClient   istiosecret.IstioCacertsSecretClient
		upstreamClient gloo.UpstreamClient
		installSyncer  install.InstallSyncer
		pathToUds      string
	)

	createInstallSnapshot := func(mtls bool, secret *core.ResourceRef, enable bool) *v1.InstallSnapshot {
		return &v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				namespace: v1.InstallList{
					&v1.Install{
						Metadata: core.Metadata{
							Namespace: namespace,
							Name:      meshName,
						},
						MeshType: &v1.Install_Consul{
							Consul: &v1.Consul{
								InstallationNamespace: namespace,
							},
						},
						ChartLocator: &v1.HelmChartLocator{
							Kind: &v1.HelmChartLocator_ChartPath{
								ChartPath: &v1.HelmChartPath{
									Path: "https://github.com/hashicorp/consul-helm/archive/5daf413626046d31dcb1030db889a7c96e078a1c.tar.gz", // this is old: https://github.com/hashicorp/consul-helm/archive/v0.3.0.tar.gz",
								},
							},
						},
						Encryption: &v1.Encryption{
							TlsEnabled: mtls,
							Secret:     secret,
						},
						Enabled: &types.BoolValue{
							Value: enable,
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
				namespace: istiosecret.IstioCacertsSecretList{
					secret,
				},
			}
		}
		return &v1.TranslatorSnapshot{
			Meshes: v1.MeshesByNamespace{
				namespace: v1.MeshList{
					mesh,
				},
			},
			Istiocerts: secrets,
		}
	}

	BeforeEach(func() {
		namespace = helpers.RandString(8)
		err := setup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		setup.TeardownKube(namespace)
	})

	BeforeEach(func() {
		util.TryCreateNamespace("supergloo-system")
		util.TryCreateNamespace("gloo-system")
		pathToUds = PathToUds // set up by before suite
		meshClient = util.GetMeshClient(kubeCache)
		upstreamClient = util.GetUpstreamClient(kubeCache)
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
			meshClient.Delete(namespace, meshName, clients.DeleteOpts{})
		}
		if secretClient != nil {
			secretClient.Delete(namespace, secretName, clients.DeleteOpts{})
		}
		util.DeleteWebhookConfigIfExists(consul.WebhookCfg)
		util.DeleteCrb(consul.CrbName)
		util.TerminateNamespaceBlocking(namespace)
		util.UninstallHelmRelease(meshName)
		util.TerminateNamespaceBlocking("supergloo-system")
		// delete gloo system to remove gloo resources like upstreams
		util.TerminateNamespaceBlocking("gloo-system")
		gexec.TerminateAndWait(2 * time.Second)
	})

	It("Can install consul with mtls enabled and custom root cert", func() {
		secret, ref := util.CreateTestEcSecret(namespace, secretName)
		snap := createInstallSnapshot(true, ref, true)
		err := installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())

		util.WaitForAvailablePods(namespace)
		mesh, err := meshClient.Read(namespace, meshName, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		tunnel, err = util.CreateConsulTunnel(namespace, consulPort)
		Expect(err).NotTo(HaveOccurred())

		meshSyncer := consulSync.ConsulSyncer{
			LocalPort: tunnel.Local,
		}
		syncSnapshot := getTranslatorSnapshot(mesh, secret)
		err = meshSyncer.Sync(context.TODO(), syncSnapshot)
		Expect(err).NotTo(HaveOccurred())

		util.CheckCertMatchesConsul(tunnel.Local, util.TestEcRoot)

		snap = createInstallSnapshot(true, ref, false)
		err = installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		util.WaitForDeletedPods(namespace)
	})

	Describe("consul + policy", func() {

		var (
			bookinfons string
		)
		BeforeEach(func() {
			bookinfons = "gloo-system"

		})

		AfterEach(func() {
			util.TerminateNamespaceBlocking(bookinfons)
		})

		/*
			deployBookInfo := func() string {
				// create namespace for bookinfo
				ns := &kubecore.Namespace{
					ObjectMeta: kubemeta.ObjectMeta{
						Name: bookinfons,
					},
				}
				util.GetKubeClient().CoreV1().Namespaces().Create(ns)

				bookinfo := "https://raw.githubusercontent.com/istio/istio/4c0a001b5e542d43b4c66ae75c1f41f2c1ff183e/samples/bookinfo/platform/kube/bookinfo.yaml"
				kubectlargs := []string{"apply", "-n", bookinfons, "-f", bookinfo}
				cmd := exec.Command("kubectl", kubectlargs...)
				err := cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				util.WaitForAvailablePods(bookinfons)
				return bookinfons
			}
		*/

		deployHelloWorld := func() string {
			// create namespace for bookinfo
			ns := &kubecore.Namespace{
				ObjectMeta: kubemeta.ObjectMeta{
					Name: bookinfons,
				},
			}
			util.GetKubeClient().CoreV1().Namespaces().Create(ns)

			bookinfo := filepath.Join(os.Getenv("GOPATH"), "src/github.com/solo-io/supergloo", "test/e2e/hello_consul.yaml")
			kubectlargs := []string{"apply", "-n", bookinfons, "-f", bookinfo}
			cmd := exec.Command("kubectl", kubectlargs...)
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			util.WaitForAvailablePods(bookinfons)
			return bookinfons
		}

		It("Can change consul policy", func() {
			snap := createInstallSnapshot(true, nil, true)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			util.WaitForAvailablePods(namespace)

			mesh, err := meshClient.Read(namespace, meshName, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())

			// portforward consul to here

			tunnel, err = util.CreateConsulTunnel(namespace, consulPort)
			Expect(err).NotTo(HaveOccurred())

			localport := tunnel.Local

			// start discovery
			cmd := exec.Command(pathToUds, "-discover", bookinfons)
			cmd.Env = os.Environ()
			addr := fmt.Sprintf("localhost:%d", localport)
			cmd.Env = append(cmd.Env, "CONSUL_HTTP_ADDR="+addr)
			_, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			os.Setenv("CONSUL_HTTP_ADDR", addr)

			deployHelloWorld()

			mesh.Policy = &v1.Policy{
				Rules: []*v1.Rule{
					{
						Source: &core.ResourceRef{
							Name:      "static-client",
							Namespace: bookinfons,
						},
						Destination: &core.ResourceRef{
							Name:      "static-server",
							Namespace: bookinfons,
						},
					},
				},
			}

			meshSyncer := consulSync.PolicySyncer{}
			syncSnapshot := getTranslatorSnapshot(mesh, nil)

			getupstreamnames := func() ([]string, error) {
				return util.GetUpstreamNames(upstreamClient)
			}

			Eventually(getupstreamnames, "60s", "1s").Should(ContainElement("static-client"))
			Eventually(getupstreamnames, "10s", "1s").Should(ContainElement("static-server"))

			ups, err := upstreamClient.List("gloo-system", clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())

			syncSnapshot.Upstreams = gloo.UpstreamsByNamespace{
				"gloo-system": ups,
			}

			// sync the snapshot to have the intentions created
			err = meshSyncer.Sync(context.TODO(), syncSnapshot)
			Expect(err).NotTo(HaveOccurred())

			client, err := api.NewClient(api.DefaultConfig())
			Expect(err).NotTo(HaveOccurred())
			intentions, _, err := client.Connect().Intentions(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(intentions).To(HaveLen(1))
		})

	})

})
