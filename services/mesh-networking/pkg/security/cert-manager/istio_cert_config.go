package cert_manager

import (
	"fmt"

	"github.com/rotisserie/eris"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	"istio.io/istio/pkg/spiffe"
)

const (
	DefaultIstioOrg              = "Istio"
	DefaultCitadelServiceAccount = "istio-citadel" // The default SPIFFE URL value for trust domain
	DefaultTrustDomain           = "cluster.local"
)

var (
	IncorrectMeshTypeError = func(mesh *smh_discovery.Mesh) error {
		return eris.Errorf("invalid mesh type (%T) passed into istio certificate config producer",
			mesh.Spec.GetMeshType())
	}
)

func NewIstioCertConfigProducer() IstioCertConfigProducer {
	return &istioCertConfigProducer{}
}

type istioCertConfigProducer struct{}

type IstioCertConfigProducer CertConfigProducer

func BuildSpiffeURI(trustDomain, namespace, sa string) string {
	return fmt.Sprintf("%s%s/ns/%s/sa/%s", spiffe.URIPrefix, trustDomain, namespace, sa)
}

func (i *istioCertConfigProducer) ConfigureCertificateInfo(
	vm *smh_networking.VirtualMesh,
	mesh *smh_discovery.Mesh,
) (*smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig, error) {
	var istioMesh *smh_discovery_types.MeshSpec_IstioMesh
	if mesh.Spec.GetIstio1_6() != nil {
		istioMesh = mesh.Spec.GetIstio1_6().GetMetadata()
	} else if mesh.Spec.GetIstio1_5() != nil {
		istioMesh = mesh.Spec.GetIstio1_5().GetMetadata()
	}

	if istioMesh == nil {
		return nil, IncorrectMeshTypeError(mesh)
	}

	trustDomain := DefaultTrustDomain
	citadelServiceAccount := DefaultCitadelServiceAccount
	citadelNamespace := istioMesh.GetInstallation().GetInstallationNamespace()

	if istioMesh.GetCitadelInfo().GetTrustDomain() != "" {
		trustDomain = istioMesh.GetCitadelInfo().GetTrustDomain()
	}
	if istioMesh.GetCitadelInfo().GetCitadelNamespace() != "" {
		citadelNamespace = istioMesh.GetCitadelInfo().GetCitadelNamespace()
	}
	if istioMesh.GetCitadelInfo().GetCitadelServiceAccount() != "" {
		citadelServiceAccount = istioMesh.GetCitadelInfo().GetCitadelServiceAccount()
	}

	meshType, err := metadata.MeshToMeshType(mesh)
	if err != nil {
		return nil, err
	}

	return &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
		// TODO: Make citadel namespace discoverable
		Hosts:    []string{BuildSpiffeURI(trustDomain, citadelNamespace, citadelServiceAccount)},
		Org:      DefaultIstioOrg,
		MeshType: meshType,
	}, nil
}
