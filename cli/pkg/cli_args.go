package cli

import (
	"context"
	"io/ioutil"

	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common/usage"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/check"
	clusterroot "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/create"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/demo"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/describe"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/get"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/install"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/mesh"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/uninstall"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/upgrade"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/version"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/grpclog"
)

// build an instance of the meshctl implementation
func BuildCli(
	ctx context.Context,
	opts *options.Options,
	usageReporter usageclient.Client,
	clusterCmd clusterroot.ClusterCommand,
	versionCmd version.VersionCommand,
	meshCmd mesh.MeshCommand,
	upgradeCmd upgrade.UpgradeCommand,
	installCmd install.InstallCommand,
	uninstallCmd uninstall.UninstallCommand,
	checkCommand check.CheckCommand,
	describeCommand describe.DescribeCommand,
	demoCommand demo.DemoCommand,
	getCommand get.GetCommand,
	createCommand create.CreateRootCmd,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cliconstants.RootCommand.Use,
		Short: cliconstants.RootCommand.Short,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			//TODO: this is failing with transport: loopyWriter.run returning. connection error: desc = "transport is closing"
			grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
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
		meshCmd,
		uninstallCmd,
		checkCommand,
		describeCommand,
		demoCommand,
		createCommand,
		getCommand,
	)
	return cmd
}
