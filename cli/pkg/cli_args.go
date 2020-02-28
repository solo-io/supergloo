package cli

import (
	"context"

	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common/usage"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	clusterroot "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/install"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/istio"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	"github.com/spf13/cobra"
)

// build an instance of the meshctl implementation
func BuildCli(
	ctx context.Context,
	opts *options.Options,
	usageReporter usageclient.Client,
	clusterCmd clusterroot.ClusterCommand,
	versionCmd version.VersionCommand,
	istioCmd istio.IstioCommand,
	upgradeCmd upgrade.UpgradeCommand,
	installCmd install.InstallCommand,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cliconstants.RootCommand.Use,
		Short: cliconstants.RootCommand.Short,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			usageReporter.StartReportingUsage(ctx, usage.UsageReportingInterval)
			return nil
		},
	}
	options.AddRootFlags(cmd, opts)
	cmd.AddCommand(
		clusterCmd,
		versionCmd,
		installCmd,
		upgradeCmd,
		istioCmd)
	return cmd
}
