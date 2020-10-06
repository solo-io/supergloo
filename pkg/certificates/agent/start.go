package agent

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/agent/input"

	"github.com/solo-io/service-mesh-hub/pkg/certificates/agent/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
)

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(ctx, "cert-agent", startReconciler, opts)
}

// start the main reconcile loop
func startReconciler(
	parameters bootstrap.StartParameters,
) error {

	snapshotBuilder := input.NewSingleClusterBuilder(parameters.MasterManager)

	return reconciliation.Start(
		parameters.Ctx,
		snapshotBuilder,
		parameters.MasterManager,
	)
}
