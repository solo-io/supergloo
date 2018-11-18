package ingresstoolbox

import (
	"fmt"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
)

func FortifyIngress(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fortify-ingress",
		Short: `configure ingress security parameters`,
		Long:  `configure ingress security parameters`,
		Run: func(c *cobra.Command, args []string) {
			ingressToolPlaceholder(opts)
		},
	}
	linkIngressToolFlags(cmd, opts)
	return cmd
}

func AddRoute(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-route",
		Short: `define new route`,
		Long:  `define new route`,
		Run: func(c *cobra.Command, args []string) {
			ingressToolPlaceholder(opts)
		},
	}
	linkIngressToolFlags(cmd, opts)
	return cmd
}

func ingressToolPlaceholder(opts *options.Options) {
	fmt.Println("this ingress feature will be available in 2019")
}

func linkIngressToolFlags(cmd *cobra.Command, opts *options.Options) {
	pflags := cmd.PersistentFlags()
	pflags.StringVar(&opts.IngressTool.IngressId, "ingressid", "", "ingress to modify")
	pflags.StringVar(&opts.IngressTool.RouteId, "routeid", "", "route to modify")
}
