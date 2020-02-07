package cluster_common

import (
	"github.com/solo-io/mesh-projects/cli/pkg/options"
	"github.com/spf13/cobra"
)

const (
	remoteClusterName    = "remote-cluster-name"
	remoteWriteNamespace = "remote-write-namespace"
	remoteContext        = "remote-context"
	remoteKubeconfig     = "remote-kubeconfig"
)

func AddFlags(cmd *cobra.Command, opts *options.Options) {
	flags := cmd.PersistentFlags()

	flags.StringVar(&opts.Cluster.Register.RemoteClusterName, remoteClusterName, "",
		"Name of the cluster to be operated upon, if not set will be the same as the context name")
	flags.StringVar(&opts.Cluster.Register.RemoteWriteNamespace, remoteWriteNamespace, "default",
		"Namespace in the target cluster in which to write resources")
	flags.StringVar(&opts.Cluster.Register.RemoteContext, remoteContext, "",
		"Set the context you would like to use for the target cluster")
	flags.StringVar(&opts.Cluster.Register.RemoteKubeConfig, remoteKubeconfig, "",
		"Set the path to the kubeconfig you would like to use for the target cluster. Leave empty to use the "+
			"default.")

	cobra.MarkFlagRequired(flags, remoteClusterName)
}
