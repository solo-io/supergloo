package agent

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/translation"
	"github.com/solo-io/gloo-mesh/pkg/common/schemes"
	"github.com/solo-io/skv2/pkg/bootstrap"
)

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(
		ctx,
		StartFuncExt(func(ctx context.Context, parameters bootstrap.StartParameters) ExtensionOpts {
			return ExtensionOpts{}
		}),
		opts,
		schemes.SchemeBuilder,
		true,
	)
}

// Extended start function
func StartFuncExt(makeExtensionOpts MakeExtensionOpts) bootstrap.StartFunc {
	// start the main reconcile loop
	return func(ctx context.Context, parameters bootstrap.StartParameters) error {

		extOpts := makeExtensionOpts(ctx, parameters)
		extOpts.initDefaults(parameters)

		snapshotBuilder := input.NewSingleClusterBuilder(parameters.MasterManager)

		translator := translation.NewCertAgentTranslator(parameters.MasterManager.GetClient())
		return reconciliation.Start(
			ctx,
			snapshotBuilder,
			parameters.MasterManager,
			extOpts.CertAgentReconciler.PodBouncer,
			extOpts.CertAgentReconciler.MakeTranslator(translator),
		)
	}
}
