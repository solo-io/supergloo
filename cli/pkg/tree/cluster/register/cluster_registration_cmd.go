package register

import (
	"io"

	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"
	"github.com/spf13/cobra"
)

func ClusterRegistrationCmd(clientsFactory common.ClientsFactory, masterClusterVerifier common.MasterKubeConfigVerifier, out io.Writer) *cobra.Command {
	cmdName := "register"

	register := &cobra.Command{
		Use:   cmdName,
		Short: "Register a new cluster by creating a service account token in that cluster through which to authorize Service Mesh Hub",
	}

	flagConfig := cluster.AddFlags(register, masterClusterVerifier)

	clusterClient := NewClusterRegistrationClient(
		out,
		flagConfig,
		clientsFactory,
	)

	register.RunE = clusterClient.RegisterCluster

	return register
}
