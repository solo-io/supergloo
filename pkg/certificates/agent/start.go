package agent

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/common/schemes"

	corev1client "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"

	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation"
	pod_bouncer "github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation/internal/pod-bouncer"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/translation"
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

	podCLient := corev1client.NewPodClient(parameters.MasterManager.GetClient())
	podBouncer := pod_bouncer.NewPodBouncer(podCLient)
	translator := translation.NewCertAgentTranslator(parameters.MasterManager.GetClient())
	return reconciliation.Start(
		ctx,
		snapshotBuilder,
		parameters.MasterManager,
		podBouncer,
		translator,
	)
}
