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

	flags.StringVar(&cOpts.Mesh.Name, "mesh.namespace", "", "namespace of mesh to update")

	flags.StringVar(&cOpts.Secret.RootCa, "rootca", "", "filename of rootca for secret")

	flags.StringVar(&cOpts.Secret.PrivateKey, "privatekey", "", "filename of privatekey for secret")

	flags.StringVar(&cOpts.Secret.CertChain, "certchain", "", "filename of certchain for secret")

	flags.StringVar(&cOpts.Secret.Namespace, "secretnamespace", "", "namespace in which to store the secret")

	return cmd
}

func configureCa(opts *options.Options) {
	err := ensureFlags(opts)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("wip")
	return
}

func ensureFlags(opts *options.Options) error {
	oMeshRef := &(opts.Config.Ca).Mesh
	// TODO(mitchdraft) - this block essentially gets a mesh resource ref if your cmd needs one
	// It is very similar to logic in create/routing_rule.go
	// Abstract so that it operates on a pointer to a Mesh ResourceRef, providing validation and optional interactive selection mode
	if oMeshRef.Name == "" {
		// Q(mitchdraft) do we want to prefilter this by namespace if they have chosen one?
		meshName, namespace, err := nsutil.ChooseMesh(opts.Cache.NsResources)
		if err != nil {
			return fmt.Errorf("input error")
		}
		oMeshRef.Name = meshName
		oMeshRef.Namespace = namespace
	} else {
		if !common.Contains(opts.Cache.NsResources[oMeshRef.Namespace].Meshes, oMeshRef.Name) {
			return fmt.Errorf("Please specify a valid mesh name. Mesh %v not found in namespace %v not found", oMeshRef.Name, oMeshRef.Namespace)
		}
	}

	return nil
}
