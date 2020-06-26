package utils

import (
	"fmt"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/common/defaults"
	"github.com/solo-io/smh/pkg/mesh-discovery/utils/labelutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// construct an ObjectMeta for an Installed mesh from the control plane deployment
func DiscoveredObjectMeta(sourceResource ezkube.ResourceId) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: defaults.GetPodNamespace(),
		Name:      DiscoveredResourceName(sourceResource),
		Labels:    labelutils.ClusterLabels(sourceResource.GetClusterName()),
	}
}

// util for conventionally naming discovered resources
func DiscoveredResourceName(sourceResource ezkube.ResourceId) string {
	return kubeutils.SanitizeNameV2(fmt.Sprintf("%v-%v-%v", sourceResource.GetName(), sourceResource.GetNamespace(), sourceResource.GetClusterName()))
}
