package get_cluster

import (
	"context"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
)

func GetClusters(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	factory common.KubeClientsFactory,
	kubeLoader kubeconfig.KubeLoader,
	opts *options.Options,
) error {
	cfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
	if err != nil {
		return err
	}
	kubeClients, err := factory(cfg, opts.Root.WriteNamespace)
	if err != nil {
		return err
	}
	clusters, err := kubeClients.KubeClusterClient.ListKubernetesCluster(ctx)
	if err != nil {
		return err
	}
	clusterList := make([]*smh_discovery.KubernetesCluster, 0, len(clusters.Items))
	for _, v := range clusters.Items {
		v := v
		clusterList = append(clusterList, &v)
	}
	switch opts.Get.OutputFormat {
	case resource_printing.JSONFormat.String():
		return printers.ResourcePrinter.Print(out, clusters, resource_printing.JSONFormat)
	case resource_printing.YAMLFormat.String():
		return printers.ResourcePrinter.Print(out, clusters, resource_printing.YAMLFormat)
	default:
		return printers.KubeClusterPrinter.Print(out, clusterList)
	}
}
