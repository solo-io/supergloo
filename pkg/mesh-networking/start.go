package mesh_networking

import (
	"context"

	certissuerinput "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input"
	certissuerreconciliation "github.com/solo-io/gloo-mesh/pkg/certificates/issuer/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/common/bootstrap"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/extensions"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/appmesh"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reconciliation"
	"github.com/solo-io/skv2/pkg/multicluster"
)

type NetworkingOpts struct {
	*bootstrap.Options
	disallowIntersectionConfig bool
	watchOutputTypes           bool
}

func (opts *NetworkingOpts) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&opts.disallowIntersectionConfig, "disallow-intersecting-config", false, "if enabled, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh")
	flags.BoolVar(&opts.watchOutputTypes, "watch-output-types", true, "if disabled, Gloo Mesh will not resync upon changes to service mesh config managed by Gloo Mesh")
}

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts *NetworkingOpts) error {
	starter := networkingStarter{
		disallowIntersectingConfig: opts.disallowIntersectionConfig,
		watchOutputTypes:           opts.watchOutputTypes,
	}

	return bootstrap.Start(ctx, "networking", starter.startReconciler, *opts.Options, false)
}

type networkingStarter struct {
	disallowIntersectingConfig bool
	watchOutputTypes           bool
}

// start the main reconcile loop
func (s networkingStarter) startReconciler(parameters bootstrap.StartParameters) error {

	extensionClientset := extensions.NewClientset(parameters.Ctx)

	localSnapshotBuilder := input.NewSingleClusterLocalBuilder(parameters.MasterManager)

	// contains output resource types read from all registered clusters
	remoteSnapshotBuilder := input.NewMultiClusterRemoteBuilder(parameters.Clusters, parameters.McClient)

	reporter := reporting.NewPanickingReporter(parameters.Ctx)
	translator := translation.NewTranslator(
		istio.NewIstioTranslator(extensionClientset),
		appmesh.NewAppmeshTranslator(),
		osm.NewOSMTranslator(),
	)

	validatingTranslator := translation.NewTranslator(
		istio.NewIstioTranslator(nil), // the applier should not call the extender
		appmesh.NewAppmeshTranslator(),
		osm.NewOSMTranslator(),
	)
	applier := apply.NewApplier(validatingTranslator)

	startCertIssuer(
		parameters.Ctx,
		parameters.MasterManager,
		parameters.McClient,
		parameters.Clusters,
	)

	return reconciliation.Start(
		parameters.Ctx,
		localSnapshotBuilder,
		remoteSnapshotBuilder,
		applier,
		reporter,
		translator,
		parameters.Clusters,
		parameters.McClient,
		parameters.MasterManager,
		parameters.SnapshotHistory,
		parameters.VerboseMode,
		parameters.SettingsRef,
		extensionClientset,
		s.disallowIntersectingConfig,
		s.watchOutputTypes,
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
