package helm

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/external-apis/pkg/api/k8s/apiextensions.k8s.io/v1beta1"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/labels"
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

	if kubeConfig == "" {
		kubeConfig = clientcmd.RecommendedHomeFile
	}

	kubeClient, err := utils.BuildClient(kubeConfig, kubeContext)
	if err != nil {
		return err
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
		_, err := client.Run(releaseName)
		if err != nil {
			return eris.Wrapf(err, "uninstalling helm release %s", releaseName)
		}

		if err := deleteCrds(ctx, kubeClient); err != nil {
			return eris.Wrapf(err, "deleting Gloo Mesh CRDs")
		}

		logrus.Infof("Finished uninstalling release %s", releaseName)
	} else {
		logrus.Warnf("release %s does not exist, nothing to uninstall", releaseName)
	}

	return nil
}

// Helm does not delete CRDs upon uninstall, https://helm.sh/docs/chart_best_practices/custom_resource_definitions/#some-caveats-and-explanations
// so we need custom logic for deleting Gloo Mesh CRDs
func deleteCrds(ctx context.Context, kubeClient client.Client) error {

	crdClient := v1beta1.NewCustomResourceDefinitionClient(kubeClient)

	return crdClient.DeleteAllOfCustomResourceDefinition(ctx, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			// Generated Gloo Mesh CRDs guaranteed to have label pair, reference: https://github.com/solo-io/skv2/blob/0cef270b47e41c15980e8a6b83c1735307953a39/codegen/render/manifests_renderer.go#L173
			LabelSelector: labels.SelectorFromSet(map[string]string{"app": "gloo-mesh"}),
		},
	})
}
