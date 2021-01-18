package mesh_discovery

import (
	"context"

	"github.com/spf13/pflag"

	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/reconciliation"
)

type DiscoveryOpts struct {
	*bootstrap.Options
	agentCluster string
}

func (opts *DiscoveryOpts) AddToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.agentCluster, "agent-cluster", "", "If set, Discovery will run in *agent mode*, in which discovery resources will only be generated for the local cluster. Agent mode does not require the local cluster to be registered. The value of --agent-cluster should be equal to the name of the cluster as it was registered with Gloo Mesh.")
}

// the mesh-discovery controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts DiscoveryOpts) error {
	return bootstrap.Start(ctx, "discovery", func(parameters bootstrap.StartParameters) error {
		return startReconciler(opts.agentCluster, parameters)
	}, *opts.Options, opts.agentCluster != "")
}

// start the main reconcile loop
func startReconciler(
	agentCluster string,
	parameters bootstrap.StartParameters,
) error {
	return reconciliation.Start(
		parameters.Ctx,
		agentCluster,
		parameters.MasterManager,
		parameters.Clusters,
		parameters.McClient,
		parameters.SnapshotHistory,
		parameters.VerboseMode,
		&parameters.SettingsRef,
	)
}
