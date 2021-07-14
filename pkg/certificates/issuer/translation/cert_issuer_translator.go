package translation

import (
	"context"

	"github.com/rotisserie/eris"
	corev1clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/utils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"go.uber.org/zap"
)

// Information required by cert issuer reconciler to fulfill CertificateRequest
type Output struct {
	// Newly signed certirficate to be used by the mesh
	SignedCertificate []byte
	// Root CA used to sign this certificate
	SigningRootCa []byte
}

//go:generate mockgen -source ./cert_issuer_translator.go -destination mocks/translator.go

// The cert issuer translator represents an entity which translates the input resources, into
// the output resources as defined by the `Output` resource below.
// See Output struct above
type Translator interface {
	// Translate the input resources into the SignedCert and SigningRootCa
	// If Output and Err are nil, this translator is not responsible for this resource
	Translate(
		ctx context.Context,
		certificateRequest *certificatesv1.CertificateRequest,
		issuedCertificate *certificatesv1.IssuedCertificate,
	) (*Output, error)
}

func NewTranslator(mgmtClusterSecretClient corev1clients.SecretClient) Translator {
	return &secretTranslator{
		mgmtClusterSecretClient: mgmtClusterSecretClient,
	}
}

type secretTranslator struct {
	mgmtClusterSecretClient corev1clients.SecretClient
}

func (s *secretTranslator) Translate(
	ctx context.Context,
	certificateRequest *certificatesv1.CertificateRequest,
	issuedCertificate *certificatesv1.IssuedCertificate,
) (*Output, error) {

	ctx = contextutils.WithLoggerValues(
		ctx,
		zap.String("CertificateRequest", sets.Key(certificateRequest)),
		zap.String("IssuedCertificate", sets.Key(issuedCertificate)),
	)

	signingCert := GetSigningSecret(issuedCertificate)
	// This translator only cares about CA with local secrets
	if signingCert == nil {
		contextutils.LoggerFrom(ctx).Debugf("No signing cert found, not this translator's responsiliblity to sign CSR")
		return nil, nil
	}

	signingCertificateSecret, err := s.mgmtClusterSecretClient.GetSecret(ctx, ezkube.MakeClientObjectKey(signingCert))
	if err != nil {
		return nil, eris.Wrapf(err, "failed to find issuer's signing certificate matching issued request %v", sets.Key(issuedCertificate))
	}

	signingCA := secrets.RootCADataFromSecretData(signingCertificateSecret.Data)

	// generate the issued cert PEM encoded bytes
	signedCert, err := utils.GenCertForCSR(
		issuedCertificate.Spec.Hosts,
		certificateRequest.Spec.GetCertificateSigningRequest(),
		signingCA.RootCert,
		signingCA.PrivateKey,
		issuedCertificate.Spec.GetCertOptions().GetTtlDays(),
	)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to generate signed cert for certificate request %v", sets.Key(certificateRequest))
	}

	return &Output{
		SignedCertificate: signedCert,
		SigningRootCa:     signingCA.RootCert,
	}, nil
}

// Public for use in enterprise
func GetSigningSecret(issuedCertificate *certificatesv1.IssuedCertificate) *skv2corev1.ObjectRef {
	//Handle deprecated field
	if issuedCertificate.Spec.GetGlooMeshCa().GetSigningCertificateSecret() != nil {
		return issuedCertificate.Spec.GetGlooMeshCa().GetSigningCertificateSecret()
	} else if issuedCertificate.Spec.SigningCertificateSecret != nil {
		return issuedCertificate.Spec.SigningCertificateSecret
	}
	return nil
}
