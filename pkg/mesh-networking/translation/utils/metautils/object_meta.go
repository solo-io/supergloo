package metautils

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// the key used to differentiate translated resources by
	// the GlooMesh instance which produced them
	OwnershipLabelKey = fmt.Sprintf("owner.%s", v1.SchemeGroupVersion.Group)

	// the key used to differentiate translated resources by
	// the GlooMesh Agent instance which produced them
	AgentLabelKey = fmt.Sprintf("agent.%s", v1.SchemeGroupVersion.Group)

	// Annotation key indicating that the resource configures a federated Destination
	FederationLabelKey = fmt.Sprintf("federation.%s", v1.SchemeGroupVersion.Group)

	// Annotation key for tracking the parent resources that were translated in the creation of a child resource
	ParentLabelkey = fmt.Sprintf("parents.%s", v1.SchemeGroupVersion.Group)
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
	meshInstallation *discoveryv1.MeshSpec_MeshInstallation,
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

// Return true if the object is translated by Gloo Mesh
func IsTranslated(object metav1.Object) bool {
	translatedObjectLabels := TranslatedObjectLabels()
	objLabels := object.GetLabels()
	// AreLabelsInWhiteList returns true if whitelist labels are empty, so we need to check for that case
	return len(objLabels) > 0 && labels.AreLabelsInWhiteList(translatedObjectLabels, objLabels)
}

// add a parent to the annotation for a given child object
func AppendParent(
	ctx context.Context,
	child metav1.Object,
	parentId ezkube.ResourceId,
	parentGVK schema.GroupVersionKind,
) {

	if reflect.ValueOf(child).IsNil() {
		return
	}

	annotations := child.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	parentsAnnotation := make(map[string][]*skv2corev1.ObjectRef)
	if paStr, ok := annotations[ParentLabelkey]; ok {
		if err := json.Unmarshal([]byte(paStr), &parentsAnnotation); err != nil {
			contextutils.LoggerFrom(ctx).DPanicf("internal error: could not unmarshal %s annotation", ParentLabelkey)
			return
		}
	}

	typeParents, ok := parentsAnnotation[parentGVK.String()]
	if !ok {
		typeParents = make([]*skv2corev1.ObjectRef, 0, 1)
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

// Retrieve parent objects from child's annotations, returned a mapping from GVK string to *skv2corev1.ObjectRef
func RetrieveParents(
	ctx context.Context,
	child metav1.Object,
) map[string][]*skv2corev1.ObjectRef {
	if reflect.ValueOf(child).IsNil() {
		return nil
	}

	annotations := child.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	parents := make(map[string][]*skv2corev1.ObjectRef)

	if paStr, ok := annotations[ParentLabelkey]; ok {
		if err := json.Unmarshal([]byte(paStr), &parents); err != nil {
			contextutils.LoggerFrom(ctx).DPanicf("internal error: could not unmarshal %s annotation", ParentLabelkey)
			return nil
		}
	}

	return parents
}
