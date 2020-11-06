package metautils

import (
	"fmt"
	"strings"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.gloomesh.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.gloomesh.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// the key used to differentiate translated resources by
	// the SMH instance which produced them
	OwnershipLabelKey = fmt.Sprintf("owner.%s", v1alpha2.SchemeGroupVersion.Group)

	// Annotation key indicating that the resource configures a federated traffic target
	FederationLabelKey = fmt.Sprintf("federation.%s", v1alpha2.SchemeGroupVersion.Group)
)

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
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[FederationLabelKey] = sets.Key(sourceObj)

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
	return strings.Join([]string{sourceObj.GetName(), sourceObj.GetNamespace(), sourceObj.GetClusterName()}, "-")
}

// ownership label defaults to current namespace to allow multiple SMH tenancy within a cluster.
func TranslatedObjectLabels() map[string]string {
	return map[string]string{OwnershipLabelKey: defaults.GetPodNamespace()}
}
