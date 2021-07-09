package checks

import (
	"context"
	"errors"
	"fmt"
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

const (
	mgmtDeployName = "enterprise-networking"
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
				MetricsPort: defaults.MetricsPort,
				Namespace:   ns,
				InCluster:   true,
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
				MetricsPort: remotePort,
				Namespace:   ns,
				InCluster:   false,
			}},
	}

}

func (c *InClusterCheckContext) AccessMgmtServerAdminPort(ctx context.Context, op func(ctx context.Context, addr string) (error, string)) (error, string) {

	// note: the metrics port is not exposed on the service (as it should be).
	// so we need to find the ip of the deployed pod:
	d, err := v1.NewDeploymentClient(c.Cli).GetDeployment(ctx, client.ObjectKey{
		Namespace: c.Env.Namespace,
		Name:      mgmtDeployName,
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
	return op(ctx, fmt.Sprintf("http://%v:%v/metrics", pod.Status.PodIP, c.Env.MetricsPort))
}

func (c *OutOfClusterCheckContext) AccessMgmtServerAdminPort(ctx context.Context, op func(ctx context.Context, addr string) (error, string)) (error, string) {
	shouldRunCheck, err := isEnterpriseVersion(ctx, c.Cli, c.Env.Namespace)
	if err != nil {
		return err, ""
	}

	if !shouldRunCheck {
		contextutils.LoggerFrom(ctx).Debugf("skipping relay connectivity check, enterprise not detected")
		return nil, ""
	}

	portFwdContext, cancelPtFwd := context.WithCancel(ctx)
	defer cancelPtFwd()

	// start port forward to mgmt server stats port
	localPort, err := utils.PortForwardFromDeployment(
		portFwdContext,
		c.mgmtKubeConfig,
		c.mgmtKubeContext,
		mgmtDeployName,
		c.Env.Namespace,
		fmt.Sprintf("%v", c.localPort),
		fmt.Sprintf("%v", c.remotePort),
	)
	if err != nil {
		return err, fmt.Sprintf("try verifying that `kubectl port-forward -n %v deployment/%v %v:%v` can be run successfully.", c.Env.Namespace, mgmtDeployName, c.localPort, c.remotePort)
	}
	// request metrics page from mgmt deployment
	metricsUrl := fmt.Sprintf("http://localhost:%v/metrics", localPort)
	return op(portFwdContext, metricsUrl)
}

func isEnterpriseVersion(ctx context.Context, c client.Client, installNamespace string) (bool, error) {
	_, err := v1.NewDeploymentClient(c).GetDeployment(ctx, client.ObjectKey{
		Namespace: installNamespace,
		Name:      mgmtDeployName,
	})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
