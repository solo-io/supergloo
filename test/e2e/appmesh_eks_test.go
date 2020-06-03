package e2e

import (
	"context"
	"fmt"
	"os"

	zephyr_core "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	v1 "k8s.io/api/core/v1"
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
	settingsYaml = fmt.Sprintf(`
apiVersion: core.zephyr.solo.io/v1alpha1
kind: Settings
metadata:
  namespace: service-mesh-hub
  name: settings
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
`, AwsAccountId, AppmeshArn, EksArn)

	buildSecretYaml = func(awsAccessKeyId, awsSecretAccessKey string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Secret
type: solo.io/register/aws-credentials
metadata:
  name: e2e-appmesh-eks
  namespace: service-mesh-hub
data:
  aws_access_key_id: %s
  aws_secret_access_key: %s
`, awsAccessKeyId, awsSecretAccessKey)
	}

	kubeClusterName = metadata.BuildEksClusterName(EksClusterName, Region)
)

var _ = Describe("Appmesh EKS ", func() {
	AfterEach(func() {
		testLabels := map[string]string{"test": "true"}
		opts := &client.DeleteAllOfOptions{}
		opts.LabelSelector = labels.SelectorFromSet(testLabels)
		opts.Namespace = "service-mesh-hub"
		env.Management.TrafficPolicyClient.DeleteAllOfTrafficPolicy(context.Background(), opts)
	})

	/*
	 CLEANUP
		1. AWS creds
		2. settings CRD
	 */

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
		existingSettings, err := env.Management.SettingsClient.GetSettings(context.Background(), client.ObjectKey{Name: metadata.GlobalSettingsName, Namespace: SmhNamespace})
		Expect(err).NotTo(HaveOccurred())
		existingSettings.Spec = newSettings.Spec
		err = env.Management.SettingsClient.UpdateSettings(context.Background(), existingSettings)
		Expect(err).NotTo(HaveOccurred())
	}

	var expectGetKubeCluster = func(name string) {
		Eventually(
			KubeClusterShouldExist(client.ObjectKey{Name: name, Namespace: SmhNamespace}, env.Management),
			"1m", "1s").
			ShouldNot(BeNil())
	}

	FIt("should discover Appmesh mesh and EKS cluster", func() {
		// Register AWS account credentials
		registerAwsSecret()
		// Discover Appmesh mesh and EKS cluster
		applySettings(settingsYaml)
		expectGetKubeCluster(kubeClusterName)
	})
})
