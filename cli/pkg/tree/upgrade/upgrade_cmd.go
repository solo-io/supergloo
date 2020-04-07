package upgrade

import (
	"context"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/spf13/cobra"
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
	options.AddUpgradeFlags(cmd, &opts.Upgrade)
	return cmd
}
