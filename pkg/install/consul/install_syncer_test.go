package consul_test

import (
	"context"
	"os"
	"path/filepath"

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
)

/*
End to end tests for consul installs with and without mTLS enabled.
Tests assume you already have a Kubernetes environment with Helm / Tiller set up, and with a "supergloo-system" namespace.
The tests will install Consul and get it configured and validate all services up and running, then tear down and
clean up all resources created. This will take about 45 seconds with mTLS, and 20 seconds without.
*/
var _ = Describe("ConsulInstallSyncer", func() {

	namespace := "consul"

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

	getSnapshot := func(mtls bool) *v1.InstallSnapshot {
		return &v1.InstallSnapshot{
			Installs: v1.InstallsByNamespace{
				"not_used": v1.InstallList{
					&v1.Install{
						Metadata: core.Metadata{
							Namespace: "supergloo-system",
							Name:      "test-consul-mesh",
						},
						Consul: &v1.ConsulInstall{
							Path:      "https://github.com/hashicorp/consul-helm/archive/v0.3.0.tar.gz",
							Namespace: namespace,
						},
						Encryption: &v1.Encryption{
							TlsEnabled: mtls,
						},
					},
				},
			},
		}
	}

	terminateNamespaceBlocking := func() {
		client := getKubeClient()
		client.CoreV1().Namespaces().Delete(namespace, &kubemeta.DeleteOptions{})
		Eventually(func() error {
			_, err := client.CoreV1().Namespaces().Get(namespace, kubemeta.GetOptions{})
			return err
		}, "60s", "1s").ShouldNot(BeNil()) // will be non-nil when NS is gone
	}

	waitForAvailablePods := func() {
		client := getKubeClient()
		Eventually(func() bool {
			podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
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

	AfterEach(func() {
		deleteWebhookConfigIfExists()
		deleteCrb()
		terminateNamespaceBlocking()
		// delete meshes
	})

	It("Can install consul with mtls enabled", func() {
		syncer := consul.ConsulInstallSyncer{
			Kube:       getKubeClient(),
			MeshClient: getMeshClient(getKubeConfig()),
		}
		snap := getSnapshot(true)
		err := syncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		waitForAvailablePods()
	})

	It("Can install consul without mtls enabled", func() {
		syncer := consul.ConsulInstallSyncer{
			Kube:       getKubeClient(),
			MeshClient: getMeshClient(getKubeConfig()),
		}
		snap := getSnapshot(false)
		err := syncer.Sync(context.TODO(), snap)
		Expect(err).NotTo(HaveOccurred())
		waitForAvailablePods()
	})

})
