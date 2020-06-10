package get_vmcsr

import (
	"context"
	"io"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kube/kubeconfig"
)

func GetVirtualMeshCertificateSigningRequests(
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
	virtualMeshCSRs, err := kubeClients.VirtualMeshCSRClient.ListVirtualMeshCertificateSigningRequest(ctx)
	if err != nil {
		return err
	}
	virtualMeshCSRList := make([]*smh_security.VirtualMeshCertificateSigningRequest, 0, len(virtualMeshCSRs.Items))
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
