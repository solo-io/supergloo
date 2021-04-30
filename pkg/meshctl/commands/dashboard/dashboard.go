package dashboard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/browser"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var ConsoleNotFoundError = errors.New("Console image not found. Your Gloo Mesh enterprise install may be a bad state.")

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Port forwards the Gloo Mesh Enterprise UI and opens it in a browser if available",
		RunE: func(cmd *cobra.Command, args []string) error {
			return forwardDashboard(ctx, opts.kubeconfig, opts.kubecontext, opts.namespace, opts.port)
		},
	}
	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
	port        uint32
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", "gloo-mesh", "The namespace that the Gloo Mesh UI is deployed in")
	flags.Uint32VarP(&o.port, "port", "p", 8090, "The local port to forward to the dashboard")
}

func forwardDashboard(ctx context.Context, kubeconfigPath, kubectx, namespace string, localPort uint32) error {
	staticPort, err := getStaticPort(ctx, kubeconfigPath, kubectx, namespace)
	if err != nil {
		return err
	}
	portFwdCmd, err := forwardPort(namespace, fmt.Sprint(localPort), staticPort)
	if err != nil {
		return err
	}
	defer portFwdCmd.Wait()
	if err := browser.OpenURL(fmt.Sprintf("http://localhost:%d", localPort)); err != nil {
		return err
	}

	return nil
}

func getStaticPort(ctx context.Context, kubeconfigPath, kubectx, namespace string) (string, error) {
	cfg, err := kubeconfig.GetRestConfigWithContext(kubeconfigPath, kubectx, "")
	if err != nil {
		return "", err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", err
	}
	dep, err := client.AppsV1().Deployments(namespace).Get(ctx, "dashboard", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			fmt.Printf("No Gloo Mesh dashboard found as part of the installation in namespace %s. "+
				"The full dashboard is part of Gloo Mesh enterprise by default. "+
				"Check that your kubeconfig is pointing at the Gloo Mesh management cluster. ", namespace)
		}

		return "", err
	}

	var staticPort string
	for _, container := range dep.Spec.Template.Spec.Containers {
		if container.Name == "console" {
			for _, port := range container.Ports {
				if port.Name == "static" {
					staticPort = fmt.Sprint(port.ContainerPort)
				}
			}
		}
	}
	if staticPort == "" {
		return "", ConsoleNotFoundError
	}

	return staticPort, nil
}

func forwardPort(namespace, localPort, kubePort string) (*exec.Cmd, error) {
	cmd := exec.Command(
		"kubectl", "port-forward", "-n", namespace, "deployment/dashboard", localPort+":"+kubePort,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := waitForDashboard(cmd, localPort); err != nil {
		return nil, err
	}

	return cmd, nil
}

func waitForDashboard(portFwdCmd *exec.Cmd, localPort string) error {
	ticker, timer := time.NewTicker(250*time.Millisecond), time.NewTimer(30*time.Second)
	errs := &multierror.Error{}
	for {
		err := func() error {
			res, err := http.Get("http://localhost:" + localPort)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("invalid status code: %d %s", res.StatusCode, res.Status)
			}
			io.Copy(ioutil.Discard, res.Body)
			return nil
		}()
		if err == nil {
			return nil
		}

		errs = multierror.Append(errs, err)

		select {
		case <-timer.C:
			if portFwdCmd.Process != nil {
				portFwdCmd.Process.Kill()
				portFwdCmd.Process.Release()
			}

			return fmt.Errorf("timed out waiting for dashboard port forward to be ready: %s", errs.Error())
		case <-ticker.C:
		}
	}
}
