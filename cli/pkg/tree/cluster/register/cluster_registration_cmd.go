package register

import (
	"io"

	"github.com/solo-io/mesh-projects/cli/pkg/tree/cluster"

	"github.com/solo-io/mesh-projects/cli/pkg/common"
	"github.com/spf13/cobra"
)

func ClusterRegistrationCmd(clientsFactory common.ClientsFactory, globalFlagConfig *common.GlobalFlagConfig, out io.Writer) *cobra.Command {
	cmdName := "register"

	register := &cobra.Command{
		Use:   cmdName,
		Short: "Register a new cluster by creating a service account token in that cluster through which to authorize Service Mesh Hub",
	}

	clusterFlagConfig := &cluster.FlagConfig{GlobalFlagConfig: globalFlagConfig}
	cluster.AddFlags(register, clusterFlagConfig)
	clusterClient := NewClusterRegistrationClient(out, clusterFlagConfig, clientsFactory)

	register.RunE = clusterClient.RegisterCluster

	return register
}
