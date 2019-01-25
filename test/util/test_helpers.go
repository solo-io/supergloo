package util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	factory2 "github.com/solo-io/supergloo/pkg/factory"

	"github.com/solo-io/supergloo/pkg/secret"

	"github.com/solo-io/supergloo/pkg/install/helm"

	"github.com/hashicorp/consul/api"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	gloo "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	istiosecret "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	"k8s.io/client-go/kubernetes"

	kubecore "k8s.io/api/core/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	helmkube "k8s.io/helm/pkg/kube"

	security "github.com/openshift/client-go/security/clientset/versioned"
	// love me google.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	istiov1 "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
)

var kubeConfig *rest.Config
var kubeClient *kubernetes.Clientset
var apiExtsClient apiexts.Interface
var upstreamClient gloo.UpstreamClient
var secretClient istiov1.IstioCacertsSecretClient

var testEcKey = "-----BEGIN PRIVATE KEY-----\nMIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDBoI1sMdiOTvBBdjWlS\nZ8qwNuK9xV4yKuboLZ4Sx/OBfy1eKZocxTKvnjLrHUe139uhZANiAAQMTIR56O8U\nTIqf6uUHM4i9mZYLj152up7elS06Gi6lk7IeUQDHxP0NnOnbhC7rmtOV6myLNApL\nQ92kZKg7qa8q7OY/4w1QfC4ch7zZKxjNkSIiuAx7V/lzF6FYDcqT3js=\n-----END PRIVATE KEY-----"
var TestEcRoot = "-----BEGIN CERTIFICATE-----\nMIIB7jCCAXUCCQC2t6Lqc2xnXDAKBggqhkjOPQQDAjBhMQswCQYDVQQGEwJVUzEW\nMBQGA1UECAwNTWFzc2FjaHVzZXR0czESMBAGA1UEBwwJQ2FtYnJpZGdlMQwwCgYD\nVQQKDANPcmcxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xODExMTgxMzQz\nMDJaFw0xOTExMTgxMzQzMDJaMGExCzAJBgNVBAYTAlVTMRYwFAYDVQQIDA1NYXNz\nYWNodXNldHRzMRIwEAYDVQQHDAlDYW1icmlkZ2UxDDAKBgNVBAoMA09yZzEYMBYG\nA1UEAwwPd3d3LmV4YW1wbGUuY29tMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEDEyE\neejvFEyKn+rlBzOIvZmWC49edrqe3pUtOhoupZOyHlEAx8T9DZzp24Qu65rTleps\nizQKS0PdpGSoO6mvKuzmP+MNUHwuHIe82SsYzZEiIrgMe1f5cxehWA3Kk947MAoG\nCCqGSM49BAMCA2cAMGQCMCytVFc8sBdbM7DaBCz0N2ptdb0T7LFFfxDTzn4gjiDq\nVCd/3dct21TUWsthKXF2VgIwXEMI5EQiJ5kjR/y1KNBC9b4wfDiKRvG33jYe9gn6\ntzXUS00SoqG9D27/7aK71/xv\n-----END CERTIFICATE-----"
var testCertChain = ""

var testRsaCaKey = `-----BEGIN PRIVATE KEY-----
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
2vvcOAJTiST+8EBuFnY2uRvE93JhsQ==
-----END PRIVATE KEY-----`

var testRsaCaCert = `-----BEGIN CERTIFICATE-----
MIIFPjCCAyYCCQDqnfUQmMxq3zANBgkqhkiG9w0BAQsFADBhMQswCQYDVQQGEwJV
UzEWMBQGA1UECAwNTWFzc2FjaHVzZXR0czESMBAGA1UEBwwJQ2FtYnJpZGdlMQww
CgYDVQQKDANPcmcxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xODExMjky
MDEwMzdaFw0xOTExMjkyMDEwMzdaMGExCzAJBgNVBAYTAlVTMRYwFAYDVQQIDA1N
YXNzYWNodXNldHRzMRIwEAYDVQQHDAlDYW1icmlkZ2UxDDAKBgNVBAoMA09yZzEY
MBYGA1UEAwwPd3d3LmV4YW1wbGUuY29tMIICIjANBgkqhkiG9w0BAQEFAAOCAg8A
MIICCgKCAgEA9CHyWDa2ABp9dFWfqCH5YD++q2karZvzxFSQ1rn3BMo2RlxbFcfn
2lydQrtoMwlE9jIxMkl1/5lyJdzhwf7pagz9ALJkct6txdjMNN/lYvD1mU645NJy
ASjJvki67HiV6D/+yDXZtn9ZAyFc/QcICkLRPN/LYO1xwiIyMD2s3RTDYO8SSeRU
H0ZpmeJONt2v3wSQfvW225tM/G066nT6P9tALxC2DemmDP4GPd8wMDDjoFGVpVh/
DjqxHAuWiLYiVaaDz1J2XPZ5A09SLnPM0TuIZb/0bvVEyyGzfD3Bhdk1FtdXYtAD
6noREGjMv0zvBji+P38Ngbz8meq9H/WkJkL67qVDjgeqhedC4MSGdM3Qom6uJUpu
S4S9PhP+87Juhb1vouEZpwu4od1RX8maC8rGhb4JmhOv3mEI1EeqMPOU+1vOaZPB
w9gyLmog4v4luNVXvu5JS5VeoCg063wjf97nNIMizwwo+Js5sHDuqa1R1BvSBMFF
F0cXb6G4r0yw9tz56k0W+93bhbYVyqqoBxMR49O9Ylk7FbYA1g5v+ZYVg8hQIMXk
O+y03aZdD4XR3wif15QyqOlv8gZlS/HOAfBv7fpAAh+M+MLffliZroi7e/Q3HX7D
kgSq8Va6KANh6gW9kSsMAoWIya+PqLukZSOH1Ed4BJid4FtuEFjm5ZMCAwEAATAN
BgkqhkiG9w0BAQsFAAOCAgEAA0CEYPyDjZZTE58VXhqahno8Vi/UP6+X+ThCysAk
w/DX9OfROskEpK369nOecM5qsm2VJNFJL2Qfe5HhzxhjidOCU0zixSAUP+3VSSv4
3mNDcJQPtTZrxLPUfvGGH4A2xWQCbT8QuTUt/bT+ILrLccltdFg+LpJK5BqVl4vM
Ukdmth0I69eU5B0dpQDk1IWXzgMZaiDuBaYX1zzN8GYOkktbl9KUusbYbMNJujAi
/7Z95RRPTyRAEtXmvle+TUWJl8VfE4UMOOnRvzgxg0RuFMxr8zVAvLiAedaGavyI
2yM/YT8icIjNFF7v92/ekATobXWVROebcnBSVszG/TN7zV8yZ3Hhnv4/tLUy2hmk
Q1UafIYwpIR1CLMf6u60ttr5Lyqxbcep5JfJyJDCwgAqPrdLeYx90WVdaTSwLraD
DJDgivAQuVcPfK4GthXVkz4kf4i+2CG8O6ApUJoR5z9W2q5eaHhqn/2S0himwV/7
q6/S4htUsT0dDGIvta2fIswGZENRpVlIjTzIPVk/m7OaaI26t/unpo+4aJRUG+cP
jlrm+5MdUBwLenQscf/98tFHg7hb9jkfBLcMDDNTXB2X3Y4iyE531iqm4zPHU2gB
17kPBdvecKcRSlqDm6JLqjvKHF7HGqVTVZ4rPhJQKDns+68f0/rN89rqPzAeq0he
2Jk=
-----END CERTIFICATE-----`

var testRsaCertChain = ""
var testRsaRootCert = testRsaCaCert

func GetKubeConfig() *rest.Config {
	if kubeConfig != nil {
		return kubeConfig
	}
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	kubeConfig = cfg
	return cfg
}

func GetKubeClient() *kubernetes.Clientset {
	if kubeClient != nil {
		return kubeClient
	}
	cfg := GetKubeConfig()
	client, err := kubernetes.NewForConfig(cfg)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	kubeClient = client
	return client
}

func GetApiExtsClient() apiexts.Interface {
	if apiExtsClient != nil {
		return apiExtsClient
	}
	cfg := GetKubeConfig()
	client, err := apiexts.NewForConfig(cfg)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	apiExtsClient = client
	return client
}

func GetSecurityClient() *security.Clientset {
	securityClient, err := security.NewForConfig(GetKubeConfig())
	ExpectWithOffset(1, err).To(BeNil())
	return securityClient
}

func GetSecretClient() istiosecret.IstioCacertsSecretClient {
	if secretClient != nil {
		return secretClient
	}
	client, err := factory2.GetIstioCacertsSecretClient(GetKubeClient())
	ExpectWithOffset(1, err).Should(BeNil())
	err = client.Register()
	ExpectWithOffset(1, err).Should(BeNil())
	secretClient = client
	return secretClient
}

func TryCreateNamespace(namespace string) bool {
	client := GetKubeClient()
	resource := &kubecore.Namespace{
		ObjectMeta: kubemeta.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(resource)
	if err != nil {
		ExpectWithOffset(1, apierrors.IsAlreadyExists(err)).To(BeTrue())
		return false
	}
	return true
}

func TerminateNamespace(namespace string) {
	client := GetKubeClient()
	gracePeriod := int64(0)
	deleteOptions := &kubemeta.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}
	client.CoreV1().Pods(namespace).DeleteCollection(deleteOptions, kubemeta.ListOptions{})
	client.CoreV1().Namespaces().Delete(namespace, deleteOptions)
}

func TerminateNamespaceBlocking(namespace string) {
	TerminateNamespace(namespace)
	WaitForTerminatedNamespace(namespace)
}

func WaitForTerminatedNamespace(namespace string) {
	client := GetKubeClient()
	EventuallyWithOffset(1, func() error {
		_, err := client.CoreV1().Namespaces().Get(namespace, kubemeta.GetOptions{})
		return err
	}, "120s", "1s").ShouldNot(BeNil()) // will be non-nil when NS is gone
}

func WaitForAvailablePodsWithTimeout(namespace string, timeout string) int {
	// use helper function so that stack offset is consistent
	return waitForAvailablePodsWithTimeout(namespace, timeout)
}

func waitForAvailablePodsWithTimeout(namespace, timeout string) int {
	client := GetKubeClient()
	var podNum int

	EventuallyWithOffset(2, func() (bool, error) {
		podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
		if err != nil {
			return false, err
		}
		podNum = len(podList.Items)
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
		return done, nil
	}, timeout, "1s").Should(BeTrue())
	return podNum
}

func GetPodWithSubstringInName(namespace string, substring string) (*kubecore.Pod, error) {
	podList, err := GetKubeClient().CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, substring) {
			return pod.DeepCopy(), nil
		}
	}
	Expect(false).To(BeTrue())
	return nil, errors.Errorf("Couldn't find pod with name containing %s in namespace %s.\n", substring, namespace)
}

func WaitForDeletedPodsWithTimeout(namespace string, timeout string) {
	client := GetKubeClient()
	Eventually(func() bool {
		podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
		Expect(err).To(BeNil())
		return len(podList.Items) == 0
	}, timeout, "1s").Should(BeTrue())
}

func WaitForAvailablePods(namespace string) int {
	return waitForAvailablePodsWithTimeout(namespace, "120s")
}

func WaitForDeletedPods(namespace string) {
	WaitForDeletedPodsWithTimeout(namespace, "120s")
}

func GetMeshClient() v1.MeshClient {
	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         GetKubeConfig(),
		SharedCache: kube.NewKubeCache(),
	})
	ExpectWithOffset(1, err).Should(BeNil())
	err = meshClient.Register()
	ExpectWithOffset(1, err).Should(BeNil())
	return meshClient
}

func GetUpstreamClient() gloo.UpstreamClient {
	client, err := gloo.NewUpstreamClient(&factory.KubeResourceClientFactory{
		Crd:         gloo.UpstreamCrd,
		Cfg:         GetKubeConfig(),
		SharedCache: kube.NewKubeCache(),
	})
	ExpectWithOffset(1, err).Should(BeNil())
	err = client.Register()
	ExpectWithOffset(1, err).Should(BeNil())
	return client
}

func DeleteCrb(crbName string) {
	client := GetKubeClient()
	client.RbacV1().ClusterRoleBindings().Delete(crbName, &kubemeta.DeleteOptions{})
}

func CrbDoesntExist(crbName string) bool {
	client := GetKubeClient()
	_, err := client.RbacV1().ClusterRoleBindings().Get(crbName, kubemeta.GetOptions{})
	return apierrors.IsNotFound(err)
}

func DeleteWebhookConfigIfExists(webhookName string) {
	client := GetKubeClient()
	hooks, err := client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().List(kubemeta.ListOptions{})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	for _, hook := range hooks.Items {
		if strings.HasSuffix(hook.Name, webhookName) {
			client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(hook.Name, &kubemeta.DeleteOptions{})
		}
	}
}

func WebhookConfigNotFound(webhookName string) bool {
	client := GetKubeClient()
	_, err := client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(webhookName, kubemeta.GetOptions{})
	return apierrors.IsNotFound(err)
}

func GetConsulServerPodName(namespace string) string {
	consulServerPodSuffix := "consul-mesh-server-0"
	client := GetKubeClient()
	var podName string
	EventuallyWithOffset(1, func() (string, error) {
		podList, err := client.CoreV1().Pods(namespace).List(kubemeta.ListOptions{})
		if err != nil {
			return "", err
		}
		for _, pod := range podList.Items {
			if strings.HasSuffix(pod.Name, consulServerPodSuffix) {
				podName = pod.Name
				return podName, nil
			}
		}
		return "", errors.Errorf("pod with name %v not found", consulServerPodSuffix)
	}, time.Minute).Should(HaveSuffix(consulServerPodSuffix))
	return podName
}

// New creates a new and initialized tunnel.
func CreateConsulTunnel(namespace string, port int) (*helmkube.Tunnel, error) {
	podName := GetConsulServerPodName(namespace)
	t := helmkube.NewTunnel(GetKubeClient().CoreV1().RESTClient(), GetKubeConfig(), namespace, podName, port)
	return t, t.ForwardPort()
}

func createSecret(secret *istiosecret.IstioCacertsSecret, namespace string, name string) (*istiosecret.IstioCacertsSecret, *core.ResourceRef) {
	GetSecretClient().Delete(namespace, name, clients.DeleteOpts{})
	_, err := GetSecretClient().Write(secret, clients.WriteOpts{})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ref := &core.ResourceRef{
		Namespace: namespace,
		Name:      name,
	}
	return secret, ref
}

func CreateTestEcSecret(namespace string, name string) (*istiosecret.IstioCacertsSecret, *core.ResourceRef) {
	secret := &istiosecret.IstioCacertsSecret{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      name,
		},
		CaCert:    TestEcRoot,
		CaKey:     testEcKey,
		RootCert:  TestEcRoot,
		CertChain: testCertChain,
	}
	return createSecret(secret, namespace, name)
}

func CreateTestRsaSecret(namespace string, name string) (*istiosecret.IstioCacertsSecret, *core.ResourceRef) {
	secret := &istiosecret.IstioCacertsSecret{
		Metadata: core.Metadata{
			Namespace: namespace,
			Name:      name,
		},
		CaCert:    testRsaCaCert,
		CaKey:     testRsaCaKey,
		RootCert:  testRsaRootCert,
		CertChain: testRsaCertChain,
	}
	return createSecret(secret, namespace, name)
}

func CheckCertMatchesConsul(consulTunnelPort int, rootCert string) {
	config := &api.Config{
		Address: fmt.Sprintf("127.0.0.1:%d", consulTunnelPort),
	}
	client, err := api.NewClient(config)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	var queryOpts api.QueryOptions
	currentConfig, _, err := client.Connect().CAGetConfig(&queryOpts)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	currentRoot := currentConfig.Config["RootCert"]
	ExpectWithOffset(1, currentRoot).To(BeEquivalentTo(rootCert))
}

func CheckCertMatchesIstio(installNamespace string) {
	actual, err := GetSecretClient().Read(installNamespace, secret.CustomRootCertificateSecretName, clients.ReadOpts{})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, actual.RootCert).Should(BeEquivalentTo(testRsaRootCert))
	ExpectWithOffset(1, actual.CaCert).Should(BeEquivalentTo(testRsaCaCert))
	ExpectWithOffset(1, actual.CaKey).Should(BeEquivalentTo(testRsaCaKey))
	ExpectWithOffset(1, actual.CertChain).Should(BeEquivalentTo(testRsaCertChain))
}

func UninstallHelmRelease(releaseName string) error {
	helmWrapper := helm.KubeHelmClient{}
	return helmWrapper.DeleteHelmRelease(context.TODO(), releaseName)
}

func HelmReleaseDoesntExist(releaseName string) bool {
	helmWrapper := helm.KubeHelmClient{}
	return helmWrapper.HelmReleaseDoesntExist(releaseName)
}

func TryDeleteIstioCrds() {
	crdClient := GetApiExtsClient()
	crdList, err := crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(kubemeta.ListOptions{})
	if err != nil {
		return
	}
	for _, crd := range crdList.Items {
		//TODO: use labels
		if strings.Contains(crd.Name, "istio.io") {
			crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.Name, &kubemeta.DeleteOptions{})
		}
	}
}

func GetUpstreamNames(upstreamClient gloo.UpstreamClient, namespace string) ([]string, error) {
	ul, err := upstreamClient.List(namespace, clients.ListOpts{})
	if err != nil {
		return nil, err
	}
	ups := []string{}
	for _, up := range ul {
		ups = append(ups, up.Metadata.Name)
	}
	return ups, nil
}
