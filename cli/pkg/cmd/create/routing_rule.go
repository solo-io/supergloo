package create

import (
	"fmt"
	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox"

	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/supergloo/cli/pkg/cmd/meshtoolbox/routerule"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"

	"github.com/spf13/cobra"
)

func RoutingRuleCmd(opts *options.Options) *cobra.Command {
	rrOpts := &(opts.Create).InputRoutingRule
	cmd := &cobra.Command{
		Use:   "routingrule",
		Short: `Create a route rule with the given name`,
		Long:  `Create a route rule with the given name`,
		Args:  common.RequiredNameArg,
		RunE: func(c *cobra.Command, args []string) error {
			if err := meshtoolbox.EnsureName(rrOpts, args); err != nil {
				return err
			}
			if err := routerule.CreateRoutingRule(routerule.USE_ALL_ROUTING_RULES, opts); err != nil {
				return err
			}
			fmt.Printf("Created routing rule [%v] in namespace [%v]\n", rrOpts.RouteName, rrOpts.TargetMesh.Namespace)
			return nil
		},
	}

	cmd.SetUsageTemplate(common.UsageTemplate("rule-name"))

	routerule.AddBaseFlags(cmd, opts)
	routerule.AddTrafficShiftingFlags(cmd, opts)
	routerule.AddTimeoutFlags(cmd, opts)
	routerule.AddRetryFlags(cmd, opts)
	routerule.AddFaultFlags(cmd, opts)
	routerule.AddCorsFlags(cmd, opts)
	routerule.AddMirrorFlags(cmd, opts)
	routerule.AddHeaderManipulationFlags(cmd, opts)

	return cmd
}
