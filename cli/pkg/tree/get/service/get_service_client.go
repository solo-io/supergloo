package get_service

import (
	"context"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

func GetMeshServices(
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
	services, err := kubeClients.MeshServiceClient.List(ctx)
	if err != nil {
		return err
	}
	serviceList := make([]*v1alpha1.MeshService, 0, len(services.Items))
	for _, v := range services.Items {
		v := v
		serviceList = append(serviceList, &v)
	}
	switch opts.Get.OutputFormat {
	case resource_printing.JSONFormat.String():
		return printers.ResourcePrinter.Print(out, services, resource_printing.JSONFormat)
	case resource_printing.YAMLFormat.String():
		return printers.ResourcePrinter.Print(out, services, resource_printing.YAMLFormat)
	default:
		return printers.MeshServicePrinter.Print(out, serviceList)
	}
}
