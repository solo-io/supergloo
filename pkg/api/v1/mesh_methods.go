package v1

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

const (
	MeshIngressNamespace   = "gloo-system"
	MeshIngressServiceName = "gateway-proxy-v2"
	MeshIngressPort        = "mesh-bridge"
)

func BuildMeshIngress(meshRef *core.ResourceRef, meshDiscoveryLabels map[string]string) *MeshIngress {
	return &MeshIngress{
		Metadata: core.Metadata{
			Name:      meshRef.Name,
			Namespace: meshRef.Namespace,
			Labels:    meshDiscoveryLabels,
		},
		IngressType: &MeshIngress_Gloo{
			Gloo: &MeshIngress_GlooIngress{
				Namespace:   MeshIngressNamespace,
				ServiceName: MeshIngressServiceName,
				Port:        MeshIngressPort,
			},
		},
	}
}

func (m *Mesh) GetMeshIngress() *MeshIngress {
	ref := m.Metadata.Ref()
	return BuildMeshIngress(&ref, m.Metadata.Labels)
}
