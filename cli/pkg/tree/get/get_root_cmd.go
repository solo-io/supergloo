package get

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/common/resource_printing"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	get_cluster "github.com/solo-io/mesh-projects/cli/pkg/tree/get/cluster"
	get_mesh "github.com/solo-io/mesh-projects/cli/pkg/tree/get/mesh"
	get_service "github.com/solo-io/mesh-projects/cli/pkg/tree/get/service"
	get_virtual_mesh "github.com/solo-io/mesh-projects/cli/pkg/tree/get/virtual_mesh"
	get_vmcsr "github.com/solo-io/mesh-projects/cli/pkg/tree/get/vmcsr"
	get_workload "github.com/solo-io/mesh-projects/cli/pkg/tree/get/workload"
	"github.com/spf13/cobra"
)

type GetCommand *cobra.Command

var (
	GetSet = wire.NewSet(
		get_mesh.GetMeshSet,
		get_service.GetServiceSet,
		get_workload.GetWorkloadSet,
		get_cluster.GetClusterSet,
		get_virtual_mesh.GetVirtualMeshSet,
		get_vmcsr.GetVirtualMeshCSRSet,
		GetRootCommand,
	)
	prettyFormat       = "pretty"
	validOutputFormats = []string{prettyFormat, resource_printing.JSONFormat.String(), resource_printing.YAMLFormat.String()}
)

func GetRootCommand(
	getMeshCommand get_mesh.GetMeshCommand,
	getWorkloadCommand get_workload.GetWorkloadCommand,
	getServiceCommand get_service.GetServiceCommand,
	getClusterCommand get_cluster.GetClusterCommand,
	getVirtualMeshCommand get_virtual_mesh.GetVirtualMeshCommand,
	getVirtualMeshCSRCommand get_vmcsr.GetVirtualMeshCSRCommand,
	opts *options.Options,
) GetCommand {
	cmd := &cobra.Command{
		Use:     cliconstants.GetCommand.Root.Use,
		Aliases: cliconstants.GetCommand.Root.Aliases,
		Short:   cliconstants.GetCommand.Root.Short,
		RunE:    common.NonTerminalCommand(cliconstants.GetCommand.Root.Use),
	}

	cmd.AddCommand(
		getMeshCommand,
		getServiceCommand,
		getWorkloadCommand,
		getClusterCommand,
		getVirtualMeshCommand,
		getVirtualMeshCSRCommand,
	)

	options.AddGetFlags(cmd, opts, prettyFormat, validOutputFormats)
	return cmd
}
