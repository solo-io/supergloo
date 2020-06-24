package failover

import (
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	istio_client_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

type InputSnapshot struct {
	FailoverServices []*smh_networking.FailoverService
	MeshServices     []*smh_discovery.MeshService
	// For validation
	KubeClusters  []*smh_discovery.KubernetesCluster
	Meshes        []*smh_discovery.Mesh
	VirtualMeshes []*smh_networking.VirtualMesh
}

type OutputSnapshot struct {
	FailoverServices []*smh_networking.FailoverService
	MeshOutputs      MeshOutputs
}

type MeshOutputs struct {
	// Istio
	ServiceEntries []*istio_client_networking.ServiceEntry
	EnvoyFilters   []*istio_client_networking.EnvoyFilter
}

// Append entries from the given OutputSnapshot to this OutputSnapshot
func (this OutputSnapshot) Append(that OutputSnapshot) OutputSnapshot {
	this.FailoverServices = append(this.FailoverServices, that.FailoverServices...)
	this.MeshOutputs = this.MeshOutputs.Append(that.MeshOutputs)
	return this
}

func (this MeshOutputs) Append(that MeshOutputs) MeshOutputs {
	this.ServiceEntries = append(this.ServiceEntries, that.ServiceEntries...)
	this.EnvoyFilters = append(this.EnvoyFilters, that.EnvoyFilters...)
	return this
}
