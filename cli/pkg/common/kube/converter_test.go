package kube_test

import (
	"encoding/base64"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_files "github.com/solo-io/service-mesh-hub/cli/pkg/common/files/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("Converter", func() {
	var (
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should convert a single KubeConfig to a single secret", func() {
		name := "secret-name"
		namespace := "secret-namespace"
		clusterName := "test-cluster-name"
		contextName := "test-context-name"
		caData := base64.StdEncoding.EncodeToString([]byte("test-ca-data"))
		kubeConfigRaw := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ` + caData + `
    server: test-server
  name: ` + clusterName + `
contexts:
- context:
    cluster: ` + clusterName + `
    user: test-user
  name: ` + contextName + `
current-context: ` + contextName + `
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: alphanumericgarbage
`

		expectedSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    map[string]string{kube.KubeConfigSecretLabel: "true"},
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				clusterName: []byte(kubeConfigRaw),
			},
			Type: v1.SecretTypeOpaque,
		}

		config, err := clientcmd.Load([]byte(kubeConfigRaw))
		Expect(err).NotTo(HaveOccurred())
		secret, err := kube.NewConverter(nil).ConfigToSecret(name, namespace, &kube.KubeConfig{
			Config:  *config,
			Cluster: clusterName,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(secret).To(Equal(expectedSecret))
	})

	It("can read a CA file when provided", func() {
		fileReader := mock_files.NewMockFileReader(ctrl)
		caBytes := []byte("BEGIN TEST CA CONTENT END CA CONTENT")
		caFileName := "/tmp/does/not/exist/test-ca.pem"

		fileReader.EXPECT().
			Read(caFileName).
			Return(caBytes, nil)

		name := "secret-name"
		namespace := "secret-namespace"
		clusterName := "test-cluster-name"
		contextName := "test-context-name"

		kubeConfigRaw := `apiVersion: v1
clusters:
- cluster:
    certificate-authority: ` + caFileName + `
    server: test-server
  name: ` + clusterName + `
contexts:
- context:
    cluster: ` + clusterName + `
    user: test-user
  name: ` + contextName + `
current-context: ` + contextName + `
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: alphanumericgarbage
`

		expectedRawKubeConfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ` + base64.StdEncoding.EncodeToString(caBytes) + `
    server: test-server
  name: ` + clusterName + `
contexts:
- context:
    cluster: ` + clusterName + `
    user: test-user
  name: ` + contextName + `
current-context: ` + contextName + `
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: alphanumericgarbage
`

		expectedSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    map[string]string{kube.KubeConfigSecretLabel: "true"},
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				clusterName: []byte(expectedRawKubeConfig),
			},
			Type: v1.SecretTypeOpaque,
		}

		config, err := clientcmd.Load([]byte(kubeConfigRaw))
		Expect(err).NotTo(HaveOccurred())
		secret, err := kube.NewConverter(fileReader).ConfigToSecret(name, namespace, &kube.KubeConfig{
			Config:  *config,
			Cluster: clusterName,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(secret).To(Equal(expectedSecret))
	})
})
