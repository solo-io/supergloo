package get

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "get information about supergloo objects",
		// 		Long: `Creates or updates Rule resources which the SuperGloo controller
		// will use to configure an installed mesh.
		//
		// This set of commands creates Kubernetes CRDs which the SuperGloo controller
		// reads asynchronously.
		//
		// To view these crds:
		//
		// kubectl get routingrule [-n supergloo-system]
		//
		// `,
	}

	cmd.AddCommand(getMeshIngressCmd(opts))

	return cmd
}
