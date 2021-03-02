package enterprise

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RegistrationOptions struct {
	registration.Options
	RelayServerAddress string
}

func RegisterCluster(ctx context.Context, opts RegistrationOptions) error {
	chartPath, err := opts.GetChartPath(ctx, opts.AgentChartPathOverride, gloomesh.EnterpriseAgentChartUriTemplate)
	if err != nil {
		return err
	}
	if err := (helm.Installer{
		KubeConfig:  opts.KubeConfigPath,
		KubeContext: opts.RemoteContext,
		ChartUri:    chartPath,
		Namespace:   opts.RemoteNamespace,
		ReleaseName: gloomesh.EnterpriseAgentReleaseName,
		ValuesFile:  opts.AgentChartValuesPath,
		Verbose:     opts.Verbose,
		Values: map[string]string{
			"relay.serverAddress": opts.RelayServerAddress,
			"relay.authority":     "enterprise-networking.gloo-mesh",
			"relay.insecure":      "true",
			"relay.cluster":       opts.ClusterName,
		},
	}).InstallChart(ctx); err != nil {
		return err
	}

	kubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	clusterClient := v1alpha1.NewKubernetesClusterClient(kubeClient)
	return clusterClient.CreateKubernetesCluster(ctx, &v1alpha1.KubernetesCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ClusterName,
			Namespace: opts.MgmtNamespace,
		},
		Spec: v1alpha1.KubernetesClusterSpec{
			ClusterDomain: opts.ClusterDomain,
		},
	})
}

func DeregisterCluster(ctx context.Context, opts RegistrationOptions) error {
	if err := (helm.Uninstaller{
		KubeConfig:  opts.KubeConfigPath,
		KubeContext: opts.RemoteContext,
		Namespace:   opts.RemoteNamespace,
		ReleaseName: gloomesh.EnterpriseAgentReleaseName,
		Verbose:     opts.Verbose,
	}).UninstallChart(ctx); err != nil {
		return err
	}

	kubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	clusterKey := client.ObjectKey{Name: opts.ClusterName, Namespace: opts.MgmtNamespace}
	return v1alpha1.NewKubernetesClusterClient(kubeClient).DeleteKubernetesCluster(ctx, clusterKey)
}
