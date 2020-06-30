package e2e

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	istio_networking "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	smh_core "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/test/e2e/kubectl"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Env struct {
	Management KubeContext
	Remote     KubeContext
}

func (e Env) DumpState() {
	dumpState()
}

func newEnv(mgmt, remote string) Env {
	return Env{
		Management: NewKubeContext(mgmt),
		Remote:     NewKubeContext(remote),
	}
}

type KubeContext struct {
	Context               string
	Config                clientcmd.ClientConfig
	Clientset             *kubernetes.Clientset
	TrafficPolicyClient   smh_networking.TrafficPolicyClient
	KubeClusterClient     smh_discovery.KubernetesClusterClient
	MeshClient            smh_discovery.MeshClient
	SettingsClient        smh_core.SettingsClient
	SecretClient          kubernetes_core.SecretClient
	VirtualMeshClient     smh_networking.VirtualMeshClient
	FailoverServiceClient smh_networking.FailoverServiceClient
	VirtualServiceClient  v1alpha3.VirtualServiceClient
}

// If kubecontext is empty string, use current context.
func NewKubeContext(kubecontext string) KubeContext {
	cfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	config := clientcmd.NewNonInteractiveClientConfig(*cfg, kubecontext, &clientcmd.ConfigOverrides{}, nil)
	restcfg, err := config.ClientConfig()
	Expect(err).NotTo(HaveOccurred())

	clientset, err := kubernetes.NewForConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	kubeCoreClientset, err := kubernetes_core.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	networkingClientset, err := smh_networking.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	discoveryClientset, err := smh_discovery.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	coreClientset, err := smh_core.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	istioNetworkingClientset, err := istio_networking.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	return KubeContext{
		Context:               kubecontext,
		Config:                config,
		Clientset:             clientset,
		TrafficPolicyClient:   networkingClientset.TrafficPolicies(),
		VirtualMeshClient:     networkingClientset.VirtualMeshes(),
		MeshClient:            discoveryClientset.Meshes(),
		KubeClusterClient:     discoveryClientset.KubernetesClusters(),
		SettingsClient:        coreClientset.Settings(),
		SecretClient:          kubeCoreClientset.Secrets(),
		FailoverServiceClient: networkingClientset.FailoverServices(),
		VirtualServiceClient:  istioNetworkingClientset.VirtualServices(),
	}
}

func (k *KubeContext) Curl(ctx context.Context, ns, fromDeployment, fromContainer, url string) string {
	return kubectl.Curl(ctx, k.Context, ns, fromDeployment, fromContainer, url)
}

func (k *KubeContext) WaitForRollout(ctx context.Context, ns, deployment string) {
	kubectl.WaitForRollout(ctx, k.Context, ns, deployment)
}

func (k *KubeContext) DeployBookInfo(ctx context.Context, ns string) {
	kubectl.DeployBookInfo(ctx, k.Context, ns)
}

func (k *KubeContext) CreateNamespace(ctx context.Context, ns string) {
	kubectl.CreateNamespace(ctx, k.Context, ns)
}

func (k *KubeContext) DeleteNamespace(ctx context.Context, ns string) {
	kubectl.DeleteNamespace(ctx, k.Context, ns)
}

func (k *KubeContext) LabelNamespace(ctx context.Context, ns, label string) {
	kubectl.LabelNamespace(ctx, k.Context, ns, label)
}

func (k *KubeContext) SetDeploymentEnvVars(
	ctx context.Context,
	ns string,
	deploymentName string,
	containerName string,
	envVars map[string]string) {
	kubectl.SetDeploymentEnvVars(ctx, k.Context, ns, deploymentName, containerName, envVars)
}

// Modify the deployment's container entrypoint command to "sleep 20h" to disable the application.
func (k *KubeContext) DisableAppContainer(
	ctx context.Context,
	ns string,
	deploymentName string,
	containerName string,
) {
	kubectl.DisableAppContainer(ctx, k.Context, ns, deploymentName, containerName)
}

// Remove the sleep command to re-enable the application container.
func (k *KubeContext) EnableAppContainer(
	ctx context.Context,
	ns string,
	deploymentName string,
) {
	kubectl.EnableAppContainer(ctx, k.Context, ns, deploymentName)
}

type Pod struct {
	corev1.Pod
	Cluster *KubeContext
}

func (p *Pod) Curl(ctx context.Context, args ...string) string {
	return kubectl.CurlWithEphemeralPod(ctx, p.Cluster.Context, p.Namespace, p.Name, args...)
}

func (k *KubeContext) GetPod(ns, app string) *Pod {
	pl, err := k.Clientset.CoreV1().Pods(ns).List(v1.ListOptions{LabelSelector: "app=" + app})
	Expect(err).NotTo(HaveOccurred())
	Expect(pl.Items).NotTo(BeEmpty())

	return &Pod{
		Pod:     pl.Items[0],
		Cluster: k,
	}
}

var env Env
var envOnce sync.Once

func StartEnvOnce(ctx context.Context) Env {
	envOnce.Do(func() {
		env = StartEnv(ctx)
	})

	return env
}

func GetEnv() Env {
	return env
}

func ClearEnv(ctx context.Context) error {
	if useExisting := os.Getenv("USE_EXISTING"); useExisting != "" {
		// dont clear existing env
		return nil
	}
	cmd := exec.CommandContext(ctx, "./ci/setup-kind.sh", "cleanup", strconv.Itoa(GinkgoParallelNode()))
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	cmd.Dir = "../.."
	return cmd.Run()
}

func StartEnv(ctx context.Context) Env {

	if useExisting := os.Getenv("USE_EXISTING"); useExisting != "" {
		mgmt := "kind-management-plane-1"
		target := "kind-target-cluster-1"
		if fields := strings.Split(useExisting, ","); len(fields) == 2 {
			mgmt = fields[0]
			target = fields[1]
		}
		return newEnv(mgmt, target)
	}

	eg, ctx := errgroup.WithContext(ctx)

	r, w, err := os.Pipe()
	Expect(err).NotTo(HaveOccurred())
	defer r.Close()

	cmd := exec.CommandContext(ctx, "./ci/setup-kind.sh", strconv.Itoa(GinkgoParallelNode()))
	cmd.Dir = "../.."
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	cmd.ExtraFiles = append(cmd.ExtraFiles, w)
	err = cmd.Start()
	// close this end after start, as we dont need it.
	w.Close()
	Expect(err).NotTo(HaveOccurred())

	eg.Go(cmd.Wait)

	var data []byte
	eg.Go(func() error {
		var err error
		data, err = ioutil.ReadAll(r)
		return err
	})

	err = eg.Wait()
	if err != nil {
		dumpState()
	}

	Expect(err).NotTo(HaveOccurred())

	// read our contexts:
	fields := strings.Fields(string(data))
	return newEnv(fields[0], fields[1])
}

func dumpState() {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbgCmd := exec.CommandContext(timeoutCtx, "./ci/print-kind-info.sh", strconv.Itoa(GinkgoParallelNode()))
	dbgCmd.Dir = "../.."
	dbgCmd.Stdout = GinkgoWriter
	dbgCmd.Stderr = GinkgoWriter
	dbgCmd.Run()
}

func ParseYaml(yml string, msg interface{}) {
	var buf bytes.Buffer
	buf.WriteString(yml)
	decoder := yaml.NewYAMLOrJSONDecoder(&buf, 1024)
	err := decoder.Decode(msg)
	Expect(err).NotTo(HaveOccurred())
}
