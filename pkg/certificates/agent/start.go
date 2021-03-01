package agent

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/common/schemes"

	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"

	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation"
	"github.com/solo-io/skv2/pkg/bootstrap"
)

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(ctx, StartFunc, opts, schemes.SchemeBuilder, true)
}

// start the main reconcile loop
func StartFunc(
	ctx context.Context,
	parameters bootstrap.StartParameters,
) error {

	snapshotBuilder := input.NewSingleClusterBuilder(parameters.MasterManager)

	return reconciliation.Start(
		ctx,
		snapshotBuilder,
		parameters.MasterManager,
	)
}
