package mtls

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/rotisserie/eris"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"istio.io/istio/pkg/spiffe"
	"istio.io/istio/security/pkg/pki/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./mtls_translator.go -destination mocks/mtls_translator.go

const (
	defaultIstioOrg              = "Istio"
	defaultCitadelServiceAccount = "istio-citadel"
	defaultTrustDomain           = "cluster.local" // The default SPIFFE URL value for trust domain
	defaultIstioNamespace        = "istio-system"
	// name of the istio root CA secret
	// https://istio.io/latest/docs/tasks/security/cert-management/plugin-ca-cert/
	istioCaSecretName = "cacerts"
	// name of the istio root CA configmap distributed to all namespaces
	// copied from https://github.com/istio/istio/blob/88a2bfb/pilot/pkg/serviceregistry/kube/controller/namespacecontroller.go#L39
	// not imported due to issues with dependeny imports
	istioCaConfigMapName = "istio-ca-root-cert"

	defaultRootCertTTLDays                = 365
	defaultRootCertRsaKeySize             = 4096
	defaultOrgName                        = "gloo-mesh"
	defaultSecretRotationGracePeriodRatio = 0.10
)

var (
	signingCertSecretType = corev1.SecretType(
		fmt.Sprintf("%s/generated_signing_cert", certificatesv1.SchemeGroupVersion.Group),
	)

	// used when the user provides a nil root cert
	defaultSelfSignedRootCa = &networkingv1.RootCertificateAuthority{
		CaSource: &networkingv1.RootCertificateAuthority_Generated{
			Generated: &certificatesv1.CommonCertOptions{
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
		mesh *discoveryv1.Mesh,
		virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
		istioOutputs istio.Builder,
		localOutputs local.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx       context.Context
	secrets   corev1sets.SecretSet
	workloads discoveryv1sets.WorkloadSet
}

func NewTranslator(
	ctx context.Context,
	secrets corev1sets.SecretSet,
	workloads discoveryv1sets.WorkloadSet,
) Translator {
	return &translator{
		ctx:       ctx,
		secrets:   secrets,
		workloads: workloads,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	mesh *discoveryv1.Mesh,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.Type)
		return
	}

	if err := t.updateMtlsOutputs(mesh, virtualMesh, istioOutputs, localOutputs); err != nil {
		reporter.ReportVirtualMeshToMesh(mesh, virtualMesh.Ref, err)
	}
}

func (t *translator) updateMtlsOutputs(
	mesh *discoveryv1.Mesh,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
) error {
	mtlsConfig := virtualMesh.Spec.MtlsConfig
	if mtlsConfig == nil {
		// nothing to do
		contextutils.LoggerFrom(t.ctx).Debugf("no translation for VirtualMesh %v which has no mTLS configuration", sets.Key(mesh))
		return nil
	}

	if mtlsConfig.TrustModel == nil {
		return eris.Errorf("must specify trust model to use for issuing certificates")
	}

	switch trustModel := mtlsConfig.TrustModel.(type) {
	case *networkingv1.VirtualMeshSpec_MTLSConfig_Shared:
		return t.configureSharedTrust(
			mesh,
			trustModel.Shared,
			virtualMesh.Ref,
			istioOutputs,
			localOutputs,
			mtlsConfig.AutoRestartPods,
		)
	case *networkingv1.VirtualMeshSpec_MTLSConfig_Limited:
		return eris.Errorf("limited trust not supported in version %v of Gloo Mesh", version.Version)
	}

	return nil
}

// will create the secret if it is self-signed,
// otherwise will return the user-provided secret ref in the mtls config
func (t *translator) configureSharedTrust(
	mesh *discoveryv1.Mesh,
	sharedTrust *networkingv1.SharedTrust,
	virtualMeshRef *skv2corev1.ObjectRef,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	autoRestartPods bool,
) error {

	agentInfo := mesh.Spec.AgentInfo
	if agentInfo == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("cannot configure root certificates for mesh %v which has no cert-agent", sets.Key(mesh))
		return nil
	}

	// Construct the skeleton of the issuedCertificate
	issuedCertificate, podBounceDirective := t.constructIssuedCertificate(
		mesh,
		sharedTrust,
		agentInfo.AgentNamespace,
		autoRestartPods,
	)

	switch typedCa := sharedTrust.GetCertificateAuthority().(type) {
	case *networkingv1.SharedTrust_IntermediateCertificateAuthority:
		// Copy intermediate CA data to IssuedCertificate
		issuedCertificate.Spec.CertificateAuthority = &certificatesv1.IssuedCertificateSpec_AgentCa{
			AgentCa: typedCa.IntermediateCertificateAuthority,
		}
	case *networkingv1.SharedTrust_RootCertificateAuthority:
		switch typedCaSource := typedCa.RootCertificateAuthority.GetCaSource().(type) {
		case *networkingv1.RootCertificateAuthority_Generated:
			// Generated CA cert secret.
			// Check if it exists
			rootCaSecret, err := t.getOrCreateGeneratedCaSecret(
				typedCaSource.Generated,
				virtualMeshRef,
				localOutputs,
			)
			if err != nil {
				return err
			}
			issuedCertificate.Spec.CertificateAuthority = &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
				GlooMeshCa: &certificatesv1.RootCertificateAuthority{
					CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
						SigningCertificateSecret: rootCaSecret,
					},
				},
			}
			// Set deprecated field for backwards compatibility
			issuedCertificate.Spec.SigningCertificateSecret = rootCaSecret
		case *networkingv1.RootCertificateAuthority_Secret:
			issuedCertificate.Spec.CertificateAuthority = &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
				GlooMeshCa: &certificatesv1.RootCertificateAuthority{
					CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
						SigningCertificateSecret: typedCaSource.Secret,
					},
				},
			}
			issuedCertificate.Spec.SigningCertificateSecret = typedCaSource.Secret
			// Set deprecated field for backwards compatibility
		default:
			return eris.Errorf("No root ca source specified for Virtual Mesh (%s)", sets.Key(virtualMeshRef))
		}
	default:
		return eris.Errorf("No ca source specified for Virtual Mesh (%s)", sets.Key(virtualMeshRef))
	}

	// Append the VirtualMesh as a parent to each output resource
	metautils.AppendParent(t.ctx, issuedCertificate, virtualMeshRef, networkingv1.VirtualMesh{}.GVK())
	metautils.AppendParent(t.ctx, podBounceDirective, virtualMeshRef, networkingv1.VirtualMesh{}.GVK())

	istioOutputs.AddIssuedCertificates(issuedCertificate)
	istioOutputs.AddPodBounceDirectives(podBounceDirective)
	return nil
}

// will create the secret if it is self-signed,
// otherwise will return the user-provided secret ref in the mtls config
func (t *translator) getOrCreateGeneratedCaSecret(
	generatedRootCa *certificatesv1.CommonCertOptions,
	virtualMeshRef *skv2corev1.ObjectRef,
	localOutputs local.Builder,
) (*skv2corev1.ObjectRef, error) {

	if generatedRootCa == nil {
		generatedRootCa = defaultSelfSignedRootCa.GetGenerated()
	}

	generatedSecretName := virtualMeshRef.Name + "." + virtualMeshRef.Namespace
	// write the signing secret to the gloomesh namespace
	generatedSecretNamespace := defaults.GetPodNamespace()
	// use the existing secret if it exists
	rootCaSecret := &skv2corev1.ObjectRef{
		Name:      generatedSecretName,
		Namespace: generatedSecretNamespace,
	}
	selfSignedCertSecret, err := t.secrets.Find(rootCaSecret)
	if err != nil {
		selfSignedCert, err := generateSelfSignedCert(generatedRootCa)
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

	// Append the VirtualMesh as a parent to the output secret
	metautils.AppendParent(t.ctx, selfSignedCertSecret, virtualMeshRef, networkingv1.VirtualMesh{}.GVK())

	localOutputs.AddSecrets(selfSignedCertSecret)

	return rootCaSecret, nil
}

func (t *translator) constructIssuedCertificate(
	mesh *discoveryv1.Mesh,
	sharedTrust *networkingv1.SharedTrust,
	agentNamespace string,
	autoRestartPods bool,
) (*certificatesv1.IssuedCertificate, *certificatesv1.PodBounceDirective) {
	istioMesh := mesh.Spec.GetIstio()

	trustDomain := istioMesh.GetTrustDomain()
	if trustDomain == "" {
		trustDomain = defaultTrustDomain
	}
	istiodServiceAccount := istioMesh.GetIstiodServiceAccount()
	if istiodServiceAccount == "" {
		istiodServiceAccount = defaultCitadelServiceAccount
	}
	istioNamespace := istioMesh.GetInstallation().GetNamespace()
	if istioNamespace == "" {
		istioNamespace = defaultIstioNamespace
	}

	issuedCertificateMeta := BuildMeshResourceObjectMeta(mesh)

	// get the pods that need to be bounced for this mesh
	podsToBounce := getPodsToBounce(mesh, sharedTrust, t.workloads, autoRestartPods)
	var (
		podBounceDirective *certificatesv1.PodBounceDirective
		podBounceRef       *skv2corev1.ObjectRef
	)
	if len(podsToBounce) > 0 {
		podBounceDirective = &certificatesv1.PodBounceDirective{
			ObjectMeta: issuedCertificateMeta,
			Spec: certificatesv1.PodBounceDirectiveSpec{
				PodsToBounce: podsToBounce,
			},
		}
		podBounceRef = ezkube.MakeObjectRef(podBounceDirective)
	}

	issuedCert := &certificatesv1.IssuedCertificate{
		ObjectMeta: issuedCertificateMeta,
		Spec: certificatesv1.IssuedCertificateSpec{
			Hosts: []string{buildSpiffeURI(trustDomain, istioNamespace, istiodServiceAccount)},
			CertOptions: buildDefaultCertOptions(
				sharedTrust.GetIntermediateCertOptions(),
				defaultIstioOrg,
			),
			// Set deprecated field for backwards compatibility
			Org:                defaultIstioOrg,
			PodBounceDirective: podBounceRef,
		},
	}

	// Only set issuedCert when not using vault CA
	if sharedTrust.GetIntermediateCertificateAuthority().GetVault() == nil {
		// the default location of the istio CA Certs secret
		// the certificate workflow will produce a cert with this ref
		issuedCert.Spec.IssuedCertificateSecret = &skv2corev1.ObjectRef{
			Name:      istioCaSecretName,
			Namespace: istioNamespace,
		}
	}

	// issue a certificate to the mesh agent
	return issuedCert, podBounceDirective
}

func buildDefaultCertOptions(
	options *certificatesv1.CommonCertOptions,
	orgName string,
) *certificatesv1.CommonCertOptions {
	result := proto.Clone(options).(*certificatesv1.CommonCertOptions)
	if result == nil {
		result = &certificatesv1.CommonCertOptions{}
	}
	if result.GetOrgName() == "" {
		result.OrgName = orgName
	}
	if result.GetTtlDays() == 0 {
		result.TtlDays = defaultRootCertTTLDays
	}
	if result.GetRsaKeySizeBytes() == 0 {
		result.RsaKeySizeBytes = defaultRootCertRsaKeySize
	}
	if result.GetSecretRotationGracePeriodRatio() == 0 {
		result.SecretRotationGracePeriodRatio = defaultSecretRotationGracePeriodRatio
	}
	return result
}

func generateSelfSignedCert(
	builtinCA *certificatesv1.CommonCertOptions,
) (*secrets.RootCAData, error) {
	certOptions := buildDefaultCertOptions(builtinCA, defaultOrgName)
	options := util.CertOptions{
		Org:          certOptions.GetOrgName(),
		IsCA:         true,
		IsSelfSigned: true,
		TTL:          time.Duration(certOptions.GetTtlDays()) * 24 * time.Hour,
		RSAKeySize:   int(certOptions.GetRsaKeySizeBytes()),
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
func getPodsToBounce(
	mesh *discoveryv1.Mesh,
	sharedTrust *networkingv1.SharedTrust,
	allWorkloads discoveryv1sets.WorkloadSet,
	autoRestartPods bool,
) []*certificatesv1.PodBounceDirectiveSpec_PodSelector {
	// if autoRestartPods is false, we rely on the user to manually restart their pods
	if !autoRestartPods {
		return nil
	}
	istioMesh := mesh.Spec.GetIstio()
	istioInstall := istioMesh.GetInstallation()

	// bounce the control plane pod first
	// order matters
	var podsToBounce []*certificatesv1.PodBounceDirectiveSpec_PodSelector
	// If the pki-sidecar is fulfilling the issued certificate request,
	// then the control-plane should not be bounced.
	if sharedTrust.GetIntermediateCertificateAuthority().GetVault() == nil {
		podsToBounce = append(podsToBounce, &certificatesv1.PodBounceDirectiveSpec_PodSelector{
			Namespace: istioInstall.Namespace,
			Labels:    istioInstall.PodLabels,
			// ensure at least one replica of istiod is ready before restarting the other pods
			WaitForReplicas: 1,
		})
	}

	// bounce the ingress gateway pods
	for _, gateway := range istioMesh.IngressGateways {
		podsToBounce = append(podsToBounce, &certificatesv1.PodBounceDirectiveSpec_PodSelector{
			Namespace: istioInstall.Namespace,
			Labels:    gateway.WorkloadLabels,
			RootCertSync: &certificatesv1.PodBounceDirectiveSpec_PodSelector_RootCertSync{
				SecretRef: &skv2corev1.ObjectRef{
					Name:      istioCaSecretName,
					Namespace: istioInstall.Namespace,
				},
				SecretKey: secrets.RootCertID,
				ConfigMapRef: &skv2corev1.ObjectRef{
					Name:      istioCaConfigMapName,
					Namespace: istioInstall.Namespace,
				},
				ConfigMapKey: secrets.RootCertID,
			},
		})
	}

	// collect selectors from workloads matching this mesh
	allWorkloads.List(func(workload *discoveryv1.Workload) bool {
		kubeWorkload := workload.Spec.GetKubernetes()

		if kubeWorkload != nil && ezkube.RefsMatch(workload.Spec.Mesh, mesh) {
			podsToBounce = append(podsToBounce, &certificatesv1.PodBounceDirectiveSpec_PodSelector{
				Namespace: kubeWorkload.Controller.GetNamespace(),
				Labels:    kubeWorkload.PodLabels,
				RootCertSync: &certificatesv1.PodBounceDirectiveSpec_PodSelector_RootCertSync{
					SecretRef: &skv2corev1.ObjectRef{
						Name:      istioCaSecretName,
						Namespace: istioInstall.Namespace,
					},
					SecretKey: secrets.RootCertID,
					ConfigMapRef: &skv2corev1.ObjectRef{
						Name:      istioCaConfigMapName,
						Namespace: kubeWorkload.Controller.GetNamespace(),
					},
					ConfigMapKey: secrets.RootCertID,
				},
			})
		}

		return false
	})

	return podsToBounce
}

// Build the common ObjectMeta used for child certificate resources of this mesh
// Exposed for use in enterprise
func BuildMeshResourceObjectMeta(
	mesh *discoveryv1.Mesh,
) metav1.ObjectMeta {
	istioMesh := mesh.Spec.GetIstio()
	clusterName := istioMesh.GetInstallation().GetCluster()
	return metav1.ObjectMeta{
		Name: mesh.Name,
		// write to the agent namespace
		Namespace: mesh.Spec.GetAgentInfo().GetAgentNamespace(),
		// write to the mesh cluster
		ClusterName: clusterName,
		Labels:      metautils.TranslatedObjectLabels(),
	}

}
