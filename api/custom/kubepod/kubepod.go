package kubepod

import (
	"reflect"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Pod struct {
	v1.Pod
}

func (p *Pod) Clone() *Pod {
	copy := p.DeepCopy()
	return &Pod{
		Pod: *copy,
	}
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
