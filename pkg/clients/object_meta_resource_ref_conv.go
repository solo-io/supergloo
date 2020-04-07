package clients

import (
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// returns true if the object metas seem to point to the same resource, in terms of name/namespace/cluster
// (cluster probably isn't set on these objects, but this is just in the interest of future-proofing)
func SameObject(this v1.ObjectMeta, that v1.ObjectMeta) bool {
	return this.GetName() == that.GetName() && this.GetNamespace() == that.GetNamespace() && this.GetClusterName() == that.GetClusterName()
}

func ObjectMetaToResourceRef(objMeta v1.ObjectMeta) *core_types.ResourceRef {
	return &core_types.ResourceRef{
		Name:      objMeta.GetName(),
		Namespace: objMeta.GetNamespace(),
	}
}

func ObjectMetaToObjectKey(objMeta v1.ObjectMeta) client.ObjectKey {
	return client.ObjectKey{
		Name:      objMeta.GetName(),
		Namespace: objMeta.GetNamespace(),
	}
}

func ResourceRefToObjectMeta(ref *core_types.ResourceRef) v1.ObjectMeta {
	return v1.ObjectMeta{
		Name:      ref.GetName(),
		Namespace: ref.GetNamespace(),
	}
}

func ResourceRefToObjectKey(ref *core_types.ResourceRef) client.ObjectKey {
	return client.ObjectKey{
		Name:      ref.GetName(),
		Namespace: ref.GetNamespace(),
	}
}
