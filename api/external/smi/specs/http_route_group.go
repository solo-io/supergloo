package specs

import (
	"reflect"

	"github.com/deislabs/smi-sdk-go/pkg/apis/specs/v1alpha1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
)

type HTTPRouteGroup v1alpha1.HTTPRouteGroup

func (p *HTTPRouteGroup) GetMetadata() core.Metadata {
	return kubeutils.FromKubeMeta(p.ObjectMeta)
}

func (p *HTTPRouteGroup) SetMetadata(meta core.Metadata) {
	p.ObjectMeta = kubeutils.ToKubeMeta(meta)
}

func (p *HTTPRouteGroup) Equal(that interface{}) bool {
	return reflect.DeepEqual(p, that)
}

func (p *HTTPRouteGroup) Clone() *HTTPRouteGroup {
	vp := v1alpha1.HTTPRouteGroup(*p)
	copy := vp.DeepCopy()
	newP := HTTPRouteGroup(*copy)
	return &newP
}
