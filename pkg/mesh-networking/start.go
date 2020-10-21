package mesh_networking

import (
	"context"

	certissuerinput "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/issuer/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/settings.smh.solo.io/v1alpha2"
	certissuerreconciliation "github.com/solo-io/service-mesh-hub/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/service-mesh-hub/pkg/common/bootstrap"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/apply"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/extensions"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm"
	"github.com/solo-io/skv2/pkg/ezkube"
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
	settings, err := v1alpha2.NewSettingsClient(parameters.MasterManager.GetClient()).GetSettings(
		parameters.Ctx,
		ezkube.MakeClientObjectKey(&parameters.SettingsRef),
	)
	if err != nil {
		return err
	}
	extensionClients, err := extensions.NewClientsFromSettings(
		parameters.Ctx,
		settings.Spec.NetworkingExtensionServers,
	)
	if err != nil {
		return err
	}

	snapshotBuilder := input.NewSingleClusterBuilder(parameters.MasterManager)
	reporter := reporting.NewPanickingReporter(parameters.Ctx)
	translator := translation.NewTranslator(
		istio.NewIstioTranslator(extensionClients),
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
		parameters.SettingsRef,
		extensionClients,
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
