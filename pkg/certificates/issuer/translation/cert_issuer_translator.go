package translation

import (
	"context"

	"github.com/rotisserie/eris"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

type TranslationExtensionFunc func() ([]Translator, error)

// the networking translator translates an istio input networking snapshot to an istiooutput snapshot of mesh config resources
type Translator interface {
	// errors reflect an internal translation error and should never happen
	Translate(
		ctx context.Context,
		certificateRequest *certificatesv1.CertificateRequest,
		issuedCertificate *certificatesv1.IssuedCertificate,
	) (*Output, error)
}

type Output struct {
	SignedCertificate []byte
	SigningRootCa     []byte
}

func NewChainTranslator(translators ...Translator) Translator {
	return &chainTranslator{
		translators: translators,
	}
}

type chainTranslator struct {
	translators []Translator
}

func (c *chainTranslator) Translate(
	ctx context.Context,
	certificateRequest *certificatesv1.CertificateRequest,
	issuedCertificate *certificatesv1.IssuedCertificate,
) (*Output, error) {
	for _, t := range c.translators {
		output, err := t.Translate(ctx, certificateRequest, issuedCertificate)
		if err != nil {
			return nil, err
		}
		if output != nil {
			return output, nil
		}
	}
	return nil, eris.Errorf("no cert issuer translator issuer impl worked for %s", sets.TypedKey(certificateRequest))
}
