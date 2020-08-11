package agent

import (
	"context"
	"fmt"

	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/helm"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var settings = cli.New()

const certAgentChartUriTemplate = "https://storage.googleapis.com/service-mesh-hub/cert-agent/cert-agent-%s.tgz"

type Installer struct {
	HelmChartOverride string
	HelmValuesPath    string
	KubeConfig        string
	KubeContext       string
	Namespace         string
	Verbose           bool
}

func (i Installer) InstallCertAgent(
	ctx context.Context,
) error {

	helmChartOverride := i.HelmChartOverride
	helmValuesPath := i.HelmValuesPath
	kubeConfig := i.KubeConfig
	kubeContext := i.KubeContext
	namespace := i.Namespace
	verbose := i.Verbose

	if helmChartOverride == "" {
		helmChartOverride = fmt.Sprintf(certAgentChartUriTemplate, version.Version)
	}

	if err := ensureNamespace(ctx, kubeConfig, kubeContext, namespace); err != nil {
		return err
	}

	return helm.Installer{
		KubeConfig:  kubeConfig,
		KubeContext: kubeContext,
		ChartUri:    helmChartOverride,
		Namespace:   namespace,
		ReleaseName: "cert-agent",
		ValuesFile:  helmValuesPath,
		Verbose:     verbose,
	}.InstallChart(ctx)
}

func ensureNamespace(ctx context.Context, kubeConfig, kubeContext, namespace string) error {
	if kubeConfig != "" {
		kubeConfig = clientcmd.RecommendedHomeFile
	}

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
	return namespaces.UpsertNamespace(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Spec: corev1.NamespaceSpec{Finalizers: []corev1.FinalizerName{"kubernetes"}},
	})
}
