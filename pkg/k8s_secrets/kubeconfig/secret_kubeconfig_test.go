package kubeconfig

import (
	"fmt"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("Secret KubeConfig Conversions", func() {
	var (
		clusterName1  = "test-name1"
		clusterName2  = "test-name2"
		kubeConfigRaw = `apiVersion: v1
clusters:
- cluster:
    server: test-server
  name: %s
contexts:
- context:
    cluster: test-name
    user: test-name
  name: test-name
current-context: test-name
kind: Config
preferences: {}
users:
- name: test-name
  user:
    token: alphanumericgarbage
`
		kubeConfigRaw1 = fmt.Sprintf(kubeConfigRaw, clusterName1)
		kubeConfigRaw2 = fmt.Sprintf(kubeConfigRaw, clusterName2)
		kubeConfig1    KubeConfig
		kubeConfig2    KubeConfig
	)

	BeforeEach(func() {
		config1, err := clientcmd.Load([]byte(kubeConfigRaw1))
		Expect(err).NotTo(HaveOccurred())
		config2, err := clientcmd.Load([]byte(kubeConfigRaw2))
		Expect(err).NotTo(HaveOccurred())
		kubeConfig1 = KubeConfig{
			Config:  *config1,
			Cluster: clusterName1,
		}
		kubeConfig2 = KubeConfig{
			Config:  *config2,
			Cluster: clusterName2,
		}
	})

	It("should convert a single KubeConfig to a single secret", func() {
		name := "secret-name"
		namespace := "secret-namespace"
		expectedSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    map[string]string{KubeConfigSecretLabel: "true"},
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				clusterName1: []byte(kubeConfigRaw1),
			},
			Type: v1.SecretTypeOpaque,
		}
		secret, err := KubeConfigToSecret(name, namespace, &kubeConfig1)
		Expect(err).NotTo(HaveOccurred())
		Expect(secret).To(Equal(expectedSecret))
	})

	It("should convert multiple KubeConfigs to a single secret", func() {
		name := "secret-name"
		namespace := "secret-namespace"
		expectedSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    map[string]string{KubeConfigSecretLabel: "true"},
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				clusterName1: []byte(kubeConfigRaw1),
				clusterName2: []byte(kubeConfigRaw2),
			},
			Type: v1.SecretTypeOpaque,
		}
		secret, err := KubeConfigsToSecret(name, namespace, []*KubeConfig{&kubeConfig1, &kubeConfig2})
		Expect(err).NotTo(HaveOccurred())
		Expect(secret).To(Equal(expectedSecret))
	})

	It("should fail if multiple clusters share the same name", func() {
		name := "secret-name"
		namespace := "secret-namespace"
		_, err := KubeConfigsToSecret(name, namespace, []*KubeConfig{&kubeConfig1, &kubeConfig1})
		Expect(err).To(testutils.HaveInErrorChain(DuplicateClusterName(clusterName1)))
	})
})
