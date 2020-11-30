package mesh_networking

import (
	"context"

	certissuerinput "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input/issuer"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input/user"
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

	input "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input/networking"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reconciliation"
	"github.com/solo-io/skv2/pkg/multicluster"
)

type NetworkingOpts struct {
	*bootstrap.Options
	disallowIntersectionConfig bool
}

func (opts *NetworkingOpts) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&opts.disallowIntersectionConfig, "disallow-intersecting-config", false, "if enabled, Gloo Mesh will detect and report errors when outputting service mesh configuration that overlaps with existing config not managed by Gloo Mesh")
}

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts *NetworkingOpts) error {
	return bootstrap.Start(ctx, "networking", func(parameters bootstrap.StartParameters) error {
		return startReconciler(opts.disallowIntersectionConfig, parameters)
	}, *opts.Options)
}

// start the main reconcile loop
func startReconciler(
	disallowIntersectingConfig bool,
	parameters bootstrap.StartParameters,
) error {

	extensionClientset := extensions.NewClientset(parameters.Ctx)

	snapshotBuilder := input.NewSingleClusterBuilder(parameters.MasterManager)

	// contains user-supplied service mesh config read from all registered clusters
	userSuppliedSnapshotBuilder := user.NewMultiClusterBuilder(parameters.Clusters, parameters.McClient)

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
		snapshotBuilder,
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
		userSuppliedSnapshotBuilder,
		disallowIntersectingConfig,
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
