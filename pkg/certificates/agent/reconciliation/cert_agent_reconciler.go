package reconciliation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/agent/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/agent/utils"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/common/secrets"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// all output resources are labeled to prevent
// resource collisions & garbage collection of
// secrets the agent doesn't own
var agentLabels = map[string]string{
	fmt.Sprintf("agent.%v", v1alpha2.SchemeGroupVersion.Group): defaults.GetPodNamespace(),
}

type certAgentReconciler struct {
	ctx         context.Context
	builder     input.Builder
	localClient client.Client
}

func Start(
	ctx context.Context,
	builder input.Builder,
	mgr manager.Manager,
) error {
	d := &certAgentReconciler{
		ctx:         ctx,
		builder:     builder,
		localClient: mgr.GetClient(),
	}

	return input.RegisterSingleClusterReconciler(ctx, mgr, d.reconcile)
}

const (
	// the map key used to store the agent's private key in a kube Secret
	privateKeySecretKey = "private-key"
)

// reconcile global state
func (r *certAgentReconciler) reconcile(_ ezkube.ResourceId) (bool, error) {
	inputSnap, err := r.builder.BuildSnapshot(r.ctx, "cert-agent", input.BuildOptions{})
	if err != nil {
		// failed to read from cache; should never happen
		return false, err
	}

	certificateRequests := v1alpha2sets.NewCertificateRequestSet()
	secrets := v1sets.NewSecretSet()

	// process issued certificates
	for _, issuedCertificate := range inputSnap.IssuedCertificates().List() {
		if err := r.reconcileIssuedCertificate(
			issuedCertificate,
			inputSnap.Secrets(), secrets,
			inputSnap.CertificateRequests(), certificateRequests,
		); err != nil {
			issuedCertificate.Status.Error = err.Error()
		}
	}

	return false, inputSnap.SyncStatuses(r.ctx, r.localClient)
}

func (r *certAgentReconciler) reconcileIssuedCertificate(
	issuedCertificate *v1alpha2.IssuedCertificate,
	inputSecrets, outputSecrets v1sets.SecretSet,
	inputCertificateRequests, outputCertificateRequests v1alpha2sets.CertificateRequestSet,
) error {
	// if observed generation is out of sync, treat the issued certificate as Pending (spec has been modified)
	if issuedCertificate.Status.ObservedGeneration != issuedCertificate.Generation {
		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_PENDING
	}

	// reset & update status
	issuedCertificate.Status.ObservedGeneration = issuedCertificate.Generation
	issuedCertificate.Status.Error = ""

	// state-machine style processor
	switch issuedCertificate.Status.State {
	case v1alpha2.IssuedCertificateStatus_FINISHED:
		// ensure issued cert secret exists, nothing to do for this issued certificate
		if _, err := inputSecrets.Find(issuedCertificate.Spec.IssuedCertificateSecret); err == nil {
			return nil
		}
		// otherwise, restart the workflow from PENDING
		fallthrough
	case v1alpha2.IssuedCertificateStatus_FAILED:
		// restart the workflow from PENDING
		fallthrough
	case v1alpha2.IssuedCertificateStatus_PENDING:
		// create a new private key
		privateKey, err := utils.GeneratePrivateKey()
		if err != nil {
			return err
		}
		outputSecrets.Insert(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuedCertificate.Name,
				Namespace: issuedCertificate.Namespace,
				Labels:    agentLabels,
			},
			Data: map[string][]byte{privateKeySecretKey: privateKey},
		})

		// create certificate request for private key
		csrBytes, err := utils.GenerateCertificateSigningRequest(
			issuedCertificate.Spec.Hosts,
			issuedCertificate.Spec.Org,
			privateKey,
		)
		if err != nil {
			return err
		}
		certificateRequest := &v1alpha2.CertificateRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuedCertificate.Name,
				Namespace: issuedCertificate.Namespace,
				Labels:    agentLabels,
			},
			Spec: v1alpha2.CertificateRequestSpec{
				CertificateSigningRequest: csrBytes,
			},
		}
		outputCertificateRequests.Insert(certificateRequest)

		// set status to REQUESTED
		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_REQUESTED
	case v1alpha2.IssuedCertificateStatus_REQUESTED:

		// retrieve private key
		privateKeySecret, err := inputSecrets.Find(issuedCertificate)
		if err != nil {
			return err
		}
		privateKey := privateKeySecret.Data[privateKeySecretKey]

		if len(privateKey) == 0 {
			return eris.Errorf("invalid private key found, no data provided")
		}

		// retrieve signed certificate
		certificateRequest, err := inputCertificateRequests.Find(issuedCertificate)
		if err != nil {
			return err
		}

		switch certificateRequest.Status.State {
		case v1alpha2.CertificateRequestStatus_PENDING:
			contextutils.LoggerFrom(r.ctx).Infof("waiting for certificate request %v to be signed by Issuer", sets.Key(certificateRequest))

			// if the certificate signing request has not been
			// fulfilled, return and wait for the next reconcile
			return nil
		case v1alpha2.CertificateRequestStatus_FAILED:
			return eris.Errorf("certificate request %v failed to be signed by Issuer: %v", sets.Key(certificateRequest), certificateRequest.Status.Error)
		}

		signedCert := certificateRequest.Status.SignedCertificate
		signingRootCA := certificateRequest.Status.SigningRootCa

		issuedCertificateData := secrets.IntermediateCAData{
			RootCAData: secrets.RootCAData{
				RootCert: signingRootCA,
			},
			CertChain:    utils.AppendRootCerts(signedCert, signingRootCA),
			CaCert:       signedCert,
			CaPrivateKey: privateKey,
		}

		// finally, create the secret
		issuedCertificateSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      issuedCertificate.Spec.IssuedCertificateSecret.Name,
				Namespace: issuedCertificate.Spec.IssuedCertificateSecret.Namespace,
				Labels:    agentLabels,
			},
			Data: issuedCertificateData.ToSecretData(),
		}
		outputSecrets.Insert(issuedCertificateSecret)

		issuedCertificate.Status.State = v1alpha2.IssuedCertificateStatus_FINISHED
	default:
		return eris.Errorf("unknown issued certificate state: %v", issuedCertificate.Status.State)
	}

	return nil
}
