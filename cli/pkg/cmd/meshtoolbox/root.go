package meshtoolbox

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox/mtls"
	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox/policy"
	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox/routerule"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func TrafficShifting(opts *options.Options) *cobra.Command {
	cmd := generateRouteCmd("traffic-shifting", "Configure traffic shifting parameters", routerule.TrafficShifting_Rule, opts)
	routerule.AddTrafficShiftingFlags(cmd, opts)
	return cmd
}

func FaultInjection(opts *options.Options) *cobra.Command {
	cmd := generateRouteCmd("fault-injection", "Stress test your mesh with faults", routerule.FaultInjection_Rule, opts)
	routerule.AddFaultFlags(cmd, opts)
	return cmd
}

func Retries(opts *options.Options) *cobra.Command {
	cmd := generateRouteCmd("retries", "Configure retry parameters", routerule.Retries_Rule, opts)
	routerule.AddRetryFlags(cmd, opts)
	return cmd
}

func Timeout(opts *options.Options) *cobra.Command {
	cmd := generateRouteCmd("timeout", "Configure timeout parameters", routerule.Timeout_Rule, opts)
	routerule.AddTimeoutFlags(cmd, opts)
	return cmd
}

func CorsPolicy(opts *options.Options) *cobra.Command {
	cmd := generateRouteCmd("cors", "Configure cors policy parameters", routerule.CorsPolicy_Rule, opts)
	routerule.AddCorsFlags(cmd, opts)
	return cmd
}

func Mirror(opts *options.Options) *cobra.Command {
	cmd := generateRouteCmd("mirror", "Configure mirror parameters", routerule.Mirror_Rule, opts)
	routerule.AddMirrorFlags(cmd, opts)
	return cmd
}

func HeaderManipulation(opts *options.Options) *cobra.Command {
	cmd := generateRouteCmd("header-manipulation", "Configure header manipulation parameters", routerule.HeaderManipulaition_Rule, opts)
	routerule.AddHeaderManipulationFlags(cmd, opts)
	return cmd
}

func Policy(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: `Apply a policy`,
		Long:  `Apply, update, or remove a policy`,
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
		},
	}
	linkMeshToolFlags(cmd, opts)
	policy.LinkPolicyFlags(cmd, opts)
	cmd.AddCommand(
		policy.Add(opts),
		policy.Remove(opts),
		policy.Clear(opts),
	)
	return cmd
}

func ToggleMtls(opts *options.Options) *cobra.Command {
	cmd := mtls.Root(opts)
	linkMeshToolFlags(cmd, opts)
	return cmd
}

func linkMeshToolFlags(cmd *cobra.Command, opts *options.Options) {
	meshRef := &(opts.MeshTool).Mesh
	pflags := cmd.PersistentFlags()
	pflags.StringVar(&meshRef.Name, "mesh.name", "", "name of mesh to update")
	pflags.StringVar(&meshRef.Namespace, "mesh.namespace", "", "namespace of mesh to update")
	pflags.StringVar(&opts.MeshTool.ServiceId, "serviceid", "", "service to modify")
}

func meshToolPlaceholder(opts *options.Options) {
	fmt.Println("this mesh feature will be available in 2019")
}

func generateRouteCmd(useString string, description string, ruleTypeID string, opts *options.Options) *cobra.Command {
	rrOpts := &(opts.Create).InputRoutingRule
	cmd := &cobra.Command{
		Use:   useString,
		Short: description,
		Long:  description,
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			rrOpts.RouteName = args[0]
			if err := routerule.CreateRoutingRule(ruleTypeID, opts); err != nil {
				return err
			}
			fmt.Printf("Created %v routing rule [%v] in namespace [%v]\n", routerule.RoutingRuleDisplayName[ruleTypeID], args[0], rrOpts.TargetMesh.Namespace)
			return nil
		},
	}
	linkMeshToolFlags(cmd, opts)
	return cmd
}
