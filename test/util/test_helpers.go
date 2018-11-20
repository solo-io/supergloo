package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/solo-io/supergloo/pkg/install/helm"

	"github.com/hashicorp/consul/api"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	gloo "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"k8s.io/client-go/kubernetes"

	kubecore "k8s.io/api/core/v1"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	helmlib "k8s.io/helm/pkg/helm"
	helmkube "k8s.io/helm/pkg/kube"

	security "github.com/openshift/client-go/security/clientset/versioned"
	client "k8s.io/apiextensions-apiserver/pkg/client/clientset/internalclientset/typed/apiextensions/internalversion"
)

var kubeConfig *rest.Config
var kubeClient *kubernetes.Clientset

var testKey = "-----BEGIN PRIVATE KEY-----\nMIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDBoI1sMdiOTvBBdjWlS\nZ8qwNuK9xV4yKuboLZ4Sx/OBfy1eKZocxTKvnjLrHUe139uhZANiAAQMTIR56O8U\nTIqf6uUHM4i9mZYLj152up7elS06Gi6lk7IeUQDHxP0NnOnbhC7rmtOV6myLNApL\nQ92kZKg7qa8q7OY/4w1QfC4ch7zZKxjNkSIiuAx7V/lzF6FYDcqT3js=\n-----END PRIVATE KEY-----"
var TestRoot = "-----BEGIN CERTIFICATE-----\nMIIB7jCCAXUCCQC2t6Lqc2xnXDAKBggqhkjOPQQDAjBhMQswCQYDVQQGEwJVUzEW\nMBQGA1UECAwNTWFzc2FjaHVzZXR0czESMBAGA1UEBwwJQ2FtYnJpZGdlMQwwCgYD\nVQQKDANPcmcxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xODExMTgxMzQz\nMDJaFw0xOTExMTgxMzQzMDJaMGExCzAJBgNVBAYTAlVTMRYwFAYDVQQIDA1NYXNz\nYWNodXNldHRzMRIwEAYDVQQHDAlDYW1icmlkZ2UxDDAKBgNVBAoMA09yZzEYMBYG\nA1UEAwwPd3d3LmV4YW1wbGUuY29tMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEDEyE\neejvFEyKn+rlBzOIvZmWC49edrqe3pUtOhoupZOyHlEAx8T9DZzp24Qu65rTleps\nizQKS0PdpGSoO6mvKuzmP+MNUHwuHIe82SsYzZEiIrgMe1f5cxehWA3Kk947MAoG\nCCqGSM49BAMCA2cAMGQCMCytVFc8sBdbM7DaBCz0N2ptdb0T7LFFfxDTzn4gjiDq\nVCd/3dct21TUWsthKXF2VgIwXEMI5EQiJ5kjR/y1KNBC9b4wfDiKRvG33jYe9gn6\ntzXUS00SoqG9D27/7aK71/xv\n-----END CERTIFICATE-----"
var testCertChain = ""

func GetKubeConfig() *rest.Config {
	if kubeConfig != nil {
		return kubeConfig
	}
	kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	Expect(err).NotTo(HaveOccurred())
	kubeConfig = cfg
	return cfg
}

func GetKubeClient() *kubernetes.Clientset {
	if kubeClient != nil {
		return kubeClient
	}
	cfg := GetKubeConfig()
	client, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	kubeClient = client
	return client
}

func GetSecurityClient() *security.Clientset {
	securityClient, err := security.NewForConfig(GetKubeConfig())
	Expect(err).To(BeNil())
	return securityClient
}

func GetSecretClient() gloo.SecretClient {
	kube := GetKubeClient()
	secretClient, err := gloo.NewSecretClient(&factory.KubeSecretClientFactory{
		Clientset: kube,
	})
	Expect(err).Should(BeNil())
	err = secretClient.Register()
	Expect(err).Should(BeNil())
	return secretClient
}

func TerminateNamespaceBlocking(namespace string) {
	client := GetKubeClient()
	client.CoreV1().Namespaces().Delete(namespace, &kubemeta.DeleteOptions{})
	Eventually(func() error {
		_, err := client.CoreV1().Namespaces().Get(namespace, kubemeta.GetOptions{})
		return err
	}, "120s", "1s").ShouldNot(BeNil()) // will be non-nil when NS is gone
}

func WaitForAvailablePods(namespace string) {
	client := GetKubeClient()
	Eventually(func() bool {
		podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
		Expect(err).To(BeNil())
		done := true
		for _, pod := range podList.Items {
			for _, condition := range pod.Status.Conditions {
				if pod.Status.Phase == kubecore.PodSucceeded {
					continue
				}
				if condition.Type == kubecore.PodReady && condition.Status != kubecore.ConditionTrue {
					done = false
				}
			}
		}
		return done
	}, "120s", "1s").Should(BeTrue())
}

func GetMeshClient(kubeCache *kube.KubeCache) v1.MeshClient {
	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         GetKubeConfig(),
		SharedCache: kubeCache,
	})
	Expect(err).Should(BeNil())
	err = meshClient.Register()
	Expect(err).Should(BeNil())
	return meshClient
}

func DeleteCrb(crbName string) {
	client := GetKubeClient()
	client.RbacV1().ClusterRoleBindings().Delete(crbName, &kubemeta.DeleteOptions{})
}

func DeleteWebhookConfigIfExists(webhookName string) {
	client := GetKubeClient()
	client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(webhookName, &kubemeta.DeleteOptions{})
}

func GetConsulServerPodName(namespace string) string {
	client := GetKubeClient()
	podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
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
func CreateConsulTunnel(namespace string, port int) (*helmkube.Tunnel, error) {
	podName := GetConsulServerPodName(namespace)
	t := helmkube.NewTunnel(GetKubeClient().CoreV1().RESTClient(), GetKubeConfig(), namespace, podName, port)
	return t, t.ForwardPort()
}

func CreateTestSecret(namespace string, name string) (*gloo.Secret, *core.ResourceRef) {
	tls := gloo.TlsSecret{
		RootCa:     TestRoot,
		PrivateKey: testKey,
		CertChain:  testCertChain,
	}
	tlsWrapper := gloo.Secret_Tls{
		Tls: &tls,
	}
	secret := &gloo.Secret{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      name,
		},
		Kind: &tlsWrapper,
	}
	GetSecretClient().Delete(namespace, name, clients.DeleteOpts{})
	_, err := GetSecretClient().Write(secret, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	ref := &core.ResourceRef{
		Namespace: namespace,
		Name:      name,
	}
	return secret, ref
}

func CheckCertMatches(consulTunnelPort int, rootCert string) {
	config := &api.Config{
		Address: fmt.Sprintf("127.0.0.1:%d", consulTunnelPort),
	}
	client, err := api.NewClient(config)
	Expect(err).NotTo(HaveOccurred())
	var queryOpts api.QueryOptions
	currentConfig, _, err := client.Connect().CAGetConfig(&queryOpts)
	Expect(err).NotTo(HaveOccurred())

	currentRoot := currentConfig.Config["RootCert"]
	Expect(currentRoot).To(BeEquivalentTo(rootCert))
}

func UninstallHelmRelease(releaseName string) error {
	// helm install
	helmClient, err := helm.GetHelmClient()
	if err != nil {
		return err
	}
	_, err = helmClient.DeleteRelease(releaseName, helmlib.DeletePurge(true))
	helm.Teardown()
	return err
}

func TryDeleteIstioCrds() {
	crdClient, err := client.NewForConfig(GetKubeConfig())
	if err != nil {
		return
	}
	crdList, err := crdClient.CustomResourceDefinitions().List(kubemeta.ListOptions{})
	if err != nil {
		return
	}
	for _, crd := range crdList.Items {
		//TODO: use labels
		if strings.Contains(crd.Name, "istio.io") {
			crdClient.CustomResourceDefinitions().Delete(crd.Name, &kubemeta.DeleteOptions{})
		}
	}
}
