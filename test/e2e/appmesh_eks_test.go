package e2e

import (
	"context"
	"fmt"
	"os"
	"time"

	zephyr_core "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	AwsAccountId   = "410461945957"
	Region         = "us-east-2"
	AppmeshArn     = "arn:aws:appmesh:us-east-2:410461945957:mesh/smh-e2e-test"
	EksArn         = "arn:aws:eks:us-east-2:410461945957:cluster/smh-e2e-test"
	EksClusterName = "smh-e2e-test"
	SmhNamespace   = "service-mesh-hub"
)

var (
	settingsObjKey    = client.ObjectKey{Name: metadata.GlobalSettingsName, Namespace: SmhNamespace}
	secretObjKey      = client.ObjectKey{Name: "e2e-appmesh-eks", Namespace: SmhNamespace}
	kubeClusterObjKey = client.ObjectKey{Name: metadata.BuildEksKubernetesClusterName(EksClusterName, Region), Namespace: SmhNamespace}

	defaultSettingsYaml = fmt.Sprintf(`
apiVersion: core.zephyr.solo.io/v1alpha1
kind: Settings
metadata:
  name: %s
  namespace: %s
spec:
  aws:
    disabled: true
`, settingsObjKey.Name, settingsObjKey.Namespace)

	settingsYaml = fmt.Sprintf(`
apiVersion: core.zephyr.solo.io/v1alpha1
kind: Settings
metadata:
  name: %s
  namespace: %s
spec:
  aws:
    disabled: false
    accounts:
      - accountId: "%s"
        meshDiscovery:
          resourceSelectors:
          - arn: %s
        eksDiscovery:
          resourceSelectors:
          - arn: %s
`, settingsObjKey.Name, settingsObjKey.Namespace, AwsAccountId, AppmeshArn, EksArn)

	buildSecretYaml = func(awsAccessKeyId, awsSecretAccessKey string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Secret
type: solo.io/register/aws-credentials
metadata:
  name: %s
  namespace: %s
data:
  aws_access_key_id: %s
  aws_secret_access_key: %s
`, secretObjKey.Name, secretObjKey.Namespace, awsAccessKeyId, awsSecretAccessKey)
	}

	kubeClusterName = metadata.BuildEksKubernetesClusterName(EksClusterName, Region)
)

var _ = Describe("Appmesh EKS ", func() {
	// Cleans up discovery resources
	var cleanup = func() {
		// Reset back to default settings. This must be done before removing the AWS secret.
		settings, err := env.Management.SettingsClient.GetSettings(context.Background(), settingsObjKey)
		Expect(err).NotTo(HaveOccurred())
		var defaultSettings zephyr_core.Settings
		ParseYaml(defaultSettingsYaml, &defaultSettings)
		settings.Spec = defaultSettings.Spec
		err = env.Management.SettingsClient.UpdateSettings(context.Background(), settings)
		Expect(err).NotTo(HaveOccurred())

		// Wait for mesh-discovery to clean up discovered resources
		time.Sleep(2 * time.Second)

		// Delete AWS credentials
		err = env.Management.SecretClient.DeleteSecret(context.Background(), secretObjKey)
		if errors.IsNotFound(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())
		// Delete remote cluster k8s credentials
		err = env.Management.SecretClient.DeleteSecret(context.Background(), kubeClusterObjKey)
		if errors.IsNotFound(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())

		// Delete KubernetesCluster
		err = env.Management.KubeClusterClient.DeleteKubernetesCluster(context.Background(), kubeClusterObjKey)
		if errors.IsNotFound(err) {
			err = nil
		}
		Expect(err).NotTo(HaveOccurred())
	}

	AfterEach(func() {
		testLabels := map[string]string{"test": "true"}
		opts := &client.DeleteAllOfOptions{}
		opts.LabelSelector = labels.SelectorFromSet(testLabels)
		opts.Namespace = "service-mesh-hub"
		if useExisting := os.Getenv("USE_EXISTING"); useExisting != "" {
			// Reset SMH state for subsequent tests
			cleanup()
		}
	})

	// Fetch base64 encoded AWS credentials from environment
	var registerAwsSecret = func() {
		awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
		awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		var secret v1.Secret
		secretYaml := buildSecretYaml(awsAccessKeyId, awsSecretAccessKey)
		ParseYaml(secretYaml, &secret)
		err := env.Management.SecretClient.CreateSecret(context.Background(), &secret)
		Expect(err).NotTo(HaveOccurred())
	}

	var applySettings = func(settingsYaml string) {
		var newSettings zephyr_core.Settings
		ParseYaml(settingsYaml, &newSettings)
		existingSettings, err := env.Management.SettingsClient.GetSettings(context.Background(), settingsObjKey)
		Expect(err).NotTo(HaveOccurred())
		if !existingSettings.Spec.Equal(newSettings.Spec) {
			existingSettings.Spec = newSettings.Spec
			err = env.Management.SettingsClient.UpdateSettings(context.Background(), existingSettings)
			Expect(err).NotTo(HaveOccurred())
		}
	}

	var expectGetKubeCluster = func(name string) {
		Eventually(
			KubeClusterShouldExist(client.ObjectKey{Name: name, Namespace: SmhNamespace}, env.Management),
			"30s", "1s").
			ShouldNot(BeNil())
	}

	It("should discover Appmesh mesh and EKS cluster", func() {
		// Register AWS account credentials
		registerAwsSecret()
		// Discover Appmesh mesh and EKS cluster
		applySettings(settingsYaml)
		expectGetKubeCluster(kubeClusterName)
	})
})
