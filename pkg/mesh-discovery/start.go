package mesh_discovery

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation"
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
	snapshotBuilder := input.NewMultiClusterBuilder(parameters.Clusters, parameters.McClient)
	translator := translation.NewTranslator(translation.DefaultDependencyFactory)
	reconciliation.Start(
		parameters.Ctx,
		snapshotBuilder,
		translator,
		parameters.MasterManager.GetClient(),
		parameters.Clusters,
		parameters.SnapshotHistory,
		parameters.VerboseMode,
	)
	return nil
}
