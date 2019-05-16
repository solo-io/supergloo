package linkerd

import (
	"reflect"

	"github.com/deislabs/smi-sdk-go/pkg/apis/access/v1alpha1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
)

type TrafficTarget v1alpha1.TrafficTarget

func (p *TrafficTarget) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *TrafficTarget) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *TrafficTarget) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}

func (p *TrafficTarget) Clone() *TrafficTarget {
	vp := v1alpha1.TrafficTarget(*p)
	copy := vp.DeepCopy()
	newP := TrafficTarget(*copy)
	return &newP
}
