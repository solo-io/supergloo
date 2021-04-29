package reconciliation

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/issuer/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	v1sets "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/translation"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/translation/secret"
	"github.com/solo-io/go-utils/contextutils"
	skinput "github.com/solo-io/skv2/contrib/pkg/input"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// function which defines how the cert issuer reconciler should be registered with internal components.
type RegisterReconcilerFunc func(
	ctx context.Context,
	reconcile skinput.MultiClusterReconcileFunc,
	reconcileInterval time.Duration,
) error

// function which defines how the cert issuer should update the statuses of objects in its input snapshot
type SyncStatusFunc func(ctx context.Context, snapshot input.Snapshot) error

type certIssuerReconciler struct {
	ctx               context.Context
	builder           input.Builder
	syncInputStatuses SyncStatusFunc
	translator        translation.Translator
}

func Start(
	ctx context.Context,
	registerReconciler RegisterReconcilerFunc,
	builder input.Builder,
	syncInputStatuses SyncStatusFunc,
	translatorExtensionFunc translation.TranslationExtensionFunc,
	masterManager manager.Manager,
) error {
	translators := []translation.Translator{
		secret.NewSecretTranslator(corev1.NewSecretClient(masterManager.GetClient())),
	}
	if translatorExtensionFunc != nil {
		extensionTranslators, err := translatorExtensionFunc()
		if err != nil {
			return err
		}
		translators = append(translators, extensionTranslators...)
	}

	r := &certIssuerReconciler{
		ctx:               ctx,
		builder:           builder,
		syncInputStatuses: syncInputStatuses,
		translator:        translation.NewChainTranslator(translators...),
	}

	return registerReconciler(ctx, r.reconcile, time.Second/2)
}

// reconcile global state
func (r *certIssuerReconciler) reconcile(_ ezkube.ClusterResourceId) (bool, error) {
	inputSnap, err := r.builder.BuildSnapshot(r.ctx, "cert-issuer", input.BuildOptions{})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	for _, certificateRequest := range inputSnap.CertificateRequests().List() {
		if err := r.reconcileCertificateRequest(certificateRequest, inputSnap.IssuedCertificates()); err != nil {
			contextutils.LoggerFrom(r.ctx).Warnf("certificate request could not be processed: %v", err)
			certificateRequest.Status.Error = err.Error()
			certificateRequest.Status.State = v1.CertificateRequestStatus_FAILED
		}
	}

	return false, r.syncInputStatuses(r.ctx, inputSnap)
}

func (r *certIssuerReconciler) reconcileCertificateRequest(certificateRequest *v1.CertificateRequest, issuedCertificates v1sets.IssuedCertificateSet) error {
	// if observed generation is out of sync, treat the issued certificate as Pending (spec has been modified)
	if certificateRequest.Status.ObservedGeneration != certificateRequest.Generation {
		certificateRequest.Status.State = v1.CertificateRequestStatus_PENDING
	}

	// reset & update status
	certificateRequest.Status.ObservedGeneration = certificateRequest.Generation
	certificateRequest.Status.Error = ""

	switch certificateRequest.Status.State {
	case v1.CertificateRequestStatus_FINISHED:
		if len(certificateRequest.Status.SignedCertificate) > 0 {
			contextutils.LoggerFrom(r.ctx).Debugf("skipping cert request %v which has already been fulfilled", sets.Key(certificateRequest))
			return nil
		}
		// else treat as pending
		fallthrough
	case v1.CertificateRequestStatus_FAILED:
		// restart the workflow from PENDING
		fallthrough
	case v1.CertificateRequestStatus_PENDING:
		//
	default:
		return eris.Errorf("unknown certificate request state: %v", certificateRequest.Status.State)
	}

	issuedCertificate, err := issuedCertificates.Find(certificateRequest)
	if err != nil {
		return eris.Wrapf(err, "failed to find issued certificate matching certificate request")
	}

	output, err := r.translator.Translate(r.ctx, certificateRequest, issuedCertificate)
	if err != nil {
		return eris.Wrapf(err, "failed to translate certificate request + issued certificate")
	}

	certificateRequest.Status = v1.CertificateRequestStatus{
		ObservedGeneration: certificateRequest.Generation,
		State:              v1.CertificateRequestStatus_FINISHED,
		SignedCertificate:  output.SignedCertificate,
		SigningRootCa:      output.SigningRootCa,
	}

	return nil
}
