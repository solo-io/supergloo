package input

import (
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
)

// the snapshot of input resources used to compute
// the discovery output snapshot.
type Snapshot interface {
	Pods() corev1sets.PodSet
	Services() corev1sets.ServiceSet
	ConfigMaps() corev1sets.ConfigMapSet
	Deployments() appsv1sets.DeploymentSet
	ReplicaSets() appsv1sets.ReplicaSetSet
	DaemonSets() appsv1sets.DaemonSetSet
	StatefulSets() appsv1sets.StatefulSetSet
}

type snapshot struct {
	pods         corev1sets.PodSet
	services     corev1sets.ServiceSet
	configMaps   corev1sets.ConfigMapSet
	deployments  appsv1sets.DeploymentSet
	replicaSets  appsv1sets.ReplicaSetSet
	daemonSets   appsv1sets.DaemonSetSet
	statefulSets appsv1sets.StatefulSetSet
}

func NewSnapshot(
	pods corev1sets.PodSet,
	services corev1sets.ServiceSet,
	configMaps corev1sets.ConfigMapSet,
	deployments appsv1sets.DeploymentSet,
	replicaSets appsv1sets.ReplicaSetSet,
	daemonSets appsv1sets.DaemonSetSet,
	statefulSets appsv1sets.StatefulSetSet,
) Snapshot {
	return &snapshot{
		pods:         pods,
		services:     services,
		configMaps:   configMaps,
		deployments:  deployments,
		replicaSets:  replicaSets,
		daemonSets:   daemonSets,
		statefulSets: statefulSets,
	}
}

func (s snapshot) Pods() corev1sets.PodSet {
	return s.pods
}

func (s snapshot) Services() corev1sets.ServiceSet {
	return s.services
}

func (s snapshot) ConfigMaps() corev1sets.ConfigMapSet {
	return s.configMaps
}

func (s snapshot) Deployments() appsv1sets.DeploymentSet {
	return s.deployments
}

func (s snapshot) ReplicaSets() appsv1sets.ReplicaSetSet {
	return s.replicaSets
}

func (s snapshot) DaemonSets() appsv1sets.DaemonSetSet {
	return s.daemonSets
}

func (s snapshot) StatefulSets() appsv1sets.StatefulSetSet {
	return s.statefulSets
}
