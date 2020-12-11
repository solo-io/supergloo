package mesh_discovery

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation"
)

// the mesh-discovery controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(ctx, "discovery", startReconciler, opts)
}

// start the main reconcile loop
func startReconciler(
	parameters bootstrap.StartParameters,
) error {
	remoteSnapshotBuilder := input.NewMultiClusterRemoteBuilder(parameters.Clusters, parameters.McClient)
	localSnapshotBuilder := input.NewSingleClusterLocalBuilder(parameters.MasterManager)
	translator := translation.NewTranslator(translation.DefaultDependencyFactory)
	return reconciliation.Start(
		parameters.Ctx,
		remoteSnapshotBuilder,
		localSnapshotBuilder,
		translator,
		parameters.MasterManager,
		parameters.Clusters,
		parameters.SnapshotHistory,
		parameters.VerboseMode,
		parameters.SettingsRef,
	)
}
