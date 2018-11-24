package policy

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	"github.com/spf13/cobra"
)

func Add(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: `apply a policy`,
		Long:  `apply a policy`,
		Run: func(c *cobra.Command, args []string) {
			if err := addPolicy(opts); err != nil {
				fmt.Println(err)
				return
				// return err
			}
		},
	}
	return cmd
}

func Remove(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: `remove a single policy`,
		Long:  `remove a single policy`,
		Run: func(c *cobra.Command, args []string) {
			if err := removePolicy(opts); err != nil {
				fmt.Println(err)
				return
				// return err
			}
		},
	}
	return cmd
}

func Clear(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: `clear all policies`,
		Long:  `clear all policies`,
		Run: func(c *cobra.Command, args []string) {
			if err := clearPolicies(opts); err != nil {
				fmt.Println(err)
				return
				// return err
			}
		},
	}
	return cmd
}

func LinkPolicyFlags(cmd *cobra.Command, opts *options.Options) {
	sOp := &(opts.MeshTool.AddPolicy).Source
	dOp := &(opts.MeshTool.AddPolicy).Destination
	pflags := cmd.PersistentFlags()
	pflags.StringVar(&dOp.Name, "source.name", "", "name of policy source")
	pflags.StringVar(&dOp.Namespace, "source.namespace", "", "namespace of policy source")
	pflags.StringVar(&sOp.Name, "destination.name", "", "name of policy destination")
	pflags.StringVar(&sOp.Namespace, "destination.namespace", "", "namespace of policy destination")
}

func addPolicy(opts *options.Options) error {
	if err := ensureCommonPolicyFlags(opts); err != nil {
		return err
	}
	fmt.Println("ap wip")
	return nil
}

func removePolicy(opts *options.Options) error {
	if err := ensureCommonPolicyFlags(opts); err != nil {
		return err
	}
	fmt.Println("rp wip")
	return nil
}

func clearPolicies(opts *options.Options) error {
	if err := ensureCommonPolicyFlags(opts); err != nil {
		return err
	}
	fmt.Println("cp wip")
	return nil
}

func ensureCommonPolicyFlags(opts *options.Options) error {
	meshRef := &(opts.MeshTool).Mesh
	sOp := &(opts.MeshTool.AddPolicy).Source
	dOp := &(opts.MeshTool.AddPolicy).Destination
	if err := nsutil.EnsureMesh(meshRef, opts); err != nil {
		return err
	}
	if err := nsutil.EnsureCommonResource("upstream", "policy source", sOp, opts); err != nil {
		return err
	}
	if err := nsutil.EnsureCommonResource("upstream", "policy destination", dOp, opts); err != nil {
		return err
	}
	return nil
}
