package mesh_networking

import (
	"context"

	certissuerinput "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/issuer/input"
	certissuerreconciliation "github.com/solo-io/service-mesh-hub/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi"

	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/approval"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reconciliation"
	"github.com/solo-io/skv2/pkg/multicluster"
)

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(ctx, "networking", startReconciler, opts)
}

// start the main reconcile loop
func startReconciler(
	ctx context.Context,
	masterManager manager.Manager,
	mcClient multicluster.Client,
	clusters multicluster.ClusterSet,
	clusterWatcher multicluster.ClusterWatcher,
) error {

	snapshotBuilder := input.NewSingleClusterBuilder(masterManager.GetClient())
	reporter := reporting.NewPanickingReporter(ctx)
	translator := translation.NewTranslator(
		istio.NewIstioTranslator(),
		smi.NewIstioTranslator(smi.DefaultDependencyFactory),
	)
	validator := approval.NewApprover(translator)

	startCertIssuer(
		ctx,
		masterManager,
		mcClient,
		clusters,
		clusterWatcher,
	)

	return reconciliation.Start(
		ctx,
		snapshotBuilder,
		validator,
		reporter,
		translator,
		mcClient,
		masterManager,
	)
}

func startCertIssuer(
	ctx context.Context,
	masterManager manager.Manager,
	mcClient multicluster.Client,
	clusters multicluster.ClusterSet,
	clusterWatcher multicluster.ClusterWatcher,
) {

	builder := certissuerinput.NewMultiClusterBuilder(
		clusters,
		mcClient,
	)

	certissuerreconciliation.Start(
		ctx,
		builder,
		mcClient,
		clusterWatcher,
		masterManager.GetClient(),
	)
}
