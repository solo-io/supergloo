package upgrade

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	DefaultReleaseTag = "latest"
	cmdName           = "upgrade"
)

type UpgradeCommand *cobra.Command

var UpgradeSet = wire.NewSet(
	UpgradeCmd,
)

func UpgradeCmd(ctx context.Context, opts *options.Options, out io.Writer, clients common.ClientsFactory) UpgradeCommand {
	cmd := &cobra.Command{
		Use:   cliconstants.UpgradeCommand.Use,
		Short: cliconstants.UpgradeCommand.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Upgrade(ctx, opts, out, clients)

		},
	}
	addFlags(cmd.PersistentFlags(), &opts.Upgrade)

	return cmd
}

func addFlags(flags *pflag.FlagSet, opts *options.Upgrade) {
	flags.StringVar(&opts.ReleaseTag, "release", DefaultReleaseTag, "Which meshctl release "+
		"to download. Specify a semver tag corresponding to the desired version of meshctl. "+
		"Service Mesh Hub releases can be found here: https://github.com/solo-io/service-mesh-hub/releases")
	flags.StringVar(&opts.DownloadPath, "path", "", "Desired path for your "+
		"upgraded meshctl binary. Defaults to the location of your currently executing binary.")
}
