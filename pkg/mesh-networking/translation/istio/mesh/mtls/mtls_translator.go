package mtls

import (
	"context"
	"fmt"
	"time"

	"github.com/rotisserie/eris"

	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	"github.com/solo-io/skv2/pkg/ezkube"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"istio.io/istio/pkg/spiffe"
	"istio.io/istio/security/pkg/pki/util"
	corev1 "k8s.io/api/core/v1"

	certificatesv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	istioUtils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/utils"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
)

//go:generate mockgen -source ./mtls_translator.go -destination mocks/mtls_translator.go

const (
	defaultIstioOrg              = "Istio"
	defaultCitadelServiceAccount = "istio-citadel"
	defaultTrustDomain           = "cluster.local" // The default SPIFFE URL value for trust domain
	defaultIstioNamespace        = "istio-system"
	// name of the istio root CA secret
	// https://istio.io/latest/docs/tasks/security/cert-management/plugin-ca-cert/
	istioCaSecretName   = "cacerts"
	defaultCaSecretName = "smh-cacerts"
)

var (
	signingCertSecretType = corev1.SecretType(fmt.Sprintf("%s/generated_signing_cert", certificatesv1alpha2.SchemeGroupVersion.Group))

	// used when the user provides a nil root cert
	defaultSelfSignedRootCa = &v1alpha2.VirtualMeshSpec_RootCertificateAuthority{
		CaSource: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority_Generated{
			Generated: &v1alpha2.VirtualMeshSpec_RootCertificateAuthority_SelfSignedCert{
				TtlDays:         defaultRootCertTTLDays,
				RsaKeySizeBytes: defaultRootCertRsaKeySize,
				OrgName:         defaultOrgName,
			},
		},
	}
)

// used by networking reconciler to filter ignored secrets
func IsSigningCert(secret *corev1.Secret) bool {
	return secret.Type == signingCertSecretType
}

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given Mesh.
	// returns nil if no VirtualService or DestinationRule is required for the Mesh (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the istio.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		mesh *discoveryv1alpha2.Mesh,
		virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
		istioOutputs istio.Builder,
		localOutputs local.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx       context.Context
	secrets   corev1sets.SecretSet
	workloads discoveryv1alpha2sets.WorkloadSet
}

func NewTranslator(ctx context.Context, secrets corev1sets.SecretSet, workloads discoveryv1alpha2sets.WorkloadSet) Translator {
	return &translator{
		ctx:       ctx,
		secrets:   secrets,
		workloads: workloads,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	mesh *discoveryv1alpha2.Mesh,
	virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.MeshType)
		return
	}

	if err := t.updateMtlsOutputs(mesh, virtualMesh, istioOutputs, localOutputs); err != nil {
		reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, err)
	}
}

func (t *translator) updateMtlsOutputs(
	mesh *discoveryv1alpha2.Mesh,
	virtualMesh *discoveryv1alpha2.MeshStatus_AppliedVirtualMesh,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
) error {
	mtlsConfig := virtualMesh.Spec.MtlsConfig
	if mtlsConfig == nil {
		// nothing to do
		contextutils.LoggerFrom(t.ctx).Debugf("no translation for virtual mesh %v which has no mTLS configuration", sets.Key(mesh))
		return nil
	}

	if mtlsConfig.TrustModel == nil {
		return eris.Errorf("must specify trust model to use for issuing certificates")
	}

	switch trustModel := mtlsConfig.TrustModel.(type) {
	case *v1alpha2.VirtualMeshSpec_MTLSConfig_Shared:
		return t.configureSharedTrust(
			mesh,
			trustModel.Shared,
			virtualMesh.Ref,
			istioOutputs,
			localOutputs,
			mtlsConfig.AutoRestartPods,
		)
	case *v1alpha2.VirtualMeshSpec_MTLSConfig_Limited:
		return t.configureLimitedTrust(
			mesh,
			trustModel.Limited,
			virtualMesh.Ref,
			istioOutputs,
			localOutputs,
			mtlsConfig.AutoRestartPods,
		)
	}

	return nil
}

// will create the secret if it is self-signed,
// otherwise will return the user-provided secret ref in the mtls config
func (t *translator) configureSharedTrust(
	mesh *discoveryv1alpha2.Mesh,
	sharedTrust *v1alpha2.VirtualMeshSpec_MTLSConfig_SharedTrust,
	virtualMeshRef *v1.ObjectRef,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	autoRestartPods bool,
) error {
	rootCA := sharedTrust.GetRootCertificateAuthority()

	rootCaSecret, err := t.getOrCreateRootCaSecret(
		rootCA,
		virtualMeshRef,
		localOutputs,
	)
	if err != nil {
		return err
	}

	agentInfo := mesh.Spec.AgentInfo
	if agentInfo == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("cannot configure root certificates for mesh %v which has no cert-agent", sets.Key(mesh))
		return nil
	}

	issuedCertificate, podBounceDirective := t.constructIssuedCertificate(
		mesh,
		rootCaSecret,
		agentInfo.AgentNamespace,
		autoRestartPods,
	)
	istioOutputs.AddIssuedCertificates(issuedCertificate)
	istioOutputs.AddPodBounceDirectives(podBounceDirective)
	return nil
}

// will create the secret if it is self-signed,
// otherwise will return the user-provided secret ref in the mtls config
func (t *translator) configureLimitedTrust(
	mesh *discoveryv1alpha2.Mesh,
	limitedTrust *v1alpha2.VirtualMeshSpec_MTLSConfig_LimitedTrust,
	virtualMeshRef *v1.ObjectRef,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	autoRestartPods bool,
) error {
	rootCA := limitedTrust.GetRootCertificateAuthority()

	rootCaSecret, err := t.getOrCreateRootCaSecret(
		rootCA,
		virtualMeshRef,
		localOutputs,
	)
	if err != nil {
		return err
	}

	agentInfo := mesh.Spec.AgentInfo
	if agentInfo == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("cannot configure root certificates for mesh %v which has no cert-agent", sets.Key(mesh))
		return nil
	}

	issuedCertificate, podBounceDirective := t.constructIssuedCertificateForLimitedTrust(
		mesh,
		rootCaSecret,
		virtualMeshRef,
		agentInfo.AgentNamespace,
	)
	istioOutputs.AddIssuedCertificates(issuedCertificate)
	istioOutputs.AddPodBounceDirectives(podBounceDirective)
	return nil
}

// will create the secret if it is self-signed,
// otherwise will return the user-provided secret ref in the mtls config
func (t *translator) getOrCreateRootCaSecret(
	rootCA *v1alpha2.VirtualMeshSpec_RootCertificateAuthority,
	virtualMeshRef *v1.ObjectRef,
	localOutputs local.Builder,
) (*v1.ObjectRef, error) {
	if rootCA == nil || rootCA.CaSource == nil {
		rootCA = defaultSelfSignedRootCa
	}

	var rootCaSecret *v1.ObjectRef
	switch caType := rootCA.CaSource.(type) {
	case *v1alpha2.VirtualMeshSpec_RootCertificateAuthority_Generated:
		generatedSecretName := virtualMeshRef.Name + "." + virtualMeshRef.Namespace
		// write the signing secret to the gloomesh namespace
		generatedSecretNamespace := defaults.GetPodNamespace()
		// use the existing secret if it exists
		rootCaSecret = &v1.ObjectRef{
			Name:      generatedSecretName,
			Namespace: generatedSecretNamespace,
		}
		selfSignedCertSecret, err := t.secrets.Find(rootCaSecret)
		if err != nil {
			selfSignedCert, err := generateSelfSignedCert(caType.Generated)
			if err != nil {
				// should never happen
				return nil, err
			}
			// the self signed cert goes to the master/local cluster
			selfSignedCertSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: generatedSecretName,
					// write to the agent namespace
					Namespace: generatedSecretNamespace,
					// ensure the secret is written to the maser/local cluster
					ClusterName: "",
					Labels:      metautils.TranslatedObjectLabels(),
				},
				Data: selfSignedCert.ToSecretData(),
				Type: signingCertSecretType,
			}
		}
		localOutputs.AddSecrets(selfSignedCertSecret)
	case *v1alpha2.VirtualMeshSpec_RootCertificateAuthority_Secret:
		rootCaSecret = caType.Secret
	}

	return rootCaSecret, nil
}

func (t *translator) constructIssuedCertificate(
	mesh *discoveryv1alpha2.Mesh,
	rootCaSecret *v1.ObjectRef,
	agentNamespace string,
	autoRestartPods bool,
) (*certificatesv1alpha2.IssuedCertificate, *certificatesv1alpha2.PodBounceDirective) {
	istioMesh := mesh.Spec.GetIstio()

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

	issuedCertificateMeta := metav1.ObjectMeta{
		Name: mesh.Name,
		// write to the agent namespace
		Namespace: agentNamespace,
		// write to the mesh cluster
		ClusterName: istioMesh.GetInstallation().GetCluster(),
		Labels:      metautils.TranslatedObjectLabels(),
	}

	// get the pods that need to be bounced for this mesh
	podsToBounce := getPodsToBounce(mesh, t.workloads, autoRestartPods)
	var (
		podBounceDirective *certificatesv1alpha2.PodBounceDirective
		podBounceRef       *v1.ObjectRef
	)
	if len(podsToBounce) > 0 {
		podBounceDirective = &certificatesv1alpha2.PodBounceDirective{
			ObjectMeta: issuedCertificateMeta,
			Spec: certificatesv1alpha2.PodBounceDirectiveSpec{
				PodsToBounce: podsToBounce,
			},
		}
		podBounceRef = ezkube.MakeObjectRef(podBounceDirective)
	}

	// issue a certificate to the mesh agent
	return &certificatesv1alpha2.IssuedCertificate{
		ObjectMeta: issuedCertificateMeta,
		Spec: certificatesv1alpha2.IssuedCertificateSpec{
			Hosts:                    []string{buildSpiffeURI(trustDomain, istioNamespace, citadelServiceAccount)},
			Org:                      defaultIstioOrg,
			SigningCertificateSecret: rootCaSecret,
			IssuedCertificateSecret:  istioCaCerts,
			PodBounceDirective:       podBounceRef,
			TlsType:                  certificatesv1alpha2.IssuedCertificateSpec_SHARED,
		},
	}, podBounceDirective
}

const (
	defaultRootCertTTLDays     = 365
	defaultRootCertTTLDuration = defaultRootCertTTLDays * 24 * time.Hour
	defaultRootCertRsaKeySize  = 4096
	defaultOrgName             = "gloo-mesh"
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

// get selectors for all the pods in a mesh; they need to be bounced (including the mesh control plane itself)
func getPodsToBounce(mesh *discoveryv1alpha2.Mesh, allWorkloads discoveryv1alpha2sets.WorkloadSet, autoRestartPods bool) []*certificatesv1alpha2.PodBounceDirectiveSpec_PodSelector {
	// if autoRestartPods is false, we rely on the user to manually restart their pods
	if !autoRestartPods {
		return nil
	}
	istioMesh := mesh.Spec.GetIstio()
	istioInstall := istioMesh.GetInstallation()

	// bounce the control plane pod
	podsToBounce := []*certificatesv1alpha2.PodBounceDirectiveSpec_PodSelector{
		{
			Namespace: istioInstall.Namespace,
			Labels:    istioInstall.PodLabels,
		},
	}

	// bounce the ingress gateway pods
	for _, gateway := range istioMesh.IngressGateways {
		podsToBounce = append(podsToBounce, &certificatesv1alpha2.PodBounceDirectiveSpec_PodSelector{
			Namespace: istioInstall.Namespace,
			Labels:    gateway.WorkloadLabels,
		})
	}

	// collect selectors from workloads matching this mesh
	allWorkloads.List(func(workload *discoveryv1alpha2.Workload) bool {
		kubeWorkload := workload.Spec.GetKubernetes()

		if kubeWorkload != nil && ezkube.RefsMatch(workload.Spec.Mesh, mesh) {
			podsToBounce = append(podsToBounce, &certificatesv1alpha2.PodBounceDirectiveSpec_PodSelector{
				Namespace: kubeWorkload.Controller.GetNamespace(),
				Labels:    kubeWorkload.PodLabels,
			})
		}

		return false
	})

	return podsToBounce
}

func (t *translator) constructIssuedCertificateForLimitedTrust(
	mesh *discoveryv1alpha2.Mesh,
	rootCaSecret *v1.ObjectRef,
	virtualMeshRef *v1.ObjectRef,
	agentNamespace string,
) (*certificatesv1alpha2.IssuedCertificate, *certificatesv1alpha2.PodBounceDirective) {
	istioMesh := mesh.Spec.GetIstio()

	istioNamespace := istioMesh.GetInstallation().GetNamespace()
	if istioNamespace == "" {
		istioNamespace = defaultIstioNamespace
	}

	issuedCertificate := &v1.ObjectRef{
		Name:      istioUtils.CreateCredentialsName(virtualMeshRef),
		Namespace: istioNamespace,
	}

	issuedCertificateMeta := metav1.ObjectMeta{
		Name: mesh.Name,
		// write to the agent namespace
		Namespace: agentNamespace,
		// write to the mesh cluster
		ClusterName: istioMesh.GetInstallation().GetCluster(),
		Labels:      metautils.TranslatedObjectLabels(),
	}

	// get the pods that need to be bounced for this mesh
	podsToBounce := getPodsToBounce(mesh, t.workloads, false)
	var (
		podBounceDirective *certificatesv1alpha2.PodBounceDirective
		podBounceRef       *v1.ObjectRef
	)
	if len(podsToBounce) > 0 {
		podBounceDirective = &certificatesv1alpha2.PodBounceDirective{
			ObjectMeta: issuedCertificateMeta,
			Spec: certificatesv1alpha2.PodBounceDirectiveSpec{
				PodsToBounce: podsToBounce,
			},
		}
		podBounceRef = ezkube.MakeObjectRef(podBounceDirective)
	}

	// issue a certificate to the mesh agent
	return &certificatesv1alpha2.IssuedCertificate{
		ObjectMeta: issuedCertificateMeta,
		Spec: certificatesv1alpha2.IssuedCertificateSpec{
			Hosts:                    []string{fmt.Sprintf("%s.global", istioMesh.GetInstallation().GetCluster())},
			Org:                      defaultIstioOrg,
			SigningCertificateSecret: rootCaSecret,
			IssuedCertificateSecret:  issuedCertificate,
			PodBounceDirective:       podBounceRef,
			TlsType:                  certificatesv1alpha2.IssuedCertificateSpec_LIMITED,
		},
	}, podBounceDirective
}
