package demo_init

import (
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/spf13/cobra"
)

type InitCmd *cobra.Command

const (
	IstioDemoProfileName      = "istio-multicluster"
	AppmeshEksDemoProfileName = "appmesh-eks"
)

var (
	InitSet = wire.NewSet(
		DemoInitCmd,
	)
	Profiles = []string{
		IstioDemoProfileName,
		AppmeshEksDemoProfileName,
	}
)

func DemoInitCmd(
	opts *options.Options,
	runner exec.Runner,
) InitCmd {
	init := &cobra.Command{
		Use:   cliconstants.DemoInitCommand.Use,
		Short: cliconstants.DemoInitCommand.Short,
		Long:  cliconstants.DemoInitCommand.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch opts.Demo.Profile {
			case IstioDemoProfileName:
				return IstioMulticlusterDemo(runner)
			case AppmeshEksDemoProfileName:
				return AppmeshEksDemo(runner, opts.Demo.AwsRegion)
			default:
				return eris.Errorf("Invalid profile name, must be one of [%s]", strings.Join(Profiles, ", "))
			}
		},
	}
	options.AddDemoFlags(init, opts, Profiles)
	// Silence verbose error message for non-zero exit codes.
	init.SilenceUsage = true
	return init
}
