package istio_multicluster

import (
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/spf13/cobra"
)

type CleanupCmd *cobra.Command

func Cleanup(
	runner exec.Runner,
	opts *options.Options,
) CleanupCmd {
	init := &cobra.Command{
		Use:   cliconstants.AppmeshEksCleanupCommand.Use,
		Short: cliconstants.AppmeshEksCleanupCommand.Short,
		Long:  cliconstants.AppmeshEksCleanupCommand.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return istioMulticlusterCleanup(runner)
		},
	}
	options.AddAppmeshEksCleanupFlags(init, opts)
	// Silence verbose error message for non-zero exit codes.
	init.SilenceUsage = true
	return init
}

func istioMulticlusterCleanup(runner exec.Runner) error {
	return runner.Run("bash", istioMulticlusterCleanupScript)
}

const (
	istioMulticlusterCleanupScript = `
kind get clusters | grep -E  '(management-plane|remote-cluster)-[a-z0-9]+' | while read -r r; do kind delete cluster --name "$r"; done
exit 0
`
)
