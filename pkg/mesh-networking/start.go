package mesh_networking

import (
	"context"

	certissuerinput "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/issuer/input"
	certissuerreconciliation "github.com/solo-io/service-mesh-hub/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/apply"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm"
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
	parameters bootstrap.StartParameters,
) error {

	snapshotBuilder := input.NewSingleClusterBuilder(parameters.MasterManager.GetClient())
	reporter := reporting.NewPanickingReporter(parameters.Ctx)
	translator := translation.NewTranslator(
		istio.NewIstioTranslator(),
		osm.NewOSMTranslator(),
	)
	validator := apply.NewApplier(translator)

	startCertIssuer(
		parameters.Ctx,
		parameters.MasterManager,
		parameters.McClient,
		parameters.Clusters,
		parameters.ClusterWatcher,
	)

	return reconciliation.Start(
		parameters.Ctx,
		snapshotBuilder,
		validator,
		reporter,
		translator,
		parameters.McClient,
		parameters.MasterManager,
		parameters.SnapshotHistory,
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
