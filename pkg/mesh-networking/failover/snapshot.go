package failover

import (
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	v1alpha1sets2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/sets"
)

type InputSnapshot struct {
	FailoverServices v1alpha1sets.FailoverServiceSet
	MeshServices     v1alpha1sets2.MeshServiceSet
	// For validation
	KubeClusters  v1alpha1sets2.KubernetesClusterSet
	Meshes        v1alpha1sets2.MeshSet
	VirtualMeshes v1alpha1sets.VirtualMeshSet
}

type OutputSnapshot struct {
	FailoverServices v1alpha1sets.FailoverServiceSet
	MeshOutputs      MeshOutputs
}

type MeshOutputs struct {
	// Istio
	ServiceEntries v1alpha3sets.ServiceEntrySet
	EnvoyFilters   v1alpha3sets.EnvoyFilterSet
}

func NewOutputSnapshot() OutputSnapshot {
	return OutputSnapshot{
		FailoverServices: v1alpha1sets.NewFailoverServiceSet(),
		MeshOutputs:      NewMeshOutputs(),
	}
}

func NewMeshOutputs() MeshOutputs {
	return MeshOutputs{
		ServiceEntries: v1alpha3sets.NewServiceEntrySet(),
		EnvoyFilters:   v1alpha3sets.NewEnvoyFilterSet(),
	}
}

// Append entries from the given OutputSnapshot to this OutputSnapshot
func (this OutputSnapshot) Union(that OutputSnapshot) OutputSnapshot {
	if that.FailoverServices != nil {
		this.FailoverServices = this.FailoverServices.Union(that.FailoverServices)
	}
	this.MeshOutputs = this.MeshOutputs.Union(that.MeshOutputs)
	return this
}

func (this MeshOutputs) Union(that MeshOutputs) MeshOutputs {
	if that.ServiceEntries != nil {
		this.ServiceEntries = this.ServiceEntries.Union(that.ServiceEntries)
	}
	if that.EnvoyFilters != nil {
		this.EnvoyFilters = this.EnvoyFilters.Union(that.EnvoyFilters)
	}
	return this
}

// Make ResourceRef compatible with sets.Find()
type ResourceId struct {
	ResourceRef *smh_core_types.ResourceRef
}

func (r ResourceId) GetName() string {
	return r.ResourceRef.GetName()
}

func (r ResourceId) GetNamespace() string {
	return r.ResourceRef.GetNamespace()
}

// TODO rename ResourceRef.cluster to clusterName
func (r ResourceId) GetClusterName() string {
	return r.ResourceRef.GetCluster()
}
