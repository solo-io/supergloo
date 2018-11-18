package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gloo "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/hashicorp/consul/api"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	kubecore "k8s.io/api/core/v1"

	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
var _ = Describe("ConsulInstallSyncer", func() {

	installNamespace := "consul"
	superglooNamespace := "supergloo-system" // this needs to be made before running tests
	meshName := "test-consul-mesh"
	secretName := "test-tls-secret"
	consulPort := 8500

	getKubeConfig := func() *rest.Config {
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		Expect(err).NotTo(HaveOccurred())
		return cfg
	}

	getKubeClient := func() *kubernetes.Clientset {
		cfg := getKubeConfig()
		client, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
		return client
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
						Consul: &v1.ConsulInstall{
							Path:      "https://github.com/hashicorp/consul-helm/archive/v0.3.0.tar.gz",
							Namespace: installNamespace,
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

	terminateNamespaceBlocking := func() {
		client := getKubeClient()
		client.CoreV1().Namespaces().Delete(installNamespace, &kubemeta.DeleteOptions{})
		Eventually(func() error {
			_, err := client.CoreV1().Namespaces().Get(installNamespace, kubemeta.GetOptions{})
			return err
		}, "60s", "1s").ShouldNot(BeNil()) // will be non-nil when NS is gone
	}

	waitForAvailablePods := func() {
		client := getKubeClient()
		Eventually(func() bool {
			podList, err := client.CoreV1().Pods(installNamespace).List(kubemeta.ListOptions{})
			Expect(err).To(BeNil())
			done := true
			for _, pod := range podList.Items {
				for _, condition := range pod.Status.Conditions {
					if condition.Type == kubecore.PodReady && condition.Status != kubecore.ConditionTrue {
						done = false
					}
				}
			}
			return done
		}, "60s", "1s").Should(BeTrue())
	}

	kubeCache := kube.NewKubeCache()

	getMeshClient := func(restConfig *rest.Config) v1.MeshClient {
		meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
			Crd:         v1.MeshCrd,
			Cfg:         restConfig,
			SharedCache: kubeCache,
		})
		Expect(err).Should(BeNil())
		err = meshClient.Register()
		Expect(err).Should(BeNil())
		return meshClient
	}

	deleteCrb := func() {
		client := getKubeClient()
		err := client.RbacV1().ClusterRoleBindings().Delete(consul.CrbName, &kubemeta.DeleteOptions{})
		Expect(err).Should(BeNil())
	}

	deleteWebhookConfigIfExists := func() {
		client := getKubeClient()
		client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(consul.WebhookCfg, &kubemeta.DeleteOptions{})
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

	testKey := "-----BEGIN PRIVATE KEY-----\nMIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDBoI1sMdiOTvBBdjWlS\nZ8qwNuK9xV4yKuboLZ4Sx/OBfy1eKZocxTKvnjLrHUe139uhZANiAAQMTIR56O8U\nTIqf6uUHM4i9mZYLj152up7elS06Gi6lk7IeUQDHxP0NnOnbhC7rmtOV6myLNApL\nQ92kZKg7qa8q7OY/4w1QfC4ch7zZKxjNkSIiuAx7V/lzF6FYDcqT3js=\n-----END PRIVATE KEY-----"
	testRoot := "-----BEGIN CERTIFICATE-----\nMIIB7jCCAXUCCQC2t6Lqc2xnXDAKBggqhkjOPQQDAjBhMQswCQYDVQQGEwJVUzEW\nMBQGA1UECAwNTWFzc2FjaHVzZXR0czESMBAGA1UEBwwJQ2FtYnJpZGdlMQwwCgYD\nVQQKDANPcmcxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xODExMTgxMzQz\nMDJaFw0xOTExMTgxMzQzMDJaMGExCzAJBgNVBAYTAlVTMRYwFAYDVQQIDA1NYXNz\nYWNodXNldHRzMRIwEAYDVQQHDAlDYW1icmlkZ2UxDDAKBgNVBAoMA09yZzEYMBYG\nA1UEAwwPd3d3LmV4YW1wbGUuY29tMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEDEyE\neejvFEyKn+rlBzOIvZmWC49edrqe3pUtOhoupZOyHlEAx8T9DZzp24Qu65rTleps\nizQKS0PdpGSoO6mvKuzmP+MNUHwuHIe82SsYzZEiIrgMe1f5cxehWA3Kk947MAoG\nCCqGSM49BAMCA2cAMGQCMCytVFc8sBdbM7DaBCz0N2ptdb0T7LFFfxDTzn4gjiDq\nVCd/3dct21TUWsthKXF2VgIwXEMI5EQiJ5kjR/y1KNBC9b4wfDiKRvG33jYe9gn6\ntzXUS00SoqG9D27/7aK71/xv\n-----END CERTIFICATE-----"
	testCertChain := ""

	getSecretClient := func(kube *kubernetes.Clientset) gloo.SecretClient {
		secretClient, err := gloo.NewSecretClient(&factory.KubeSecretClientFactory{
			Clientset: kube,
		})
		Expect(err).Should(BeNil())
		err = secretClient.Register()
		Expect(err).Should(BeNil())
		return secretClient
	}

	createSecret := func(secretClient gloo.SecretClient) (*gloo.Secret, *core.ResourceRef) {
		tls := gloo.TlsSecret{
			RootCa:     testRoot,
			PrivateKey: testKey,
			CertChain:  testCertChain,
		}
		tlsWrapper := gloo.Secret_Tls{
			Tls: &tls,
		}
		secret := &gloo.Secret{
			Metadata: core.Metadata{
				Namespace: superglooNamespace,
				Name:      secretName,
			},
			Kind: &tlsWrapper,
		}
		secretClient.Delete(superglooNamespace, secretName, clients.DeleteOpts{})
		_, err := secretClient.Write(secret, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
		ref := &core.ResourceRef{
			Namespace: superglooNamespace,
			Name:      secretName,
		}
		return secret, ref
	}

	getConsulServerPodName := func(client kubernetes.Interface) string {
		podList, err := client.CoreV1().Pods(installNamespace).List(kubemeta.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, pod := range podList.Items {
			if strings.Contains(pod.Name, "consul-server-0") {
				return pod.Name
			}
		}
		// Should not have happened
		Expect(false).To(BeTrue())
		return ""
	}

	// New creates a new and initialized tunnel.
	createConsulTunnel := func(client kubernetes.Interface) (*helmkube.Tunnel, error) {
		podName := getConsulServerPodName(client)
		t := helmkube.NewTunnel(client.CoreV1().RESTClient(), getKubeConfig(), installNamespace, podName, consulPort)
		return t, t.ForwardPort()
	}

	checkCertUpdated := func(port int) {
		config := &api.Config{
			Address: fmt.Sprintf("127.0.0.1:%d", port),
		}
		client, err := api.NewClient(config)
		Expect(err).NotTo(HaveOccurred())
		var queryOpts api.QueryOptions
		currentConfig, _, err := client.Connect().CAGetConfig(&queryOpts)
		Expect(err).NotTo(HaveOccurred())

		currentRoot := currentConfig.Config["RootCert"]
		Expect(currentRoot).To(BeEquivalentTo(testRoot))
	}

	var tunnel *helmkube.Tunnel
	var meshClient v1.MeshClient
	var secretClient gloo.SecretClient

	AfterEach(func() {
		fmt.Printf("Cleaning up after test")
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

		deleteWebhookConfigIfExists()
		deleteCrb()
		terminateNamespaceBlocking()
	})

	It("Can install consul with mtls enabled", func() {
		fmt.Print("Setting up clients\n")
		kubeCfg := getKubeConfig()
		secretClient = getSecretClient(getKubeClient())
		secret, ref := createSecret(secretClient)
		meshClient = getMeshClient(kubeCfg)

		fmt.Printf("Install mesh\n")
		installSyncer := consul.ConsulInstallSyncer{
			Kube:       getKubeClient(),
			MeshClient: meshClient,
		}
		snap := getSnapshot(true, ref)
		err := installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())

		fmt.Printf("Wait for pods to be ready\n")
		waitForAvailablePods()

		fmt.Printf("Reading mesh object\n")
		mesh, err := meshClient.Read(superglooNamespace, meshName, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())

		fmt.Printf("Opening consul tunnel\n")
		tunnel, err = createConsulTunnel(getKubeClient())
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Syncing mesh\n")
		meshSyncer := consulSync.ConsulSyncer{
			LocalPort: tunnel.Local,
		}
		syncSnapshot := getTranslatorSnapshot(mesh, secret)
		err = meshSyncer.Sync(context.TODO(), syncSnapshot)
		Expect(err).NotTo(HaveOccurred())

		fmt.Printf("Validating cert got updated\n")
		checkCertUpdated(tunnel.Local)
	})

	It("Can install consul without mtls enabled", func() {
		meshClient = getMeshClient(getKubeConfig())
		installSyncer := consul.ConsulInstallSyncer{
			Kube:       getKubeClient(),
			MeshClient: meshClient,
		}
		snap := getSnapshot(false, nil)
		err := installSyncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		waitForAvailablePods()

		mesh, err := meshClient.Read(superglooNamespace, meshName, clients.ReadOpts{})
		Expect(err).NotTo(HaveOccurred())
		meshSyncer := consulSync.ConsulSyncer{}
		syncSnapshot := getTranslatorSnapshot(mesh, nil)
		err = meshSyncer.Sync(context.TODO(), syncSnapshot)
		Expect(err).NotTo(HaveOccurred())
	})

})
