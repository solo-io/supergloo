package policy

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"
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

	if err := updatePolicy("add", opts); err != nil {
		return err
	}
	fmt.Printf("Added policy to mesh %v", opts.MeshTool.Mesh.Name)

	return nil
}

func removePolicy(opts *options.Options) error {
	// 	fmt.Println(`This function is not implemented yet.
	// For now, you can use "supergloo policy clear" to delete all of your policies.
	// If this is a feature you would like to see expedited, please let us know.
	// Thank you!`)
	if err := updatePolicy("remove", opts); err != nil {
		return err
	}
	fmt.Printf("Removed policy from mesh %v", opts.MeshTool.Mesh.Name)
	return nil
}

func clearPolicies(opts *options.Options) error {
	if err := updatePolicy("clear", opts); err != nil {
		return err
	}
	fmt.Printf("Cleared policies from mesh %v", opts.MeshTool.Mesh.Name)
	return nil
}

// Ensure that all the needed user-specified values have been provided
func ensureCommonPolicyFlags(operation string, opts *options.Options) error {

	// all operations require a target mesh spec
	meshRef := &(opts.MeshTool).Mesh
	if err := nsutil.EnsureMesh(meshRef, opts); err != nil {
		return err
	}

	// only the add and remove operations require rule specs
	if operation == "add" || operation == "remove" {
		sOp := &(opts.MeshTool.AddPolicy).Source
		dOp := &(opts.MeshTool.AddPolicy).Destination
		// TODO(mitchdraft) remove should only show/allow selection of the policies that are active
		if err := nsutil.EnsureCommonResource("upstream", "policy source", sOp, opts); err != nil {
			return err
		}
		if err := nsutil.EnsureCommonResource("upstream", "policy destination", dOp, opts); err != nil {
			return err
		}
	}

	return nil
}

func updatePolicy(operation string, opts *options.Options) error {
	// 1. validate/aquire arguments
	if err := ensureCommonPolicyFlags(operation, opts); err != nil {
		return err
	}

	// 2. read the existing mesh
	meshClient, err := common.GetMeshClient()
	if err != nil {
		return err
	}
	meshRef := &(opts.MeshTool).Mesh
	mesh, err := (*meshClient).Read(meshRef.Namespace, meshRef.Name, clients.ReadOpts{})
	if err != nil {
		return err
	}

	sOp := &(opts.MeshTool.AddPolicy).Source
	dOp := &(opts.MeshTool.AddPolicy).Destination

	// 3. mutate the mesh structure
	switch operation {
	case "add":
		// Note: this does not check for duplicate policies
		newRule := &superglooV1.Rule{
			Source: &core.ResourceRef{
				Name:      sOp.Name,
				Namespace: sOp.Namespace,
			},
			Destination: &core.ResourceRef{
				Name:      dOp.Name,
				Namespace: dOp.Namespace,
			},
		}

		if mesh.Policy == nil {
			mesh.Policy = &superglooV1.Policy{
				Rules: []*superglooV1.Rule{newRule},
			}
		} else {
			mesh.Policy.Rules = append(mesh.Policy.Rules, newRule)

		}
	case "remove":
		// if there are no rules to begin with, we have nothing to do
		if mesh.Policy != nil || mesh.Policy.Rules != nil {
			return fmt.Errorf("There are no policy rules to remove.")
		}
		newRules := []*superglooV1.Rule{}
		for _, rule := range mesh.Policy.Rules {
			if rule.Source.Name != sOp.Name &&
				rule.Source.Namespace != sOp.Namespace &&
				rule.Destination.Name != dOp.Name &&
				rule.Destination.Namespace != dOp.Namespace {
				newRules = append(newRules, rule)
			}
		}
		mesh.Policy.Rules = newRules
	case "clear":
		mesh.Policy = &superglooV1.Policy{}
	default:
		panic(fmt.Errorf("Operation %v not recognized", operation))
	}

	// 4. write the changes
	_, err = (*meshClient).Write(mesh, clients.WriteOpts{OverwriteExisting: true})
	if err != nil {
		return err
	}
	return nil
}
