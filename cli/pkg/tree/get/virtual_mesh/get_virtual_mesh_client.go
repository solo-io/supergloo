package get_virtual_mesh

import (
	"context"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

func GetVirtualMeshes(
	ctx context.Context,
	out io.Writer,
	printers common.Printers,
	factory common.KubeClientsFactory,
	kubeLoader common_config.KubeLoader,
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
	virtualMeshes, err := kubeClients.VirtualMeshClient.ListVirtualMesh(ctx)
	if err != nil {
		return err
	}
	virtualMeshList := make([]*networking_v1alpha1.VirtualMesh, 0, len(virtualMeshes.Items))
	for _, v := range virtualMeshes.Items {
		v := v
		virtualMeshList = append(virtualMeshList, &v)
	}
	switch opts.Get.OutputFormat {
	case resource_printing.JSONFormat.String():
		return printers.ResourcePrinter.Print(out, virtualMeshes, resource_printing.JSONFormat)
	case resource_printing.YAMLFormat.String():
		return printers.ResourcePrinter.Print(out, virtualMeshes, resource_printing.YAMLFormat)
	default:
		return printers.VirtualMeshPrinter.Print(out, virtualMeshList)
	}
}
