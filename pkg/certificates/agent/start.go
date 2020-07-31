package agent

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/agent/input"

	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/service-mesh-hub/pkg/certificates/agent/reconciliation"
	"github.com/solo-io/skv2/pkg/multicluster"
)

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(ctx, "cert-agent", startReconciler, opts)
}

// start the main reconcile loop
func startReconciler(
	ctx context.Context,
	masterManager manager.Manager,
	_ multicluster.Client,
	_ multicluster.ClusterSet,
	_ multicluster.ClusterWatcher,
) error {

	snapshotBuilder := input.NewSingleClusterBuilder(masterManager.GetClient())

	return reconciliation.Start(
		ctx,
		snapshotBuilder,
		masterManager,
	)
}
