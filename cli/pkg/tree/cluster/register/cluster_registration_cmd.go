package register

import (
	"context"
	"fmt"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/spf13/cobra"
)

type RegistrationCmd *cobra.Command

var RegistrationSet = wire.NewSet(
	ClusterRegistrationCmd,
)

func ClusterRegistrationCmd(
	ctx context.Context,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	opts *options.Options,
	kubeLoader kubeconfig.KubeLoader,
	out io.Writer,
) RegistrationCmd {
	register := &cobra.Command{
		Use:   cliconstants.ClusterRegisterCommand.Use,
		Short: cliconstants.ClusterRegisterCommand.Short,
		Long:  cliconstants.ClusterRegisterCommand.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RegisterCluster(
				ctx,
				kubeClientsFactory,
				clientsFactory,
				opts,
				kubeLoader,
			)
			if err != nil {
				fmt.Fprintf(out, "Error registering cluster %s.\n", opts.Cluster.Register.RemoteClusterName)
			} else {
				fmt.Fprintf(out, "Successfully registered cluster %s.\n", opts.Cluster.Register.RemoteClusterName)
			}
			return err
		},
	}
	options.AddClusterRegisterFlags(register, opts)
	return register
}
