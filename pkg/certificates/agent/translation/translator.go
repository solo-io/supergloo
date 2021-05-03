package translation

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/output/certagent"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/utils"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// the map key used to store the agent's private key in a kube Secret
	privateKeySecretKey = "private-key"
)

func agentLabels() map[string]string {
	labels := map[string]string{
		fmt.Sprintf("agent.%v", v1.SchemeGroupVersion.Group): defaults.GetPodNamespace(),
	}
	if agentCluster := defaults.GetAgentCluster(); agentCluster != "" {
		labels[metautils.AgentLabelKey] = agentCluster
	}
	return labels
}

func PrivateKeySecretType() corev1.SecretType {
	return corev1.SecretType(fmt.Sprintf("%s/generated_private_key", v1.SchemeGroupVersion.Group))
}

func IssuedCertificateSecretType() corev1.SecretType {
	return corev1.SecretType(fmt.Sprintf("%s/issued_certificate", v1.SchemeGroupVersion.Group))
}

type Translator interface {
	IssuedCertiticatePending(
		ctx context.Context,
		issuedCertificate *v1.IssuedCertificate,
		outputs certagent.Builder,
	) ([]byte, error)
	IssuedCertificateRequested(
		ctx context.Context,
		issuedCertificate *v1.IssuedCertificate,
		certificateRequest *v1.CertificateRequest,
		inputs input.Snapshot,
		outputs certagent.Builder,
	) error
	IssuedCertificateIssued(
		ctx context.Context,
		issuedCertificate *v1.IssuedCertificate,
		inputs input.Snapshot,
		outputs certagent.Builder,
	) error
}

type certAgentTranslator struct {
	localClient client.Client
}

func (c *certAgentTranslator) IssuedCertificateIssued(
	ctx context.Context,
	issuedCertificate *v1.IssuedCertificate,
	inputs input.Snapshot,
	outputs certagent.Builder,
) error {

	// ensure issued cert secret exists, if not, return an error (restart the workflow)
	if issuedCertificateSecret, err := inputs.Secrets().Find(issuedCertificate.Spec.IssuedCertificateSecret); err != nil {
		return err
	} else {
		// add secret output to prevent it from being GC'ed
		outputs.AddSecrets(issuedCertificateSecret)
	}
	return nil
}

func (c *certAgentTranslator) IssuedCertiticateRequested(
	ctx context.Context,
	issuedCertificate *v1.IssuedCertificate,
	certificateRequest *v1.CertificateRequest,
	inputs input.Snapshot,
	outputs certagent.Builder,
) error {

	// if issuedCertificate.Spec.GetSe

	// retrieve private key
	privateKeySecret, err := inputs.Secrets().Find(issuedCertificate)
	if err != nil {
		return err
	}
	privateKey := privateKeySecret.Data[privateKeySecretKey]

	if len(privateKey) == 0 {
		return eris.Errorf("invalid private key found, no data provided")
	}

	switch certificateRequest.Status.State {
	case v1.CertificateRequestStatus_PENDING:
		contextutils.LoggerFrom(ctx).Infof("waiting for certificate request %v to be signed by Issuer", sets.Key(certificateRequest))

		// add secret and certrequest to output to prevent them from being GC'ed
		outputs.AddSecrets(privateKeySecret)
		outputs.AddCertificateRequests(certificateRequest)

		// if the certificate signing request has not been
		// fulfilled, return and wait for the next reconcile
		return nil
	case v1.CertificateRequestStatus_FAILED:
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
			Labels:    agentLabels(),
		},
		Data: issuedCertificateData.ToSecretData(),
		Type: IssuedCertificateSecretType(),
	}
	outputs.AddSecrets(issuedCertificateSecret)
	return nil
}

func (c *certAgentTranslator) IssuedCertiticatePending(
	ctx context.Context,
	issuedCertificate *v1.IssuedCertificate,
	outputs certagent.Builder,
) ([]byte, error) {
	// create a new private key
	privateKey, err := utils.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	outputs.AddSecrets(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      issuedCertificate.Name,
			Namespace: issuedCertificate.Namespace,
			Labels:    agentLabels(),
		},
		Data: map[string][]byte{privateKeySecretKey: privateKey},
		Type: PrivateKeySecretType(),
	})

	// create certificate request for private key
	csrBytes, err := utils.GenerateCertificateSigningRequest(
		issuedCertificate.Spec.Hosts,
		issuedCertificate.Spec.Org,
		privateKey,
	)
	if err != nil {
		return nil, err
	}
	return csrBytes, err
}
