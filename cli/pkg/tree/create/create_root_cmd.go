package create

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	access_control_policy "github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/access-control-policy"
	traffic_policy "github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/traffic-policy"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/virtualmesh"
	"github.com/spf13/cobra"
)

var (
	CreateSet = wire.NewSet(
		CreateRootCommand,
		virtualmesh.CreateVirtualMeshCommand,
		traffic_policy.CreateTrafficPolicyCommand,
		access_control_policy.CreateAccessControlPolicyCommand,
	)
)

type CreateRootCmd *cobra.Command

func CreateRootCommand(
	opts *options.Options,
	createVirtualMeshCmd virtualmesh.CreateVirtualMeshCmd,
	createTrafficPolicyCmd traffic_policy.CreateTrafficPolicyCmd,
	createAccessControlPolicyCmd access_control_policy.CreateAccessControlPolicyCmd,
) CreateRootCmd {
	cmd := &cobra.Command{
		Use:   cliconstants.CreateCommand.Use,
		Short: cliconstants.CreateCommand.Short,
		RunE:  common.NonTerminalCommand(cliconstants.CreateCommand.Use),
	}
	options.AddCreateFlags(cmd, opts, resource_printing.YAMLFormat.String(), resource_printing.ValidFormats)
	cmd.AddCommand(createVirtualMeshCmd)
	cmd.AddCommand(createTrafficPolicyCmd)
	cmd.AddCommand(createAccessControlPolicyCmd)
	return cmd
}
