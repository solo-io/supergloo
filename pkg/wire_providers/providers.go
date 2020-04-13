package wire_providers

import (
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"k8s.io/client-go/rest"
)

func NewDiscoveryClients(cfg *rest.Config) (zephyr_discovery.Clientset, error) {
	return zephyr_discovery.NewClientsetFromConfig(cfg)
}

func NewNetworkingClients(cfg *rest.Config) (zephyr_networking.Clientset, error) {
	return zephyr_networking.NewClientsetFromConfig(cfg)
}

func NewSecurityClients(cfg *rest.Config) (zephyr_security.Clientset, error) {
	return zephyr_security.NewClientsetFromConfig(cfg)
}

func NewKubernetesClusterClient(clients zephyr_discovery.Clientset) zephyr_discovery.KubernetesClusterClient {
	return clients.KubernetesClusters()
}

func NewMeshServiceClient(clients zephyr_discovery.Clientset) zephyr_discovery.MeshServiceClient {
	return clients.MeshServices()
}

func NewMeshWorkloadClient(clients zephyr_discovery.Clientset) zephyr_discovery.MeshWorkloadClient {
	return clients.MeshWorkloads()
}

func NewMeshClient(clients zephyr_discovery.Clientset) zephyr_discovery.MeshClient {
	return clients.Meshes()
}

func NewTrafficPolicyClient(clients zephyr_networking.Clientset) zephyr_networking.TrafficPolicyClient {
	return clients.TrafficPolicies()
}

func NewAccessControlPolicyClient(clients zephyr_networking.Clientset) zephyr_networking.AccessControlPolicyClient {
	return clients.AccessControlPolicies()
}

func NewVirtualMeshClient(clients zephyr_networking.Clientset) zephyr_networking.VirtualMeshClient {
	return clients.VirtualMeshes()
}

func NewVirtualMeshCertificateSigningRequestClient(clients zephyr_security.Clientset) zephyr_security.VirtualMeshCertificateSigningRequestClient {
	return clients.VirtualMeshCertificateSigningRequests()
}
