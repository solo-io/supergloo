package e2e

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/rest"

	"github.com/solo-io/service-mesh-hub/test/kubectl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istionetworkingv1alpha3 "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3"
	kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	mgmtContext   = "kind-mgmt-cluster"
	remoteContext = "kind-remote-cluster"
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

type SingleClusterEnv struct {
	Management KubeContext
}

func (s SingleClusterEnv) DumpState() {
	dumpState()
}

func newSingleClusterEnv(mgmt string) SingleClusterEnv {
	return SingleClusterEnv{
		Management: NewKubeContext(mgmt),
	}
}

type KubeContext struct {
	Context               string
	Config                *rest.Config
	Clientset             *kubernetes.Clientset
	TrafficPolicyClient   networkingv1alpha2.TrafficPolicyClient
	MeshClient            discoveryv1alpha2.MeshClient
	SecretClient          kubernetes_core.SecretClient
	VirtualMeshClient     networkingv1alpha2.VirtualMeshClient
	DestinationRuleClient istionetworkingv1alpha3.DestinationRuleClient
}

// If kubecontext is empty string, use current context.
func NewKubeContext(kubecontext string) KubeContext {
	cfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	Expect(err).NotTo(HaveOccurred())

	config := clientcmd.NewNonInteractiveClientConfig(*cfg, kubecontext, &clientcmd.ConfigOverrides{}, nil)
	restcfg, err := config.ClientConfig()
	Expect(err).NotTo(HaveOccurred())

	clientset, err := kubernetes.NewForConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	kubeCoreClientset, err := kubernetes_core.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	networkingClientset, err := networkingv1alpha2.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	discoveryClientset, err := discoveryv1alpha2.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	istioNetworkingClientset, err := istionetworkingv1alpha3.NewClientsetFromConfig(restcfg)
	Expect(err).NotTo(HaveOccurred())

	return KubeContext{
		Context:               kubecontext,
		Config:                restcfg,
		Clientset:             clientset,
		TrafficPolicyClient:   networkingClientset.TrafficPolicies(),
		VirtualMeshClient:     networkingClientset.VirtualMeshes(),
		MeshClient:            discoveryClientset.Meshes(),
		SecretClient:          kubeCoreClientset.Secrets(),
		DestinationRuleClient: istioNetworkingClientset.DestinationRules(),
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
func (k *KubeContext) DisableContainer(
	ctx context.Context,
	ns string,
	deploymentName string,
	containerName string,
) {
	kubectl.DisableContainer(ctx, k.Context, ns, deploymentName, containerName)
}

// Remove the sleep command to re-enable the application container.
func (k *KubeContext) EnableContainer(
	ctx context.Context,
	ns string,
	deploymentName string,
) {
	kubectl.EnableContainer(ctx, k.Context, ns, deploymentName)
}

type Pod struct {
	corev1.Pod
	Cluster *KubeContext
}

func (p *Pod) Curl(ctx context.Context, args ...string) string {
	return kubectl.CurlWithEphemeralPod(ctx, p.Cluster.Context, p.Namespace, p.Name, args...)
}

func (k *KubeContext) GetPod(ctx context.Context, ns, app string) *Pod {
	pl, err := k.Clientset.CoreV1().Pods(ns).List(ctx, v1.ListOptions{LabelSelector: "app=" + app})
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
	if useExisting := os.Getenv("USE_EXISTING"); useExisting == "1" {
		// dont clear existing env
		return nil
	}
	cmd := exec.CommandContext(ctx, "./ci/setup-kind.sh", "cleanup", strconv.Itoa(GinkgoParallelNode()))
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	return cmd.Run()
}

func StartEnv(ctx context.Context) Env {

	if useExisting := os.Getenv("USE_EXISTING"); useExisting == "1" {
		mgmt := "kind-mgmt-cluster"
		remote := "kind-remote-cluster"
		if fields := strings.Split(useExisting, ","); len(fields) == 2 {
			mgmt = fields[0]
			remote = fields[1]
		}
		return newEnv(mgmt, remote)
	}

	cmd := exec.CommandContext(ctx, "./ci/setup-kind.sh", strconv.Itoa(GinkgoParallelNode()))
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter

	err := cmd.Run()
	if err != nil {
		dumpState()
	}
	Expect(err).NotTo(HaveOccurred())

	return newEnv(mgmtContext, remoteContext)
}

var (
	singleClusterEnv     SingleClusterEnv
	singleClusterEnvOnce sync.Once
)

func StartSingleClusterEnvOnce(ctx context.Context) SingleClusterEnv {
	singleClusterEnvOnce.Do(func() {
		singleClusterEnv = StartSingleClusterEnv(ctx)
	})

	return singleClusterEnv
}

func GetSingleClusterEnv() SingleClusterEnv {
	return singleClusterEnv
}

func ClearSingleClusterEnv(ctx context.Context) error {
	if useExisting := os.Getenv("USE_EXISTING"); useExisting == "1" {
		// dont clear existing env
		return nil
	}
	cmd := exec.CommandContext(ctx, "./ci/setup-kind.sh", "cleanup", strconv.Itoa(GinkgoParallelNode()))
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	return cmd.Run()
}

func StartSingleClusterEnv(ctx context.Context) SingleClusterEnv {

	if useExisting := os.Getenv("USE_EXISTING"); useExisting == "1" {
		return newSingleClusterEnv(mgmtContext)
	}

	// TODO: don't hardcode osm installation here
	cmd := exec.CommandContext(ctx, "./ci/setup-kind.sh", "osm")
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter

	err := cmd.Run()
	if err != nil {
		dumpState()
	}
	Expect(err).NotTo(HaveOccurred())

	return newSingleClusterEnv(mgmtContext)
}

func dumpState() {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbgCmd := exec.CommandContext(timeoutCtx, "./ci/print-kind-info.sh", strconv.Itoa(GinkgoParallelNode()))
	dbgCmd.Dir = "../.."
	dbgCmd.Stdout = GinkgoWriter
	dbgCmd.Stderr = GinkgoWriter
	_ = dbgCmd.Run()
}
