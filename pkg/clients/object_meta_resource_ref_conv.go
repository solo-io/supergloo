package clients

import (
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
