package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/output/certagent"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	podbouncer "github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation/pod-bouncer"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/translation"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/output/errhandlers"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/reconcile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	// all output resources are labeled to prevent
	// resource collisions & garbage collection of
	// secrets the agent doesn't own
	agentLabels = func() map[string]string {
		labels := map[string]string{
			fmt.Sprintf("agent.%v", certificatesv1.SchemeGroupVersion.Group): defaults.GetPodNamespace(),
		}
		if agentCluster := defaults.GetAgentCluster(); agentCluster != "" {
			labels[metautils.AgentLabelKey] = agentCluster
		}
		return labels
	}
)

type certAgentReconciler struct {
	ctx         context.Context
	builder     input.Builder
	localClient client.Client
	podBouncer  podbouncer.PodBouncer
	translator  translation.Translator
}

func Start(
	ctx context.Context,
	builder input.Builder,
	mgr manager.Manager,
	podBouncer podbouncer.PodBouncer,
	translator translation.Translator,
) error {
	ctx = contextutils.WithLogger(ctx, "cert-agent")
	d := &certAgentReconciler{
		ctx:         ctx,
		builder:     builder,
		localClient: mgr.GetClient(),
		podBouncer:  podBouncer,
		translator:  translator,
	}
	_, err := input.RegisterSingleClusterReconciler(ctx, mgr, d.reconcile, time.Second/2, reconcile.Options{})
	return err
}

// reconcile global state
func (r *certAgentReconciler) reconcile(_ ezkube.ResourceId) (bool, error) {
	inputSnap, err := r.builder.BuildSnapshot(r.ctx, "cert-agent", input.BuildOptions{})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	outputs := certagent.NewBuilder(r.ctx, "agent")

	// process issued certificates
	for _, issuedCertificate := range inputSnap.IssuedCertificates().List() {
		if err := r.reconcileIssuedCertificate(
			issuedCertificate,
			inputSnap,
			outputs,
		); err != nil {
			issuedCertificate.Status.Error = err.Error()
			issuedCertificate.Status.State = certificatesv1.IssuedCertificateStatus_FAILED
		}
	}
	outSnap, err := outputs.BuildSinglePartitionedSnapshot(agentLabels())
	if err != nil {
		return false, err
	}

	errHandler := errhandlers.AppendingErrHandler{}
	outSnap.ApplyLocalCluster(r.ctx, r.localClient, errHandler)

	errs := errHandler.Errors()
	if err := inputSnap.SyncStatuses(r.ctx, r.localClient, input.SyncStatusOptions{
		IssuedCertificate:  true,
		PodBounceDirective: true,
	}); err != nil {
		errs = multierror.Append(errs, err)
	}

	return false, errs
}

// Exposed for testing
func (r *certAgentReconciler) reconcileIssuedCertificate(
	issuedCertificate *certificatesv1.IssuedCertificate,
	inputSnap input.Snapshot,
	outputs certagent.Builder,
) error {

	// If the IssuedCertificate's IssuedCertificateSecret is nil, the cert-agent is not responsible for issuing the certificate.
	if !r.translator.ShouldProcess(r.ctx, issuedCertificate) {
		return nil
	}

	// if observed generation is out of sync, treat the issued certificate as Pending (spec has been modified)
	if issuedCertificate.Status.ObservedGeneration != issuedCertificate.Generation {
		issuedCertificate.Status.State = certificatesv1.IssuedCertificateStatus_PENDING
		// Also need to reset PBD status
		if issuedCertificate.Spec.PodBounceDirective != nil {
			podBounceDirective, err := inputSnap.PodBounceDirectives().Find(issuedCertificate.Spec.PodBounceDirective)
			if err != nil {
				return eris.Wrap(err, "failed to find specified pod bounce directive")
			}
			podBounceDirective.Status = certificatesv1.PodBounceDirectiveStatus{}
		}
	}

	// reset & update status
	issuedCertificate.Status.ObservedGeneration = issuedCertificate.Generation
	issuedCertificate.Status.Error = ""

	// state-machine style processor
	switch issuedCertificate.Status.State {
	case certificatesv1.IssuedCertificateStatus_FINISHED:

		// If the translator errors, set the Status to failed so we can restart the workflow
		if err := r.translator.IssuedCertificateFinished(r.ctx, issuedCertificate, inputSnap, outputs); err != nil {
			issuedCertificate.Status.State = certificatesv1.IssuedCertificateStatus_FAILED
			issuedCertificate.Status.Error = err.Error()
		}

	case certificatesv1.IssuedCertificateStatus_FAILED:
		// restart the workflow from PENDING
		fallthrough
	case certificatesv1.IssuedCertificateStatus_PENDING:

		csrBytes, err := r.translator.IssuedCertiticatePending(r.ctx, issuedCertificate, inputSnap, outputs)
		if err != nil {
			return err
		}

		certificateRequest := &certificatesv1.CertificateRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuedCertificate.Name,
				Namespace: issuedCertificate.Namespace,
				Labels:    agentLabels(),
			},
			Spec: certificatesv1.CertificateRequestSpec{
				CertificateSigningRequest: csrBytes,
			},
		}
		outputs.AddCertificateRequests(certificateRequest)

		// set status to REQUESTED
		issuedCertificate.Status.State = certificatesv1.IssuedCertificateStatus_REQUESTED
	case certificatesv1.IssuedCertificateStatus_REQUESTED:

		// retrieve signed certificate
		certificateRequest, err := inputSnap.CertificateRequests().Find(issuedCertificate)
		if err != nil {
			return err
		}

		wait, err := r.translator.IssuedCertificateRequested(
			r.ctx,
			issuedCertificate,
			certificateRequest,
			inputSnap,
			outputs,
		)
		if err != nil {
			return err
		} else if wait {
			// If the requested translator signals us to wait, we need to hang on
			return nil
		}

		issuedCertificate.Status.State = certificatesv1.IssuedCertificateStatus_ISSUED
	case certificatesv1.IssuedCertificateStatus_ISSUED:

		if err := r.translator.IssuedCertificateIssued(r.ctx, issuedCertificate, inputSnap, outputs); err != nil {
			return err
		}

		// see if we need to bounce pods
		if issuedCertificate.Spec.PodBounceDirective != nil {
			podBounceDirective, err := inputSnap.PodBounceDirectives().Find(issuedCertificate.Spec.PodBounceDirective)
			if err != nil {
				return eris.Wrap(err, "failed to find specified pod bounce directive")
			}

			// try to bounce the pods and see if we need to wait
			waitForConditions, err := r.podBouncer.BouncePods(
				r.ctx,
				podBounceDirective,
				inputSnap.Pods(),
				inputSnap.ConfigMaps(),
				inputSnap.Secrets(),
			)
			if err != nil {
				return eris.Wrap(err, "bouncing pods")
			}

			if waitForConditions {
				// return here without updating the status of the issued Certificate; we want to retry bouncing the pods
				// when replacements are live and data plane certs are distributed
				return nil
			}
		}

		// mark issued certificate as finished
		issuedCertificate.Status.State = certificatesv1.IssuedCertificateStatus_FINISHED
	default:
		return eris.Errorf("unknown issued certificate state: %v", issuedCertificate.Status.State)
	}

	return nil
}
