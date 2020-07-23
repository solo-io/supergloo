package mesh_discovery

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation"
	"github.com/solo-io/skv2/pkg/multicluster"
)

// the mesh-discovery controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(ctx, "discovery", startReconciler, opts)
}

// start the main reconcile loop
func startReconciler(
	ctx context.Context,
	masterManager manager.Manager,
	mcClient multicluster.Client,
	clusters multicluster.ClusterSet,
	mcWatcher multicluster.ClusterWatcher,
) error {
	snapshotBuilder := input.NewMultiClusterBuilder(clusters, mcClient)
	translator := translation.NewTranslator()
	reconciliation.Start(ctx, snapshotBuilder, translator, masterManager.GetClient(), mcWatcher)
	return nil
}
