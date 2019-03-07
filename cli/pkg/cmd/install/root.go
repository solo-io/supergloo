package install

import (
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "install a service mesh using Supergloo",
		Long: `Creates an Install resource which the supergloo controller 
will use to install a service mesh.

Installs represent a desired installation of a supported mesh.
Supergloo watches for installs and synchronizes the managed installations
with the desired configuration in the install object.

Updating the configuration of an install object will cause supergloo to 
modify the corresponding mesh.


`,
	}

	flagutils.AddMetadataFlags(cmd.PersistentFlags(), &opts.Metadata)
	flagutils.AddOutputFlag(cmd.PersistentFlags(), &opts.OutputType)
	flagutils.AddInteractiveFlag(cmd.PersistentFlags(), &opts.Interactive)
	flagutils.AddInstallFlags(cmd.PersistentFlags(), &opts.Install)

	cmd.AddCommand(installIstioCmd(opts))
	return cmd
}
