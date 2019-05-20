package split

import (
	"reflect"

	"github.com/solo-io/supergloo/imported/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
)

type TrafficSplit v1alpha1.TrafficSplit

func (p *TrafficSplit) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *TrafficSplit) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *TrafficSplit) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}

func (p *TrafficSplit) Clone() *TrafficSplit {
	vp := v1alpha1.TrafficSplit(*p)
	copy := vp.DeepCopy()
	newP := TrafficSplit(*copy)
	return &newP
}
