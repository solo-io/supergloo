package utils

import (
	"fmt"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/smh/pkg/common/defaults"
	"github.com/solo-io/smh/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// construct an ObjectMeta for an Installed mesh from the control plane deployment
func MeshObjectMeta(controlPlaneDeployment *appsv1.Deployment) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: defaults.GetPodNamespace(),
		Name:      MeshName(controlPlaneDeployment),
		Labels:    labelutils.ClusterLabels(controlPlaneDeployment.ClusterName),
	}
}

// util for conventionally naming discovered Meshes
func MeshName(deployment *appsv1.Deployment) string {
	return kubeutils.SanitizeNameV2(fmt.Sprintf("%v-%v-%v", deployment.Name, deployment.Namespace, deployment.ClusterName))
}

