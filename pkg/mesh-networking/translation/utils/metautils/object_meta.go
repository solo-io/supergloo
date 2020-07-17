package metautils

import (
	"fmt"

	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/common/defaults"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// the key used to differentiate translated resources by
// the SMH instance which produced them
var OwnershipLabelKey = fmt.Sprintf("owner.%s", v1alpha1.SchemeGroupVersion.Group)

// construct an ObjectMeta for a discovered resource from a source object (the object from which the resource was discovered)
func TranslatedObjectMeta(sourceObj ezkube.ClusterResourceId, annotations map[string]string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        sourceObj.GetName(),
		Namespace:   sourceObj.GetNamespace(),
		ClusterName: sourceObj.GetClusterName(),
		Labels:      TranslatedObjectLabels(),
		Annotations: annotations,
	}
}

// ownership label defaults to current namespace to allow multiple SMH tenancy within a cluster.
func TranslatedObjectLabels() map[string]string {
	return map[string]string{OwnershipLabelKey: defaults.GetPodNamespace()}
}
