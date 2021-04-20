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

	// if this key exists in an output object's annotations, Gloo Mesh will delete it
	GarbageCollectDirective = "MustGarbageCollect"
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

func MarkForGarbageCollection(child metav1.Object) {
	if reflect.ValueOf(child).IsNil() {
		return
	}

	annotations := child.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[GarbageCollectDirective] = ""
	child.SetAnnotations(annotations)
}

// add parent annotations for a given child object
// overwrites any existing annotations
func AnnotateParents(
	ctx context.Context,
	child metav1.Object,
	parents map[schema.GroupVersionKind][]ezkube.ResourceId,
) {

	if reflect.ValueOf(child).IsNil() {
		return
	}

	annotations := child.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// convert map[schema.GroupVersionKind][]ezkube.ResourceId to map[string][]*skv2corev1.ObjectRef
	parentsMap := map[string][]*skv2corev1.ObjectRef{}
	for gvk, objs := range parents {
		var objRefs []*skv2corev1.ObjectRef
		for _, obj := range objs {
			objRefs = append(objRefs, ezkube.MakeObjectRef(obj))
		}
		parentsMap[gvk.String()] = objRefs
	}

	b, err := json.Marshal(parentsMap)
	if err != nil {
		contextutils.LoggerFrom(ctx).DPanicf("could not marshal parents map: %v", err)
		return
	}

	annotations[ParentLabelkey] = string(b)
	child.SetAnnotations(annotations)
}

// Retrieve parent objects from child's annotations, returned a mapping from GVK string to *skv2corev1.ObjectRef
func RetrieveParents(
	ctx context.Context,
	child metav1.Object,
) map[schema.GroupVersionKind][]*skv2corev1.ObjectRef {
	if reflect.ValueOf(child).IsNil() {
		return nil
	}

	annotations := child.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	parentsMap := make(map[string][]*skv2corev1.ObjectRef)

	if paStr, ok := annotations[ParentLabelkey]; ok {
		if err := json.Unmarshal([]byte(paStr), &parentsMap); err != nil {
			contextutils.LoggerFrom(ctx).DPanicf("could not unmarshal %s annotation", ParentLabelkey)
			return nil
		}
	}

	if len(parentsMap) < 1 {
		contextutils.LoggerFrom(ctx).DPanicf("output object %s is missing expected parent annotations", ezkube.MakeObjectRef(child))
	}

	// convert map[string][]*skv2corev1.ObjectRef to map[schema.GroupVersionKind][]*skv2corev1.ObjectRef
	parents := map[schema.GroupVersionKind][]*skv2corev1.ObjectRef{}
	for gvkString, objRefs := range parentsMap {
		gvk, err := ezkube.ParseGroupVersionKindString(gvkString)
		if err != nil {
			contextutils.LoggerFrom(ctx).DPanicf("internal error: could not parse GVK string %s", gvkString)
			continue
		}
		parents[gvk] = objRefs
	}

	return parents
}
