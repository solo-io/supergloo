package get_virtual_mesh

import (
	"context"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetVirtualMeshes(
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
	virtualMeshes, err := kubeClients.VirtualMeshClient.ListVirtualMesh(ctx)
	if err != nil {
		return err
	}
	virtualMeshList := make([]*zephyr_networking.VirtualMesh, 0, len(virtualMeshes.Items))
	meshList := make([]*zephyr_discovery.Mesh, 0, len(virtualMeshes.Items))
	for _, v := range virtualMeshes.Items {
		v := v
		virtualMeshList = append(virtualMeshList, &v)
		if len(v.Spec.GetMeshes()) == 0 {
			meshList = append(meshList, nil)
		}
		mesh, err := kubeClients.MeshClient.GetMesh(
			ctx,
			client.ObjectKey{Name: v.Spec.GetMeshes()[0].GetName(), Namespace: v.Spec.GetMeshes()[0].GetNamespace()},
		)
		if err != nil {
			return err
		}
		meshList = append(meshList, mesh)
	}
	switch opts.Get.OutputFormat {
	case resource_printing.JSONFormat.String():
		return printers.ResourcePrinter.Print(out, virtualMeshes, resource_printing.JSONFormat)
	case resource_printing.YAMLFormat.String():
		return printers.ResourcePrinter.Print(out, virtualMeshes, resource_printing.YAMLFormat)
	default:
		return printers.VirtualMeshPrinter.Print(out, virtualMeshList, meshList)
	}
}
