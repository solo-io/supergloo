package mock_discovery

import (
	mp_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildDeployment(objMeta k8s_meta_v1.ObjectMeta) *k8s_apps_v1.Deployment {
	return &k8s_apps_v1.Deployment{
		ObjectMeta: objMeta,
	}
}

func BuildMesh(objMeta k8s_meta_v1.ObjectMeta) *mp_v1alpha1.Mesh {
	return &mp_v1alpha1.Mesh{
		ObjectMeta: objMeta,
	}
}
