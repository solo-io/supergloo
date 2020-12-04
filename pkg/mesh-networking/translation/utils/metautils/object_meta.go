package metautils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// the key used to differentiate translated resources by
	// the GlooMesh instance which produced them
	OwnershipLabelKey = fmt.Sprintf("owner.%s", v1alpha2.SchemeGroupVersion.Group)

	// Annotation key indicating that the resource configures a federated traffic target
	FederationLabelKey = fmt.Sprintf("federation.%s", v1alpha2.SchemeGroupVersion.Group)

	// Annotation key for tracking the parent resources that were translated in the creation of a child resource
	ParentLabelkey = fmt.Sprintf("parents.%s", v1alpha2.SchemeGroupVersion.Group)
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

// ownership label defaults to current namespace to allow multiple GlooMesh tenancy within a cluster.
func TranslatedObjectLabels() map[string]string {
	return map[string]string{OwnershipLabelKey: defaults.GetPodNamespace()}
}

// add a parent to the annotation for a given child object
func AppendParent(
	ctx context.Context,
	child metav1.Object,
	parentId ezkube.ResourceId,
	parentGVK schema.GroupVersionKind,
) {
	annotations := child.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	parentsAnnotation := make(map[string][]*v1.ObjectRef)
	if paStr, ok := annotations[ParentLabelkey]; ok {
		if err := json.Unmarshal([]byte(paStr), &parentsAnnotation); err != nil {
			contextutils.LoggerFrom(ctx).Errorf("internal error: could not unmarshal %q annotation", ParentLabelkey)
			return
		}
	}

	typeParents, ok := parentsAnnotation[parentGVK.String()]
	if !ok {
		typeParents = make([]*v1.ObjectRef, 0, 1)
	}
	parentRef := ezkube.MakeObjectRef(parentId)
	for _, parent := range typeParents {
		if parent.Equal(parentRef) {
			return
		}
	}
	parentsAnnotation[parentGVK.String()] = append(typeParents, parentRef)

	b, err := json.Marshal(parentsAnnotation)
	if err != nil {
		contextutils.LoggerFrom(ctx).Error("internal error: could not marshal parents map")
		return
	}

	annotations[ParentLabelkey] = string(b)
	child.SetAnnotations(annotations)
}
