package mtls

import (
	"context"
	"fmt"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/certificates/common/secrets"
	"istio.io/istio/pkg/spiffe"
	"istio.io/istio/security/pkg/pki/util"
	corev1 "k8s.io/api/core/v1"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	certificatesv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/certificates.smh.solo.io/v1alpha2"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultIstioOrg              = "Istio"
	defaultCitadelServiceAccount = "istio-citadel" // The default SPIFFE URL value for trust domain
	defaultTrustDomain           = "cluster.local"
	defaultIstioNamespace        = "istio-system"
	// name of the istio root CA secret
	// https://istio.io/latest/docs/tasks/security/cert-management/plugin-ca-cert/
	istioCaSecretName = "cacerts"
)

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given Mesh.
	// returns nil if no VirtualService or DestinationRule is required for the Mesh (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	//
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		mesh *discoveryv1alpha2.Mesh,
		virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
		outputs output.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx context.Context
}

func NewTranslator(ctx context.Context) Translator {
	return &translator{ctx: ctx}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	mesh *discoveryv1alpha2.Mesh,
	virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
	outputs output.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return
	}

	if virtualMesh == nil || virtualMesh.Spec.MtlsConfig == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("no translation for virtual mesh %v which has no mTLS configuration", sets.Key(mesh))
		return
	}
	mtlsConfig := virtualMesh.Spec.MtlsConfig

	// TODO(ilackarms): currently we assume a shared trust model
	// we'll want to expand this to support limited trust in the future
	sharedTrust := mtlsConfig.GetShared()
	rootCA := sharedTrust.GetRootCertificateAuthority()

	var rootCaSecret *v1.ObjectRef
	switch caType := rootCA.CaSource.(type) {
	case *v1alpha2.VirtualMeshSpec_RootCertificateAuthority_Generated:
		selfSignedCert, err := generateSelfSignedCert(caType.Generated)
		if err != nil {
			// should never happen
			reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, err)
			return
		}
		// the self signed cert goes to the master/local cluster
		selfSignedCertSecret := &corev1.Secret{
			ObjectMeta: metautils.TranslatedObjectMeta(mesh, mesh.Annotations),
			Data:       selfSignedCert.ToSecretData(),
		}
		outputs.AddSecrets(selfSignedCertSecret)
		rootCaSecret = ezkube.MakeObjectRef(selfSignedCertSecret)
	case *v1alpha2.VirtualMeshSpec_RootCertificateAuthority_Secret:
		rootCaSecret = caType.Secret
	}

	trustDomain := istioMesh.GetCitadelInfo().GetTrustDomain()
	if trustDomain == "" {
		trustDomain = defaultTrustDomain
	}
	citadelServiceAccount := istioMesh.GetCitadelInfo().GetCitadelServiceAccount()
	if citadelServiceAccount == "" {
		citadelServiceAccount = defaultCitadelServiceAccount
	}
	istioNamespace := istioMesh.GetInstallation().GetNamespace()
	if istioNamespace == "" {
		istioNamespace = defaultIstioNamespace
	}

	// the default location of the istio CA Certs secret
	// the certificate workflow will produce a cert with this ref
	istioCaCerts := &v1.ObjectRef{
		Name:      istioCaSecretName,
		Namespace: istioNamespace,
	}

	issuedCertificate := &certificatesv1alpha2.IssuedCertificate{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: certificatesv1alpha2.IssuedCertificateSpec{
			Hosts:                    []string{buildSpiffeURI(trustDomain, istioNamespace, citadelServiceAccount)},
			Org:                      defaultIstioOrg,
			SigningCertificateSecret: rootCaSecret,
			IssuedCertificateSecret:  istioCaCerts,
		},
	}
	outputs.AddIssuedCertificates(issuedCertificate)

}

const (
	defaultRootCertTTLDays     = 365
	defaultRootCertTTLDuration = defaultRootCertTTLDays * 24 * time.Hour
	defaultRootCertRsaKeySize  = 4096
	defaultOrgName             = "service-mesh-hub"
)

func generateSelfSignedCert(
	builtinCA *v1alpha2.VirtualMeshSpec_RootCertificateAuthority_SelfSignedCert,
) (*secrets.RootCAData, error) {
	org := defaultOrgName
	if builtinCA.GetOrgName() != "" {
		org = builtinCA.GetOrgName()
	}
	ttl := defaultRootCertTTLDuration
	if builtinCA.GetTtlDays() > 0 {
		ttl = time.Duration(builtinCA.GetTtlDays()) * 24 * time.Hour
	}
	rsaKeySize := defaultRootCertRsaKeySize
	if builtinCA.GetRsaKeySizeBytes() > 0 {
		rsaKeySize = int(builtinCA.GetRsaKeySizeBytes())
	}
	options := util.CertOptions{
		Org:          org,
		IsCA:         true,
		IsSelfSigned: true,
		TTL:          ttl,
		RSAKeySize:   rsaKeySize,
		PKCS8Key:     false, // currently only supporting PKCS1
	}
	cert, key, err := util.GenCertKeyFromOptions(options)
	if err != nil {
		return nil, err
	}
	rootCaData := &secrets.RootCAData{
		PrivateKey: key,
		RootCert:   cert,
	}
	return rootCaData, nil
}

func buildSpiffeURI(trustDomain, namespace, serviceAccount string) string {
	return fmt.Sprintf("%s%s/ns/%s/sa/%s", spiffe.URIPrefix, trustDomain, namespace, serviceAccount)
}
