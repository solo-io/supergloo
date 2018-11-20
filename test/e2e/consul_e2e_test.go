package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

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
End to end tests for consul installs with and without mTLS enabled.
Tests assume you already have a Kubernetes environment with Helm / Tiller set up, and with a "supergloo-system" namespace.
The tests will install Consul and get it configured and validate all services up and running, then sync the mesh to set
up any other configuration, then tear down and clean up all resources created.
This will take about 80 seconds with mTLS, and 50 seconds without.
*/
var _ = Describe("Consul Install and Encryption E2E", func() {

	const (
		installNamespace   = "consul"
		superglooNamespace = "supergloo-system" // this needs to be made before running tests
		meshName           = "test-consul-mesh"
		secretName         = "test-tls-secret"
		consulPort         = 8500
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

	BeforeEach(func() {
		ns := &kubecore.Namespace{
			ObjectMeta: kubemeta.ObjectMeta{
				Name: "supergloo-system",
			},
		}
		util.GetKubeClient().CoreV1().Namespaces().Create(ns)

		ns = &kubecore.Namespace{
			ObjectMeta: kubemeta.ObjectMeta{
				Name: "gloo-system",
			},
		}
		util.GetKubeClient().CoreV1().Namespaces().Create(ns)
	})

	AfterEach(func() {
		util.TerminateNamespaceBlocking("supergloo-system")
		// delete gloo system to remove gloo resources like upstreams
		util.TerminateNamespaceBlocking("gloo-system")
	})

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
									Path: "https://github.com/hashicorp/consul-helm/archive/5daf413626046d31dcb1030db889a7c96e078a1c.tar.gz", // this is old: https://github.com/hashicorp/consul-helm/archive/v0.3.0.tar.gz",
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

	AfterEach(func() {
		gexec.TerminateAndWait(2 * time.Second)
	})

	BeforeEach(func() {
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
			meshClient.Delete(superglooNamespace, meshName, clients.DeleteOpts{})
		}
		if secretClient != nil {
			secretClient.Delete(superglooNamespace, secretName, clients.DeleteOpts{})
		}
		util.DeleteWebhookConfigIfExists(consul.WebhookCfg)
		util.DeleteCrb(consul.CrbName)
		util.TerminateNamespaceBlocking(installNamespace)
		util.UninstallHelmRelease(meshName)
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

		util.CheckCertMatchesConsul(tunnel.Local, util.TestRoot)
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

	Describe("consul + policy", func() {

		var (
			bookinfons string
		)
		BeforeEach(func() {
			bookinfons = "bookinfo"

		})

		AfterEach(func() {
			// TODO: use helper
			util.GetKubeClient().CoreV1().Namespaces().Delete(bookinfons, &kubemeta.DeleteOptions{})
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
			snap := getSnapshot(true, nil)
			err := installSyncer.Sync(context.TODO(), snap)
			Expect(err).NotTo(HaveOccurred())

			util.WaitForAvailablePods(installNamespace)

			mesh, err := meshClient.Read(superglooNamespace, meshName, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())

			// portforward consul to here

			tunnel, err = util.CreateConsulTunnel(installNamespace, consulPort)
			Expect(err).NotTo(HaveOccurred())

			localport := tunnel.Local

			// start discovery
			cmd := exec.Command(pathToUds, "-udsonly")
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
							Namespace: "gloo-system",
						},
						Destination: &core.ResourceRef{
							Name:      "static-server",
							Namespace: "gloo-system",
						},
					},
				},
			}

			meshSyncer := consulSync.PolicySyncer{}
			syncSnapshot := getTranslatorSnapshot(mesh, nil)

			getupstreamnames := func() ([]string, error) {
				ul, err := upstreamClient.List("gloo-system", clients.ListOpts{})
				if err != nil {
					return nil, err
				}
				ups := []string{}
				for _, up := range ul {
					ups = append(ups, up.Metadata.Name)
				}
				return ups, nil
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
