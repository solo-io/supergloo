package translation

import (
	"context"

	"github.com/rotisserie/eris"
	corev1clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/utils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

type Output struct {
	SignedCertificate []byte
	SigningRootCa     []byte
}

//go:generate mockgen -source ./cert_issuer_translator.go -destination mocks/translator.go
// the networking translator translates an istio input networking snapshot to an istiooutput snapshot of mesh config resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
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

	// This translator only cares about CA with local secrets
	if issuedCertificate.Spec.GetSigningCertificateSecret() == nil {
		return nil, nil
	}

	signingCertificateSecret, err := s.mgmtClusterSecretClient.GetSecret(ctx, ezkube.MakeClientObjectKey(issuedCertificate.Spec.GetSigningCertificateSecret()))
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
		issuedCertificate.Spec.CommonCertOptions.GetTtlDays(),
	)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to generate signed cert for certificate request %v", sets.Key(certificateRequest))
	}

	return &Output{
		SignedCertificate: signedCert,
		SigningRootCa:     signingCA.RootCert,
	}, nil
}
