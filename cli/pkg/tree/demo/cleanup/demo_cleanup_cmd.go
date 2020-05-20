package cleanup

import (
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	demo_init "github.com/solo-io/service-mesh-hub/cli/pkg/tree/demo/init"
	"github.com/spf13/cobra"
)

type CleanupCmd *cobra.Command

var CleanupSet = wire.NewSet(
	DemoCleanupCmd,
)

func DemoCleanupCmd(
	runner exec.Runner,
	opts *options.Options,
) CleanupCmd {
	init := &cobra.Command{
		Use:   cliconstants.DemoCleanupCommand.Use,
		Short: cliconstants.DemoCleanupCommand.Short,
		Long:  cliconstants.DemoCleanupCommand.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch opts.Demo.Profile {
			case demo_init.IstioDemoProfileName:
				return IstioMulticlusterCleanup(runner)
			case demo_init.AppmeshEksDemoProfileName:
				return AppmeshEksCleanup(runner, opts.DemoCleanup.AwsRegion, opts.DemoCleanup.MeshName, opts.DemoCleanup.ClusterName)
			default:
				return eris.Errorf("Invalid profile name, must be one of [%s]", strings.Join(demo_init.Profiles, ", "))
			}
		},
	}
	options.AddDemoCleanupFlags(init, opts, demo_init.Profiles)
	return init
}
