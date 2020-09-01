package helm

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"helm.sh/helm/v3/pkg/action"
)

type Uninstaller struct {
	KubeConfig     *register.KubeCfg
	Namespace      string
	ReleaseName    string
	Verbose        bool
	DryRun         bool
}

func (i Uninstaller) UninstallChart(ctx context.Context) error {
	kubeConfig := i.KubeConfig
	namespace := i.Namespace
	releaseName := i.ReleaseName
	verbose := i.Verbose
	dryRun := i.DryRun

	actionConfig, settings, err := newActionConfig(kubeConfig, namespace)
	if err != nil {
		return eris.Wrapf(err, "creating helm config")
	}
	settings.Debug = verbose

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

	return nil
}
