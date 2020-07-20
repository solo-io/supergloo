package utils

import (
	"fmt"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// construct an ObjectMeta for a discovered resource from a source object (the object from which the resource was discovered)
func DiscoveredObjectMeta(sourceResource metav1.Object) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace:   defaults.GetPodNamespace(),
		Name:        DiscoveredResourceName(sourceResource),
		Labels:      labelutils.ClusterLabels(sourceResource.GetClusterName()),
		Annotations: sourceResource.GetAnnotations(),
	}
}

// util for conventionally naming discovered resources
func DiscoveredResourceName(sourceResource ezkube.ClusterResourceId) string {
	return kubeutils.SanitizeNameV2(fmt.Sprintf("%v-%v-%v", sourceResource.GetName(), sourceResource.GetNamespace(), sourceResource.GetClusterName()))
}
