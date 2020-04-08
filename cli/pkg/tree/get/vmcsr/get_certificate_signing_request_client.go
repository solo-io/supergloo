package get_vmcsr

import (
	"context"
	"io"

	"github.com/solo-io/mesh-projects/cli/pkg/common"
	common_config "github.com/solo-io/mesh-projects/cli/pkg/common/config"
	"github.com/solo-io/mesh-projects/cli/pkg/common/resource_printing"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
)

func GetVirtualMeshCertificateSigningRequests(
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
	virtualMeshCSRs, err := kubeClients.VirtualMeshCSRClient.List(ctx)
	if err != nil {
		return err
	}
	virtualMeshCSRList := make([]*v1alpha1.VirtualMeshCertificateSigningRequest, 0, len(virtualMeshCSRs.Items))
	for _, v := range virtualMeshCSRs.Items {
		v := v
		virtualMeshCSRList = append(virtualMeshCSRList, &v)
	}
	switch opts.Get.OutputFormat {
	case resource_printing.JSONFormat.String():
		return printers.ResourcePrinter.Print(out, virtualMeshCSRs, resource_printing.JSONFormat)
	case resource_printing.YAMLFormat.String():
		return printers.ResourcePrinter.Print(out, virtualMeshCSRs, resource_printing.YAMLFormat)
	default:
		return printers.VirtualMeshCSRPrinter.Print(out, virtualMeshCSRList)
	}
}
