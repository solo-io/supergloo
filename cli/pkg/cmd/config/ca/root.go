package ca

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
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
	fmt.Println(opts.Top.Static)

	// TODO(mitchdraft) - needed for basic usage:
	// - submit crd updates

	err := ensureFlags(opts)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("wip")
	meshClient, err := common.GetMeshClient()
	if err != nil {
		fmt.Println(err)
		return
		// return err
	}
	mesh, err := (*meshClient).Read(opts.Config.Ca.Mesh.Namespace, opts.Config.Ca.Mesh.Name, clients.ReadOpts{})
	if err != nil {
		fmt.Println(err)
		return
		// return err
	}
	fmt.Println(mesh)
	fmt.Printf("Configured mesh %v to use secret %v", opts.Config.Ca.Mesh.Name, opts.Config.Ca.Secret.Name)
	return
}

func ensureFlags(opts *options.Options) error {

	oMeshRef := &(opts.Config.Ca).Mesh
	if err := nsutil.EnsureMesh(oMeshRef, opts); err != nil {
		return err
	}

	oSecretRef := &(opts.Config.Ca).Secret
	if err := nsutil.EnsureSecret(oSecretRef, opts); err != nil {
		return err
	}

	return nil
}
