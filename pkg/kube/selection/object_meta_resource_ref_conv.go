package selection

import (
	"fmt"

	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// returns true if the object metas seem to point to the same resource, in terms of name/namespace/cluster
// (cluster probably isn't set on these objects, but this is just in the interest of future-proofing)
func SameObject(this k8s_meta_types.ObjectMeta, that k8s_meta_types.ObjectMeta) bool {
	return this.GetName() == that.GetName() && this.GetNamespace() == that.GetNamespace() && this.GetClusterName() == that.GetClusterName()
}

// turn an ObjectMeta into a unique (single-cluster) string that can be used in sets, map keys, etc.
func ToUniqueSingleClusterString(obj k8s_meta_types.ObjectMeta) string {
	return fmt.Sprintf("%s+%s", obj.Name, obj.Namespace)
}

func ObjectMetaToResourceRef(objMeta k8s_meta_types.ObjectMeta) *zephyr_core_types.ResourceRef {
	return &zephyr_core_types.ResourceRef{
		Name:      objMeta.GetName(),
		Namespace: objMeta.GetNamespace(),
	}
}

func ObjectMetaToObjectKey(objMeta k8s_meta_types.ObjectMeta) client.ObjectKey {
	return client.ObjectKey{
		Name:      objMeta.GetName(),
		Namespace: objMeta.GetNamespace(),
	}
}

func ResourceRefToObjectMeta(ref *zephyr_core_types.ResourceRef) k8s_meta_types.ObjectMeta {
	return k8s_meta_types.ObjectMeta{
		Name:      ref.GetName(),
		Namespace: ref.GetNamespace(),
	}
}

func ResourceRefToObjectKey(ref *zephyr_core_types.ResourceRef) client.ObjectKey {
	return client.ObjectKey{
		Name:      ref.GetName(),
		Namespace: ref.GetNamespace(),
	}
}
