package ca

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cOpts := &(opts.Config).Ca
	cmd := &cobra.Command{
		Use:   "ca",
		Short: `Update CA`,
		Long:  `Update CA`,
		Run: func(c *cobra.Command, args []string) {
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&cOpts.Mesh.Name, "mesh.name", "", "name of mesh to update")

	flags.StringVar(&cOpts.Mesh.Name, "mesh.namespace", "", "namespace of mesh to update")

	flags.StringVar(&cOpts.Secret.RootCa, "rootca", "", "filename of rootca for secret")

	flags.StringVar(&cOpts.Secret.PrivateKey, "privatekey", "", "filename of privatekey for secret")

	flags.StringVar(&cOpts.Secret.CertChain, "certchain", "", "filename of certchain for secret")

	flags.StringVar(&cOpts.Secret.Namespace, "secretnamespace", "", "namespace in which to store the secret")

	return cmd
}
