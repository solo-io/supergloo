package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aokoli/goutils"
	"github.com/golang/sync/errgroup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/metadata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EksKubeContextName    = "smh-e2e-test"
	AwsAccountId          = "410461945957"
	Region                = "us-east-2"
	AppmeshArn            = "arn:aws:appmesh:us-east-2:410461945957:mesh/smh-e2e-test"
	EksArn                = "arn:aws:eks:us-east-2:410461945957:cluster/smh-e2e-test"
	AppmeshName           = "smh-e2e-test"
	EksClusterName        = "smh-e2e-test"
	SmhNamespace          = "service-mesh-hub"
	AppmeshInjectionLabel = "appmesh.k8s.aws/sidecarInjectorWebhook=enabled"
	ClusterLockTimeout    = 20 * time.Minute
)

var (
	settingsObjKey    = client.ObjectKey{Name: metadata.GlobalSettingsName, Namespace: SmhNamespace}
	secretObjKey      = client.ObjectKey{Name: "e2e-appmesh-eks", Namespace: SmhNamespace}
	kubeClusterObjKey = client.ObjectKey{Name: metadata.BuildEksKubernetesClusterName(EksClusterName, Region), Namespace: SmhNamespace}
	appmeshObjKey     = client.ObjectKey{Name: metadata.BuildAppMeshName(AppmeshName, Region, AwsAccountId), Namespace: SmhNamespace}
	virtualMeshObjKey = client.ObjectKey{Name: "appmesh-vm", Namespace: SmhNamespace}

	// Populated during setup
	generatedNamespace string
	eksKubeContext     KubeContext

	defaultSettingsYaml = fmt.Sprintf(`
apiVersion: core.smh.solo.io/smh_networking
kind: Settings
metadata:
  name: %s
  namespace: %s
spec:
  aws:
    disabled: true
`, settingsObjKey.Name, settingsObjKey.Namespace)

	settingsYaml = fmt.Sprintf(`
apiVersion: core.smh.solo.io/smh_networking
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
	virtualMeshYaml = fmt.Sprintf(`
apiVersion: networking.smh.solo.io/smh_networking
kind: VirtualMesh
metadata:
  name: %s
  namespace: %s
spec:
  displayName: "Appmesh-VM"
  certificateAuthority:
    builtin:
      ttlDays: 356
      rsaKeySizeBytes: 4096
      orgName: "service-mesh-hub"
  federation: 
    mode: PERMISSIVE
  shared: {}
  enforceAccessControl: ENABLED
  meshes:
  - name: %s
    namespace: %s
`, virtualMeshObjKey.Name, virtualMeshObjKey.Namespace, appmeshObjKey.Name, appmeshObjKey.Namespace)

	buildSecretYaml = func(awsAccessKeyId, awsSecretAccessKey string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Secret
type: solo.io/register/aws-credentials
metadata:
  name: %s
  namespace: %s
stringData:
  aws_access_key_id: %s
  aws_secret_access_key: %s
`, secretObjKey.Name, secretObjKey.Namespace, awsAccessKeyId, awsSecretAccessKey)
	}

	kubeClusterName = metadata.BuildEksKubernetesClusterName(EksClusterName, Region)
)

func getEksKubeContext(ctx context.Context) KubeContext {
	eg, ctx := errgroup.WithContext(ctx)

	cmd := exec.CommandContext(ctx,
		"aws", "eks",
		"--region", Region,
		"update-kubeconfig",
		"--name", EksClusterName,
		"--alias", EksKubeContextName)
	cmd.Dir = "../.."
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	err := cmd.Start()
	// close this end after start, as we dont need it.

	eg.Go(cmd.Wait)

	err = eg.Wait()
	if err != nil {
		dumpState()
	}

	Expect(err).NotTo(HaveOccurred())

	// Use current context
	return NewKubeContext(EksKubeContextName)
}

func setupAppmeshEksEnvironment() string {
	// Deploy bookinfo into a new namespace on the EKS cluster
	ctx := context.Background()
	eksKubeContext = getEksKubeContext(ctx)
	config, err := eksKubeContext.Config.ClientConfig()
	Expect(err).ToNot(HaveOccurred())
	WaitForClusterLock(ctx, config, ClusterLockTimeout)

	randomString, err := goutils.RandomAlphabetic(4)
	Expect(err).ToNot(HaveOccurred())
	generatedNamespace = strings.ToLower(randomString)
	eksKubeContext.CreateNamespace(ctx, generatedNamespace)
	eksKubeContext.LabelNamespace(ctx, generatedNamespace, AppmeshInjectionLabel)
	eksKubeContext.DeployBookInfo(ctx, generatedNamespace)
	// Set SERVICES_DOMAIN env var on productpage-v1 to point at services within generated namespace
	// https://github.com/istio/istio/blob/1.5.0/samples/bookinfo/src/productpage/productpage.py#L60
	eksKubeContext.SetDeploymentEnvVars(ctx, generatedNamespace, "productpage-v1", "productpage", map[string]string{"SERVICES_DOMAIN": generatedNamespace})
	eksKubeContext.SetDeploymentEnvVars(ctx, generatedNamespace, "reviews-v1", "reviews", map[string]string{"SERVICES_DOMAIN": generatedNamespace})
	eksKubeContext.SetDeploymentEnvVars(ctx, generatedNamespace, "reviews-v2", "reviews", map[string]string{"SERVICES_DOMAIN": generatedNamespace})
	eksKubeContext.SetDeploymentEnvVars(ctx, generatedNamespace, "reviews-v3", "reviews", map[string]string{"SERVICES_DOMAIN": generatedNamespace})
	return generatedNamespace
}

func cleanupAppmeshEksEnvironment(ns string) {
	// Only clean up management-cluster resources if we're running multiple iterations of this test against the same management-cluster.
	// Otherwise the cluster will be torn down anyways, so no need to clean anything up.
	if useExisting := os.Getenv("USE_EXISTING"); useExisting != "" {
		if env.Management.SettingsClient == nil {
			// this can happen in early failure
			return
		}
		// Cleans up discovery resources on management cluster
		// Reset back to default settings. This must be done before removing the AWS secret.
		settings, err := env.Management.SettingsClient.GetSettings(context.Background(), settingsObjKey)
		Expect(err).NotTo(HaveOccurred())
		var defaultSettings smh_core.Settings
		ParseYaml(defaultSettingsYaml, &defaultSettings)
		settings.Spec = defaultSettings.Spec
		err = env.Management.SettingsClient.UpdateSettings(context.Background(), settings)
		Expect(err).NotTo(HaveOccurred())
		// Wait for mesh-discovery to clean up discovered resources
		time.Sleep(10 * time.Second)
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

	// Clean up resources on remote EKS cluster
	ctx := context.Background()
	eksKubeContext.DeleteNamespace(ctx, ns)
}

var _ = Describe("Appmesh EKS ", func() {
	BeforeEach(func() {
		if !RunEKS() {
			Skip("skipping EKS tests")
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

	var applySettings = func() {
		var newSettings smh_core.Settings
		ParseYaml(settingsYaml, &newSettings)
		existingSettings, err := env.Management.SettingsClient.GetSettings(context.Background(), settingsObjKey)
		Expect(err).NotTo(HaveOccurred())
		if !existingSettings.Spec.Equal(newSettings.Spec) {
			existingSettings.Spec = newSettings.Spec
			err = env.Management.SettingsClient.UpdateSettings(context.Background(), existingSettings)
			Expect(err).NotTo(HaveOccurred())
		}
	}

	var applyVirtualMesh = func() {
		var virtualMesh smh_networking.VirtualMesh
		ParseYaml(virtualMeshYaml, &virtualMesh)
		err := env.Management.VirtualMeshClient.CreateVirtualMesh(context.Background(), &virtualMesh)
		Expect(err).NotTo(HaveOccurred())
	}

	var deleteVirtualMesh = func() {
		err := env.Management.VirtualMeshClient.DeleteVirtualMesh(context.Background(), virtualMeshObjKey)
		// Give SMH time to make API calls to delete Appmesh resources
		time.Sleep(5 * time.Second)
		Expect(err).NotTo(HaveOccurred())
	}

	var expectGetKubeCluster = func(name string) {
		Eventually(
			KubeCluster(client.ObjectKey{Name: name, Namespace: SmhNamespace}, env.Management),
			"30s", "1s").
			ShouldNot(BeNil())
	}

	var expectGetMesh = func(name string) {
		Eventually(
			Mesh(client.ObjectKey{Name: name, Namespace: SmhNamespace}, env.Management),
			"60s", "1s").
			ShouldNot(BeNil())
	}

	var curlReviewsWithExpectedOutput = func(expectedString string, shouldExpect bool) {
		productPageDeployment := "productpage-v1"
		ctx := context.Background()
		eksKubeContext.WaitForRollout(ctx, generatedNamespace, productPageDeployment)
		eventuallyCurl := Eventually(func() string {
			return eksKubeContext.Curl(
				context.Background(),
				generatedNamespace,
				productPageDeployment,
				"curl",
				fmt.Sprintf("http://reviews.%s:9080/reviews/1", generatedNamespace))
		}, "120s", "1s")
		if shouldExpect {
			eventuallyCurl.Should(ContainSubstring(expectedString))
		} else {
			eventuallyCurl.ShouldNot(ContainSubstring(expectedString))
		}
	}

	It("should discover Appmesh mesh and EKS cluster", func() {
		// Register AWS account credentials
		registerAwsSecret()
		// Discover Appmesh mesh and EKS cluster
		applySettings()
		expectGetKubeCluster(kubeClusterName)
		expectGetMesh("appmesh-smh-e2e-test-us-east-2-410461945957")
	})

	It("should translate Appmesh resources to enable all communication between workloads and services", func() {
		applyVirtualMesh()
		curlReviewsWithExpectedOutput("The slapstick humour is refreshing", true)
	})

	It("should cleanup translated Appmesh resources and disable communication between workloads and services", func() {
		deleteVirtualMesh()
		curlReviewsWithExpectedOutput("The slapstick humour is refreshing", false)
	})
})
