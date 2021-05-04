package secret

import (
	"context"

	"github.com/rotisserie/eris"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/translation"
	"github.com/solo-io/gloo-mesh/pkg/certificates/issuer/utils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func NewSecretTranslator(mgmtClusterSecrets corev1.SecretClient) translation.Translator {
	return &secretTranslator{
		mgmtClusterSecrets: mgmtClusterSecrets,
	}
}

type secretTranslator struct {
	mgmtClusterSecrets corev1.SecretClient
}

func (s *secretTranslator) Translate(
	ctx context.Context,
	certificateRequest *certificatesv1.CertificateRequest,
	issuedCertificate *certificatesv1.IssuedCertificate,
) (*translation.Output, error) {

	// This translator only cares about CA with local secrets
	if issuedCertificate.Spec.GetSigningCertificateSecret() == nil {
		return nil, nil
	}

	signingCertificateSecret, err := s.mgmtClusterSecrets.GetSecret(ctx, ezkube.MakeClientObjectKey(issuedCertificate.Spec.GetSigningCertificateSecret()))
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
		issuedCertificate.Spec.GetTtlDays(),
	)
	if err != nil {
		return nil, eris.Wrapf(err, "failed to generate signed cert for certificate request %v", sets.Key(certificateRequest))
	}

	return &translation.Output{
		SignedCertificate: signedCert,
		SigningRootCa:     signingCA.RootCert,
	}, nil
}
