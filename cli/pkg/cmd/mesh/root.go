package mesh

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mesh",
		Aliases: []string{"m"},
		Short:   "configure a service mesh using Supergloo",
		Long: `Modifies Mesh resources which the supergloo controller
uses to configure service meshes.

Meshes represent the desired state of a supported mesh.
Supergloo watches for Meshes and synchronizes the managed installations
with the desired configuration in the Mesh object.

`,
	}

	//cmd.AddCommand(configureCommand(opts))
	return cmd
}
