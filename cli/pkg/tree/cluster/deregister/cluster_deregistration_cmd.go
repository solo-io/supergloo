package deregister

import (
	"context"
	"fmt"
	"io"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	cluster_internal "github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/internal"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeregistrationCmd *cobra.Command

var DeregistrationSet = wire.NewSet(
	ClusterDeregistrationCmd,
)

func ClusterDeregistrationCmd(
	ctx context.Context,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	opts *options.Options,
	kubeLoader kubeconfig.KubeLoader,
	out io.Writer,
) DeregistrationCmd {
	register := &cobra.Command{
		Use:   cliconstants.ClusterRegisterCommand.Use,
		Short: cliconstants.ClusterRegisterCommand.Short,
		Long:  cliconstants.ClusterRegisterCommand.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cluster_internal.VerifyMasterCluster(clientsFactory, opts); err != nil {
				return err
			}
			masterCfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
			if err != nil {
				return err
			}
			masterKubeClients, err := kubeClientsFactory(masterCfg, opts.Root.WriteNamespace)
			if err != nil {
				return err
			}
			kubeCluster, err := masterKubeClients.KubeClusterClient.GetKubernetesCluster(
				ctx,
				client.ObjectKey{Name: opts.Cluster.Deregister.RemoteClusterName, Namespace: env.GetWriteNamespace()})
			err = masterKubeClients.ClusterDeregistrationClient.Deregister(ctx, kubeCluster)
			if err != nil {
				fmt.Fprintf(out, "Error deregistering cluster %s: %+v", opts.Cluster.Deregister.RemoteClusterName, err)
			} else {
				fmt.Fprintf(out, "Successfully deregistered cluster %s.", opts.Cluster.Deregister.RemoteClusterName)
			}
			return err
		},
	}
	options.AddClusterDeregisterFlags(register, opts)
	return register
}
