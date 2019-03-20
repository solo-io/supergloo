package create

import (
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "commands for creating managing resources used for SuperGloo",
		Long:    "commands for creating managing resources used for SuperGloo",
	}

	cmd.AddCommand(createSecretCommand(opts))
	return cmd
}

func createSecretCommand(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"s"},
		Short:   "create a secret for use with SuperGloo.",
		Long: `SuperGloo uses secrets to authenticate to external APIs
and manage TLS certificates used for encryption in the mesh and ingress.
`,
	}
	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)

	cmd.AddCommand(tlsCmd(opts))
	cmd.AddCommand(awsCmd(opts))
	return cmd
}
