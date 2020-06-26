package types

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// a workload represents a controller for k8s pods
// supported types:
// - Deployment
// - DaemonSet
// - StatefulSet
type Workload interface {
	metav1.Object
	Kind() string
	GetPodTemplate() corev1.PodTemplateSpec
}

// cast to our type
func ToWorkload(workload metav1.Object) Workload {
	switch workloadType := workload.(type) {
	case *appsv1.Deployment:
		w := Deployment(*workloadType)
		return &w
	case *appsv1.DaemonSet:
		w := DaemonSet(*workloadType)
		return &w
	case *appsv1.StatefulSet:
		w := StatefulSet(*workloadType)
		return &w
	default:
		panic(fmt.Sprintf("invalid cast type: %T", workloadType))
	}
}

type Deployment appsv1.Deployment

func (w *Deployment) Kind() string {
	return "Deployment"
}

func (w *Deployment) GetPodTemplate() corev1.PodTemplateSpec {
	return w.Spec.Template
}

type DaemonSet appsv1.DaemonSet

func (w *DaemonSet) Kind() string {
	return "DaemonSet"
}

func (w *DaemonSet) GetPodTemplate() corev1.PodTemplateSpec {
	return w.Spec.Template
}

type StatefulSet appsv1.StatefulSet

func (w *StatefulSet) Kind() string {
	return "StatefulSet"
}

func (w *StatefulSet) GetPodTemplate() corev1.PodTemplateSpec {
	return w.Spec.Template
}
