package metautils

import (
	"fmt"

	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// the key used to differentiate translated resources by
// the SMH instance which produced them
var OwnershipLabelKey = fmt.Sprintf("owner.%s", v1alpha2.SchemeGroupVersion.Group)

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

// construct an ObjectMeta for a resource for a federated source object
// meshInstallation represents the mesh instance to which the object will be output
func FederatedObjectMeta(
	sourceObj ezkube.ClusterResourceId,
	meshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation,
	annotations map[string]string,
) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        federatedObjectName(sourceObj),
		Namespace:   meshInstallation.Namespace,
		ClusterName: meshInstallation.Cluster,
		Labels:      TranslatedObjectLabels(),
		Annotations: annotations,
	}
}

func federatedObjectName(
	sourceObj ezkube.ClusterResourceId,
) string {
	return sourceObj.GetName() + "-" + sourceObj.GetClusterName()
}

// ownership label defaults to current namespace to allow multiple SMH tenancy within a cluster.
func TranslatedObjectLabels() map[string]string {
	return map[string]string{OwnershipLabelKey: defaults.GetPodNamespace()}
}
