package helm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/solo-io/solo-kit/pkg/utils/contextutils"

	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/repo"

	"github.com/spf13/pflag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
)

type HelmClient interface {
	GetHelmClient(ctx context.Context) (*helm.Client, error)
	LocateChartRepoReleaseDefault(ctx context.Context, repoUrl string, release string) (string, error)
	Teardown()
}

type KubeHelmClient struct{}

func setupTillerHost(ctx context.Context) {
	tiller := "tiller-deploy.kube-system.svc.cluster.local"
	_, err := net.LookupIP(tiller)
	if err == nil {
		contextutils.LoggerFrom(ctx).Infof("Setting tiller host to in-cluster host.")
		Settings.TillerHost = tiller + ":44134"
	} else {
		contextutils.LoggerFrom(ctx).Infof("Can't resolve tiller in the cluster")
	}
}

// Create a tunnel to tiller, set up a helm client, and ping it to ensure the connection is live
// Consumers are expected to call Teardown to ensure the tunnel gets closed
// TODO: Expose configuration options inside setupConnection()
func (client *KubeHelmClient) GetHelmClient(ctx context.Context) (*helm.Client, error) {
	if err := setupConnection(ctx); err != nil {
		return nil, err
	}
	options := []helm.Option{helm.Host(Settings.TillerHost), helm.ConnectTimeout(Settings.TillerConnectionTimeout)}
	helmClient := helm.NewClient(options...)
	if err := helmClient.PingTiller(); err != nil {
		return nil, err
	}
	return helmClient, nil
}

var (
	tillerTunnel *kube.Tunnel
	Settings     helm_env.EnvSettings
)

func (client *KubeHelmClient) Teardown() {
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

func setupConnection(ctx context.Context) error {
	var flagSet pflag.FlagSet
	Settings.AddFlags(&flagSet)

	// Test if we are in the cluster and set tiller host accordingly
	setupTillerHost(ctx)

	if Settings.TillerHost == "" {
		config, client, err := getKubeClient(Settings.KubeContext, Settings.KubeConfig)
		if err != nil {
			return err
		}

		tillerTunnel, err = portforwarder.New(Settings.TillerNamespace, client, config)
		if err != nil {
			return err
		}

		Settings.TillerHost = fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local)
		contextutils.LoggerFrom(ctx).Infof("Created tunnel using local port: '%d'\n", tillerTunnel.Local)
	}

	// Set up the gRPC config.
	contextutils.LoggerFrom(ctx).Infof("SERVER: %q\n", Settings.TillerHost)

	// Plugin support.
	return nil
}

func (client *KubeHelmClient) LocateChartRepoReleaseDefault(ctx context.Context, repoUrl string, release string) (string, error) {
	return locateChartPath(ctx, repoUrl, "", "", release, "", false, "", "", "", "")
}

// locateChartPath looks for a chart directory in known places, and returns either the full path or an error.
//
// This does not ensure that the chart is well-formed; only that the requested filename exists.
//
// Order of resolution:
// - current working directory
// - if path is absolute or begins with '.', error out here
// - chart repos in $HELM_HOME
// - URL
//
// If 'verify' is true, this will attempt to also verify the chart.
func locateChartPath(ctx context.Context, repoURL, username, password, name, version string, verify bool, keyring,
	certFile, keyFile, caFile string) (string, error) {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if fi, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)
		if err != nil {
			return abs, err
		}
		if verify {
			if fi.IsDir() {
				return "", errors.New("cannot verify a directory")
			}
			if _, err := downloader.VerifyChart(abs, keyring); err != nil {
				return "", err
			}
		}
		return abs, nil
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	crepo := filepath.Join(Settings.Home.Repository(), name)
	if _, err := os.Stat(crepo); err == nil {
		return filepath.Abs(crepo)
	}

	dl := downloader.ChartDownloader{
		HelmHome: Settings.Home,
		Out:      os.Stdout,
		Keyring:  keyring,
		Getters:  getter.All(Settings),
		Username: username,
		Password: password,
	}
	if verify {
		dl.Verify = downloader.VerifyAlways
	}
	if repoURL != "" {
		chartURL, err := repo.FindChartInAuthRepoURL(repoURL, username, password, name, version,
			certFile, keyFile, caFile, getter.All(Settings))
		if err != nil {
			return "", err
		}
		name = chartURL
	}

	if _, err := os.Stat(Settings.Home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(Settings.Home.Archive(), 0744)
	}

	filename, _, err := dl.DownloadTo(name, version, Settings.Home.Archive())
	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			return filename, err
		}
		//debug("Fetched %s to %s\n", name, filename)
		return lname, nil
	} else if Settings.Debug {
		return filename, err
	}

	contextutils.LoggerFrom(ctx).Warnw("Failed to download chart", "error", err, "chart", name)

	return filename, fmt.Errorf("failed to download %q (hint: running `helm repo update` may help)", name)
}
