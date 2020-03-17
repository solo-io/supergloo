package cert_manager

import (
	"fmt"

	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	"istio.io/istio/pkg/spiffe"
)

const (
	DefaultIstioOrg              = "Istio"
	DefaultCitadelServiceAccount = "istio-citadel" // The default SPIFFE URL value for trust domain
	DefaultTrustDomain           = "cluster.local"
)

var (
	IncorrectMeshTypeError = func(mesh *discovery_v1alpha1.Mesh) error {
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
	vm *networking_v1alpha1.VirtualMesh,
	mesh *discovery_v1alpha1.Mesh,
) (*security_types.CertConfig, error) {
	istioMesh := mesh.Spec.GetIstio()
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
	return &security_types.CertConfig{
		// TODO: Make citadel namespace discoverable
		Hosts:    []string{BuildSpiffeURI(trustDomain, citadelNamespace, citadelServiceAccount)},
		Org:      DefaultIstioOrg,
		MeshType: core_types.MeshType_ISTIO,
	}, nil
}
