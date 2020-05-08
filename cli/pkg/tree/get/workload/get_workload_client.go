package get_workload

import (
	"context"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
)

func GetMeshWorkloads(
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
	meshWorkloads, err := kubeClients.MeshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return err
	}
	meshWorkloadList := make([]*zephyr_discovery.MeshWorkload, 0, len(meshWorkloads.Items))
	for _, v := range meshWorkloads.Items {
		v := v
		meshWorkloadList = append(meshWorkloadList, &v)
	}
	switch opts.Get.OutputFormat {
	case resource_printing.JSONFormat.String():
		return printers.ResourcePrinter.Print(out, meshWorkloads, resource_printing.JSONFormat)
	case resource_printing.YAMLFormat.String():
		return printers.ResourcePrinter.Print(out, meshWorkloads, resource_printing.YAMLFormat)
	default:
		return printers.MeshWorkloadPrinter.Print(out, meshWorkloadList)
	}
}
