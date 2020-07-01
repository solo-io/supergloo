package utils

import (
	"fmt"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/common/defaults"
	"github.com/solo-io/smh/pkg/mesh-discovery/utils/labelutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// construct an ObjectMeta for a discovered resource from a source object (the object from which the resource was discovered)
func DiscoveredObjectMeta(sourceResource metav1.Object) metav1.ObjectMeta {
	labels := sourceResource.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	for k, v := range labelutils.ClusterLabels(sourceResource.GetClusterName()) {
		labels[k] = v
	}
	return metav1.ObjectMeta{
		Namespace:   defaults.GetPodNamespace(),
		Name:        DiscoveredResourceName(sourceResource),
		Labels:      labels,
		Annotations: sourceResource.GetAnnotations(),
	}
}

// util for conventionally naming discovered resources
func DiscoveredResourceName(sourceResource ezkube.ResourceId) string {
	return kubeutils.SanitizeNameV2(fmt.Sprintf("%v-%v-%v", sourceResource.GetName(), sourceResource.GetNamespace(), sourceResource.GetClusterName()))
}
