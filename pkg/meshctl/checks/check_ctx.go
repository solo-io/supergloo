package checks

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	v1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/go-utils/contextutils"
	skutils "github.com/solo-io/skv2/pkg/utils"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CommonContext struct {
	Env Environment
	Cli client.Client
}

func (c *CommonContext) Environment() Environment {
	return c.Env
}
func (c *CommonContext) Client() client.Client {
	return c.Cli
}

type InClusterCheckContext struct {
	CommonContext
}

type OutOfClusterCheckContext struct {
	CommonContext

	mgmtKubeConfig  string
	mgmtKubeContext string
	localPort       uint32
	remotePort      uint32
}

func NewInClusterCheckContext() (CheckContext, error) {
	kubeClient, err := utils.BuildClient("", "")
	if err != nil {
		return nil, err
	}
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns, err = skutils.GetInClusterNamesapce()
		if err != nil {
			return nil, err
		}
	}
	return &InClusterCheckContext{
		CommonContext: CommonContext{
			Cli: kubeClient,
			Env: Environment{
				AdminPort: defaults.MetricsPort,
				Namespace: ns,
				InCluster: true,
			}},
	}, nil
}

func NewOutOfClusterCheckContext(cli client.Client, ns, mgmtKubeConfig, mgmtKubeContext string,
	localPort, remotePort uint32) CheckContext {
	return &OutOfClusterCheckContext{
		remotePort:      remotePort,
		localPort:       localPort,
		mgmtKubeConfig:  mgmtKubeConfig,
		mgmtKubeContext: mgmtKubeContext,
		CommonContext: CommonContext{
			Cli: cli,
			Env: Environment{
				AdminPort: remotePort,
				Namespace: ns,
				InCluster: false,
			}},
	}

}

func (c *InClusterCheckContext) AccessAdminPort(ctx context.Context, deployment string, op func(ctx context.Context, adminUrl *url.URL) (error, string)) (error, string) {

	// note: the metrics port is not exposed on the service (it should not be, so this is fine).
	// so we need to find the ip of the deployed pod:
	d, err := v1.NewDeploymentClient(c.Cli).GetDeployment(ctx, client.ObjectKey{
		Namespace: c.Env.Namespace,
		Name:      deployment,
	})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return err, "gloo-mesh enterprise deployment not found. Is gloo-mesh installed in this namespace?"
		}
		return err, ""
	}
	selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return err, ""
	}
	lo := &client.ListOptions{
		Namespace:     c.Env.Namespace,
		LabelSelector: selector,
		Limit:         1,
	}
	podsList, err := corev1.NewPodClient(c.Cli).ListPod(ctx, lo)
	if err != nil {
		return err, "failed listing deployment pods. is gloo-mesh installed?"
	}
	pods := podsList.Items
	if len(pods) == 0 {
		return err, "no pods are available for deployemt. please check your gloo-mesh installation?"
	}
	if podsList.RemainingItemCount != nil && *podsList.RemainingItemCount != 0 {
		contextutils.LoggerFrom(ctx).Info("You have more than one pod for gloo-mesh deployment. This test may not be accurate.")
	}
	pod := pods[0]
	if pod.Status.PodIP == "" {
		return errors.New("no pod ip"), "gloo-mesh pod doesn't have an IP address. This is usually temporary. please wait or check your gloo-mesh installation?"
	}
	adminUrl := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%v:%v", pod.Status.PodIP, c.Env.AdminPort),
	}

	return op(ctx, adminUrl)
}

func (c *OutOfClusterCheckContext) AccessAdminPort(ctx context.Context, deployment string, op func(ctx context.Context, adminUrl *url.URL) (error, string)) (error, string) {
	portFwdContext, cancelPtFwd := context.WithCancel(ctx)
	defer cancelPtFwd()

	// start port forward to mgmt server stats port
	localPort, err := utils.PortForwardFromDeployment(
		portFwdContext,
		c.mgmtKubeConfig,
		c.mgmtKubeContext,
		deployment,
		c.Env.Namespace,
		fmt.Sprintf("%v", c.localPort),
		fmt.Sprintf("%v", c.remotePort),
	)
	if err != nil {
		return err, fmt.Sprintf("try verifying that `kubectl port-forward -n %v deployment/%v %v:%v` can be run successfully.", c.Env.Namespace, deployment, c.localPort, c.remotePort)
	}
	// request metrics page from mgmt deployment
	adminUrl := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%v", localPort),
	}

	return op(portFwdContext, adminUrl)
}
