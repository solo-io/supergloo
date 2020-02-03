package register

import (
	"io"

	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	cluster_common "github.com/solo-io/mesh-projects/cli/pkg/tree/cluster/common"
	"github.com/spf13/cobra"
)

type RegistrationCmd *cobra.Command

var RegistrationSet = wire.NewSet(
	ClusterRegistrationCmd,
)

func ClusterRegistrationCmd(kubeClientsFactory common.KubeClientsFactory, clientsFactory common.ClientsFactory, opts *options.Options, out io.Writer) RegistrationCmd {
	cmdName := "register"

	register := &cobra.Command{
		Use:   cmdName,
		Short: "Register a new cluster by creating a service account token in that cluster through which to authorize Service Mesh Hub",
		Long: "In order to specify the remote cluster against which to perform this operation, one or both of the" +
			" --remote-kubeconfig or --remote-context flags must be set. The former selects the kubeconfig file, and the latter" +
			"the context within to use.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RegisterCluster(clientsFactory, kubeClientsFactory, opts, out)
		},
	}

	cluster_common.AddFlags(register, opts)

	return register
}
