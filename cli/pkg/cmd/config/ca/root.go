package ca

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cOpts := &(opts.Config).Ca
	cmd := &cobra.Command{
		Use:   "ca",
		Short: `Update CA`,
		Long:  `Update CA`,
		Run: func(c *cobra.Command, args []string) {
			configureCa(opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&cOpts.Mesh.Name, "mesh.name", "", "name of mesh to update")

	flags.StringVar(&cOpts.Mesh.Namespace, "mesh.namespace", "", "namespace of mesh to update")

	flags.StringVar(&cOpts.Secret.Name, "secret.name", "", "name of secret to apply")

	flags.StringVar(&cOpts.Secret.Namespace, "secret.namespace", "", "namespace of secret to apply")

	return cmd
}

func configureCa(opts *options.Options) {

	// TODO(mitchdraft) - needed for basic usage:
	// - submit crd updates

	err := ensureFlags(opts)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("wip")
	fmt.Printf("Configured mesh %v to use secret %v", opts.Config.Ca.Mesh.Name, opts.Config.Ca.Secret.Name)
	return
}

func ensureFlags(opts *options.Options) error {

	oMeshRef := &(opts.Config.Ca).Mesh
	nsutil.EnsureMesh(oMeshRef, opts)

	oSecretRef := &(opts.Config.Ca).Secret
	nsutil.EnsureSecret(oSecretRef, opts)

	return nil
}
