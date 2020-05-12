package deregister

import (
	"context"
	"fmt"
	"io"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	cluster_internal "github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/internal"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/cluster/register"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DeregistrationSet = wire.NewSet(
		ClusterDeregistrationCmd,
	)
	ErrorGettingKubeCluster = func(remoteClusterName string, err error) error {
		return eris.Errorf("Error retrieving KubernetesCluster object %s.%s, %s",
			remoteClusterName,
			env.GetWriteNamespace(),
			err.Error())
	}
	DeregisterNotPermitted = func(remoteClusterName string) error {
		return eris.Errorf("Cannot deregister cluster %s which was not manually registered through meshctl.", remoteClusterName)
	}
)

type DeregistrationCmd *cobra.Command

func ClusterDeregistrationCmd(
	ctx context.Context,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	opts *options.Options,
	kubeLoader kubeconfig.KubeLoader,
	out io.Writer,
) DeregistrationCmd {
	register := &cobra.Command{
		Use:   cliconstants.ClusterDeregisterCommand.Use,
		Short: cliconstants.ClusterDeregisterCommand.Short,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := deregisterCluster(ctx, kubeClientsFactory, clientsFactory, opts, kubeLoader)
			if err != nil {
				fmt.Fprintf(out, "Error deregistering cluster %s.\n", opts.Cluster.Deregister.RemoteClusterName)
			} else {
				fmt.Fprintf(out, "Successfully deregistered cluster %s.\n", opts.Cluster.Deregister.RemoteClusterName)
			}
			return err
		},
	}
	options.AddClusterDeregisterFlags(register, opts)
	return register
}

func deregisterCluster(
	ctx context.Context,
	kubeClientsFactory common.KubeClientsFactory,
	clientsFactory common.ClientsFactory,
	opts *options.Options,
	kubeLoader kubeconfig.KubeLoader,
) error {
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
	if err != nil {
		return ErrorGettingKubeCluster(opts.Cluster.Deregister.RemoteClusterName, err)
	}
	discoveredBy, ok := kubeCluster.GetLabels()[constants.DISCOVERED_BY]
	// Only allow manual deregistration of manually registered clusters, otherwise deregistration will compete with automated cluster discovery.
	// Deregistration of discovered clusters must happen through discovery configuration.
	if !(ok && discoveredBy == register.MeshctlDiscoverySource) {
		return DeregisterNotPermitted(opts.Cluster.Deregister.RemoteClusterName)
	}
	return masterKubeClients.ClusterDeregistrationClient.Deregister(ctx, kubeCluster)
}
