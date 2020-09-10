package mesh_networking

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/appmesh"

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

	snapshotBuilder := input.NewSingleClusterBuilder(parameters.MasterManager)
	reporter := reporting.NewPanickingReporter(parameters.Ctx)
	translator := translation.NewTranslator(
		istio.NewIstioTranslator(),
		appmesh.NewAppmeshTranslator(),
		osm.NewOSMTranslator(),
	)
	applier := apply.NewApplier(translator)

	startCertIssuer(
		parameters.Ctx,
		parameters.MasterManager,
		parameters.McClient,
		parameters.Clusters,
	)

	return reconciliation.Start(
		parameters.Ctx,
		snapshotBuilder,
		applier,
		reporter,
		translator,
		parameters.McClient,
		parameters.MasterManager,
		parameters.SnapshotHistory,
		parameters.VerboseMode,
	)
}

func startCertIssuer(
	ctx context.Context,
	masterManager manager.Manager,
	mcClient multicluster.Client,
	clusters multicluster.Interface,
) {

	builder := certissuerinput.NewMultiClusterBuilder(
		clusters,
		mcClient,
	)

	certissuerreconciliation.Start(
		ctx,
		builder,
		mcClient,
		clusters,
		masterManager.GetClient(),
	)
}
