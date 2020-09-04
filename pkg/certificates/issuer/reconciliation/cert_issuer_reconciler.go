package reconciliation

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/issuer/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/common/secrets"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/issuer/utils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type certIssuerReconciler struct {
	ctx           context.Context
	builder       input.Builder
	mcClient      multicluster.Client
	masterSecrets corev1.SecretClient
}

func Start(
	ctx context.Context,
	builder input.Builder,
	mcClient multicluster.Client,
	clusters multicluster.ClusterWatcher,
	masterClient client.Client,
) {
	r := &certIssuerReconciler{
		ctx:           ctx,
		builder:       builder,
		mcClient:      mcClient,
		masterSecrets: corev1.NewSecretClient(masterClient),
	}

	input.RegisterMultiClusterReconciler(ctx, clusters, r.reconcile, time.Second/2)
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
			certificateRequest.Status.Error = err.Error()
			certificateRequest.Status.State = v1alpha2.CertificateRequestStatus_FAILED
		}
	}

	return false, inputSnap.SyncStatusesMultiCluster(r.ctx, r.mcClient)
}

func (r *certIssuerReconciler) reconcileCertificateRequest(certificateRequest *v1alpha2.CertificateRequest, issuedCertificates v1alpha2sets.IssuedCertificateSet) error {
	// if observed generation is out of sync, treat the issued certificate as Pending (spec has been modified)
	if certificateRequest.Status.ObservedGeneration != certificateRequest.Generation {
		certificateRequest.Status.State = v1alpha2.CertificateRequestStatus_PENDING
	}

	// reset & update status
	certificateRequest.Status.ObservedGeneration = certificateRequest.Generation
	certificateRequest.Status.Error = ""

	switch certificateRequest.Status.State {
	case v1alpha2.CertificateRequestStatus_FINISHED:
		if len(certificateRequest.Status.SignedCertificate) > 0 {
			contextutils.LoggerFrom(r.ctx).Debugf("skipping cert request %v which has already been fulfilled", sets.Key(certificateRequest))
			return nil
		}
		// else treat as pending
		fallthrough
	case v1alpha2.CertificateRequestStatus_FAILED:
		// restart the workflow from PENDING
		fallthrough
	case v1alpha2.CertificateRequestStatus_PENDING:
		//
	default:
		return eris.Errorf("unknown certificate request state: %v", certificateRequest.Status.State)
	}

	issuedCertificate, err := issuedCertificates.Find(certificateRequest)
	if err != nil {
		return eris.Wrapf(err, "failed to find issued certificate matching certificate request")
	}

	signingCertificateSecret, err := r.masterSecrets.GetSecret(r.ctx, ezkube.MakeClientObjectKey(issuedCertificate.Spec.SigningCertificateSecret))
	if err != nil {
		return eris.Wrapf(err, "failed to find issuer's signing certificate matching issued request %v", sets.Key(issuedCertificate))
	}

	signingCA := secrets.RootCADataFromSecretData(signingCertificateSecret.Data)

	// generate the issued cert PEM encoded bytes
	signedCert, err := utils.GenCertForCSR(
		issuedCertificate.Spec.Hosts,
		certificateRequest.Spec.CertificateSigningRequest,
		signingCA.RootCert,
		signingCA.PrivateKey,
	)
	if err != nil {
		return eris.Wrapf(err, "failed to generate signed cert for certificate request %v", sets.Key(certificateRequest))
	}

	certificateRequest.Status = v1alpha2.CertificateRequestStatus{
		ObservedGeneration: certificateRequest.Generation,
		State:              v1alpha2.CertificateRequestStatus_FINISHED,
		SignedCertificate:  signedCert,
		SigningRootCa:      signingCA.RootCert,
	}

	return nil
}
