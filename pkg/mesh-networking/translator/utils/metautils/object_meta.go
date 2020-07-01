package metautils

import (
	"fmt"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/resource"
	"github.com/solo-io/smh/pkg/common/defaults"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// the key used to differentiate translated resources by
// the SMH instance which produced them
var OwnershipLabelKey = fmt.Sprintf("owner.%s", v1alpha1.SchemeGroupVersion.Group)

// construct an ObjectMeta for a discovered resource from a source object (the object from which the resource was discovered)
func TranslatedObjectMeta(sourceObj resource.Resource, labels map[string]string, annotations map[string]string) metav1.ObjectMeta {
	if labels == nil {
		labels = map[string]string{}
	}
	// ownership label defaults to current namespace to allow multiple SMH tenancy within a cluster.
	labels[OwnershipLabelKey] = defaults.GetPodNamespace()
	return metav1.ObjectMeta{
		Name:        sourceObj.GetName(),
		Namespace:   sourceObj.GetNamespace(),
		Labels:      labels,
		Annotations: annotations,
	}
}
