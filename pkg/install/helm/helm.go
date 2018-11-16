package helm

import (
	"fmt"

	"github.com/spf13/pflag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
)

// Create a tunnel to tiller, set up a helm client, and ping it to ensure the connection is live
// Consumers are expected to call Teardown to ensure the tunnel gets closed
// TODO: Expose configuration options inside setupConnection()
func GetHelmClient() (*helm.Client, error) {
	if err := setupConnection(); err != nil {
		return nil, err
	}
	options := []helm.Option{helm.Host(settings.TillerHost), helm.ConnectTimeout(settings.TillerConnectionTimeout)}
	helmClient := helm.NewClient(options...)
	if err := helmClient.PingTiller(); err != nil {
		return nil, err
	}
	return helmClient, nil
}

var (
	tillerTunnel *kube.Tunnel
	settings     helm_env.EnvSettings
)

func Teardown() {
	if tillerTunnel != nil {
		tillerTunnel.Close()
	}
}

// configForContext creates a Kubernetes REST client configuration for a given kubeconfig context.
func configForContext(context string, kubeconfig string) (*rest.Config, error) {
	config, err := kube.GetConfig(context, kubeconfig).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %s", context, err)
	}
	return config, nil
}

// getKubeClient creates a Kubernetes config and client for a given kubeconfig context.
func getKubeClient(context string, kubeconfig string) (*rest.Config, kubernetes.Interface, error) {
	config, err := configForContext(context, kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}

func setupConnection() error {
	var flagSet pflag.FlagSet
	settings.AddFlags(&flagSet)
	if settings.TillerHost == "" {
		config, client, err := getKubeClient(settings.KubeContext, settings.KubeConfig)
		if err != nil {
			return err
		}

		tillerTunnel, err = portforwarder.New(settings.TillerNamespace, client, config)
		if err != nil {
			return err
		}

		settings.TillerHost = fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local)
		//TODO: remove me
		fmt.Printf("Created tunnel using local port: '%d'\n", tillerTunnel.Local)
	}

	// Set up the gRPC config.
	// TODO: remove me
	fmt.Printf("SERVER: %q\n", settings.TillerHost)

	// Plugin support.
	return nil
}
