package kubepod

import (
	"reflect"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Pod kubev1.Pod

func (p *Pod) Clone() *Pod {
	vp := kubev1.Pod(*p)
	copy := vp.DeepCopy()
	newP := Pod(*copy)
	return &newP
}

func (p *Pod) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *Pod) SetMetadata(meta core.Metadata) {
	MergeCoreMetaIntoObjectMeta(meta, &p.ObjectMeta)
}

func (p *Pod) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}

func MergeCoreMetaIntoObjectMeta(from core.Metadata, to *metav1.ObjectMeta) {
	to.Namespace = from.Namespace
	to.Name = from.Name
	to.Annotations = from.Annotations
	to.Labels = from.Labels
	to.ResourceVersion = from.ResourceVersion
}

func FromKube(pod *kubev1.Pod) (*v1.Pod, error) {

	podCopy := pod.DeepCopy()
	kubePod := Pod(*podCopy)
	resource := &v1.Pod{
		Pod: kubePod,
	}

	return resource, nil
}

func ToKube(resource resources.Resource) (*kubev1.Pod, error) {
	podResource, ok := resource.(*v1.Pod)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to pod-only client", resources.Kind(resource))
	}

	pod := kubev1.Pod(podResource.Pod)

	return &pod, nil
}
