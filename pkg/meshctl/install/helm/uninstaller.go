package helm

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Uninstaller struct {
	KubeConfig  string
	KubeContext string
	Namespace   string
	ReleaseName string
	Verbose     bool
	DryRun      bool
}

func (i Uninstaller) UninstallChart(ctx context.Context) error {
	kubeConfig := i.KubeConfig
	kubeContext := i.KubeContext
	namespace := i.Namespace
	releaseName := i.ReleaseName
	verbose := i.Verbose
	dryRun := i.DryRun

	if kubeConfig != "" {
		kubeConfig = clientcmd.RecommendedHomeFile
	}

	actionConfig, settings, err := newActionConfig(kubeConfig, kubeContext, namespace)
	if err != nil {
		return eris.Wrapf(err, "creating helm config")
	}
	settings.Debug = verbose
	settings.KubeConfig = kubeConfig
	settings.KubeContext = kubeContext

	h, err := actionConfig.Releases.History(releaseName)
	if err == nil && len(h) > 0 {
		client := action.NewUninstall(actionConfig)
		client.DryRun = dryRun
		release, err := client.Run(releaseName)
		if err != nil {
			return eris.Wrapf(err, "uninstalling helm release %s", releaseName)
		}
		logrus.Infof("finished uninstalling release %s: %+v", releaseName, release)
	} else {
		logrus.Infof("release %s does not exist, nothing to uninstall", releaseName)
	}

	if err := deleteNamespace(ctx, kubeConfig, kubeContext, namespace); err != nil {
		return eris.Wrapf(err, "deleting namespace %s", namespace)
	}

	return nil
}

func deleteNamespace(ctx context.Context, kubeConfig, kubeContext, namespace string) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeConfig
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubeContext}

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return err
	}
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}
	namespaces := v1.NewNamespaceClient(c)
	return namespaces.DeleteNamespace(ctx, namespace)
}
