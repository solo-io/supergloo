package testutils

import (
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	networkingv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	multiclusterv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	multiclusterv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
)

// TODO(harveyxia) generate this in skv2
type InputSnapshotBuilder struct {
	name               string
	meshServices       discoveryv1alpha2sets.MeshServiceSet
	meshWorkloads      discoveryv1alpha2sets.MeshWorkloadSet
	meshes             discoveryv1alpha2sets.MeshSet
	trafficPolicies    networkingv1alpha2sets.TrafficPolicySet
	accessPolicies     networkingv1alpha2sets.AccessPolicySet
	virtualMeshes      networkingv1alpha2sets.VirtualMeshSet
	failoverServices   networkingv1alpha2sets.FailoverServiceSet
	kubernetesClusters multiclusterv1alpha1sets.KubernetesClusterSet
}

func NewInputSnapshotBuilder(name string) *InputSnapshotBuilder {
	return &InputSnapshotBuilder{
		name:               name,
		meshServices:       discoveryv1alpha2sets.NewMeshServiceSet(),
		meshWorkloads:      discoveryv1alpha2sets.NewMeshWorkloadSet(),
		meshes:             discoveryv1alpha2sets.NewMeshSet(),
		trafficPolicies:    networkingv1alpha2sets.NewTrafficPolicySet(),
		accessPolicies:     networkingv1alpha2sets.NewAccessPolicySet(),
		virtualMeshes:      networkingv1alpha2sets.NewVirtualMeshSet(),
		failoverServices:   networkingv1alpha2sets.NewFailoverServiceSet(),
		kubernetesClusters: multiclusterv1alpha1sets.NewKubernetesClusterSet(),
	}
}

func (i *InputSnapshotBuilder) Build() input.Snapshot {
	return input.NewSnapshot(
		i.name,
		i.meshServices,
		i.meshWorkloads,
		i.meshes,
		i.trafficPolicies,
		i.accessPolicies,
		i.virtualMeshes,
		i.failoverServices,
		i.kubernetesClusters,
	)
}

func (i *InputSnapshotBuilder) AddMeshServices(meshServices []*discoveryv1alpha2.MeshService) *InputSnapshotBuilder {
	i.meshServices.Insert(meshServices...)
	return i
}

func (i *InputSnapshotBuilder) AddMeshWorkloads(meshWorkloads []*discoveryv1alpha2.MeshWorkload) *InputSnapshotBuilder {
	i.meshWorkloads.Insert(meshWorkloads...)
	return i
}

func (i *InputSnapshotBuilder) AddMeshes(meshes []*discoveryv1alpha2.Mesh) *InputSnapshotBuilder {
	i.meshes.Insert(meshes...)
	return i
}

func (i *InputSnapshotBuilder) AddTrafficPolicies(trafficPolicies []*networkingv1alpha2.TrafficPolicy) *InputSnapshotBuilder {
	i.trafficPolicies.Insert(trafficPolicies...)
	return i
}

func (i *InputSnapshotBuilder) AddAccessPolicies(accessPolicies []*networkingv1alpha2.AccessPolicy) *InputSnapshotBuilder {
	i.accessPolicies.Insert(accessPolicies...)
	return i
}

func (i *InputSnapshotBuilder) AddVirtualMeshes(virtualMeshes []*networkingv1alpha2.VirtualMesh) *InputSnapshotBuilder {
	i.virtualMeshes.Insert(virtualMeshes...)
	return i
}

func (i *InputSnapshotBuilder) AddFailoverServices(failoverServices []*networkingv1alpha2.FailoverService) *InputSnapshotBuilder {
	i.failoverServices.Insert(failoverServices...)
	return i
}

func (i *InputSnapshotBuilder) AddKubernetesClusters(kubernetesClusters []*multiclusterv1alpha1.KubernetesCluster) *InputSnapshotBuilder {
	i.kubernetesClusters.Insert(kubernetesClusters...)
	return i
}
