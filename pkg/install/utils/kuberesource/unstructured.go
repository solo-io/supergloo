package kuberesource

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/solo-io/go-utils/contextutils"

	"github.com/pkg/errors"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	autoscaling "k8s.io/api/autoscaling/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type UnstructuredResources []*unstructured.Unstructured

func (urs UnstructuredResources) Filter(filter func(resource *unstructured.Unstructured) bool) UnstructuredResources {
	var filtered UnstructuredResources
	for _, res := range urs {
		if filter(res) {
			continue
		}
		filtered = append(filtered, res)
	}
	return filtered
}

func (urs UnstructuredResources) WithLabels(targetLabels map[string]string) UnstructuredResources {
	return urs.Filter(func(resource *unstructured.Unstructured) bool {
		return !labels.SelectorFromSet(targetLabels).Matches(labels.Set(resource.GetLabels()))
	})
}

func (urs UnstructuredResources) Sort() UnstructuredResources {
	sorted := make(UnstructuredResources, len(urs))
	copy(sorted, urs)
	sortUnstructured(sorted)
	return sorted
}

func (urs UnstructuredResources) ByKey() UnstructuredResourcesByKey {
	mapped := make(UnstructuredResourcesByKey)
	for _, res := range urs {
		mapped[Key(res)] = res
	}
	return mapped
}

type VersionedResources struct {
	GVK       schema.GroupVersionKind
	Resources UnstructuredResources
}

func (urs UnstructuredResources) GroupedByGVK() []VersionedResources {
	if len(urs) == 0 {
		return nil
	}
	var versionGroups []VersionedResources
addResource:
	for _, res := range urs {
		gvk := res.GroupVersionKind()
		for i, versionGroup := range versionGroups {
			if versionGroup.GVK.String() == gvk.String() {
				versionGroup.Resources = append(versionGroup.Resources, res)
				versionGroups[i] = versionGroup
				continue addResource
			}
		}
		versionGroups = append(versionGroups, VersionedResources{GVK: gvk, Resources: UnstructuredResources{res}})
	}
	sort.SliceStable(versionGroups, func(i, j int) bool {
		return installOrderLess(versionGroups[i].GVK.Kind, versionGroups[j].GVK.Kind)
	})
	for _, group := range versionGroups {
		group.Resources = group.Resources.Sort()
	}
	return versionGroups
}

type UnstructuredResourcesByKey map[ResourceKey]*unstructured.Unstructured

func (urs UnstructuredResourcesByKey) List() UnstructuredResources {
	var list UnstructuredResources
	for _, res := range urs {
		list = append(list, res)
	}
	return list.Sort()
}

type ResourceKey struct {
	Gvk             schema.GroupVersionKind
	Namespace, Name string
}

func Key(obj *unstructured.Unstructured) ResourceKey {
	return ResourceKey{
		Gvk:       obj.GroupVersionKind(),
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func (k ResourceKey) String() string {
	return fmt.Sprintf("%v.%v.%v", k.Gvk.String(), k.Namespace, k.Name)
}

func ConvertToUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	jsn, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var dynamicObj map[string]interface{}
	if err := json.Unmarshal(jsn, &dynamicObj); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{
		Object: dynamicObj,
	}, nil
}

func ConvertUnstructured(res *unstructured.Unstructured) (runtime.Object, error) {
	rawJson, err := json.Marshal(res.Object)
	if err != nil {
		return nil, err
	}

	// need to be done manually as the go structs are embedded
	var typeMeta metav1.TypeMeta
	if err := json.Unmarshal(rawJson, &typeMeta); err != nil {
		return nil, errors.Wrapf(err, "parsing raw yaml as %+v", typeMeta)
	}

	kind := typeMeta.Kind

	var obj runtime.Object
	switch kind {
	case "List":
		return nil, errors.Wrapf(err, "lists currently unsupported")
	case "Namespace":
		obj = &core.Namespace{TypeMeta: typeMeta}
	case "ServiceAccount":
		obj = &core.ServiceAccount{TypeMeta: typeMeta}
	case "ClusterRole":
		obj = &rbac.ClusterRole{TypeMeta: typeMeta}
	case "ClusterRoleBinding":
		obj = &rbac.ClusterRoleBinding{TypeMeta: typeMeta}
	case "Job":
		obj = &batch.Job{TypeMeta: typeMeta}
	case "ConfigMap":
		obj = &core.ConfigMap{TypeMeta: typeMeta}
	case "Service":
		obj = &core.Service{TypeMeta: typeMeta}
	case "Pod":
		obj = &core.Pod{TypeMeta: typeMeta}
	case "Deployment":
		switch typeMeta.APIVersion {
		case "extensions/v1beta1":
			obj = &extensionsv1beta1.Deployment{TypeMeta: typeMeta}
		case "apps/v1":
			obj = &appsv1.Deployment{TypeMeta: typeMeta}
		case "apps/v1beta2":
			obj = &appsv1beta2.Deployment{TypeMeta: typeMeta}
		default:
			return nil, errors.Errorf("unknown api version for deployment: %v", typeMeta.APIVersion)
		}
	case "DaemonSet":
		switch typeMeta.APIVersion {
		case "extensions/v1beta1":
			obj = &extensionsv1beta1.DaemonSet{TypeMeta: typeMeta}
		case "apps/v1":
			obj = &appsv1.DaemonSet{TypeMeta: typeMeta}
		case "apps/v1beta2":
			obj = &appsv1beta2.DaemonSet{TypeMeta: typeMeta}
		default:
			return nil, errors.Errorf("unknown api version for daemon set: %v", typeMeta.APIVersion)
		}
	case "CustomResourceDefinition":
		obj = &apiextensions.CustomResourceDefinition{TypeMeta: typeMeta}
	case "MutatingWebhookConfiguration":
		obj = &v1beta1.MutatingWebhookConfiguration{TypeMeta: typeMeta}
	case "HorizontalPodAutoscaler":
		obj = &autoscaling.HorizontalPodAutoscaler{TypeMeta: typeMeta}
	default:
		return nil, errors.Errorf("cannot convert kind %v", kind)
	}
	if err := json.Unmarshal(rawJson, obj); err != nil {
		return nil, errors.Wrapf(err, "parsing raw yaml as %+v", obj)
	}
	return obj, nil
}

// returns true if there is a difference between the objects, after having zeroed out
// note that this function updates the object for writing, which is fine as it
// zeroes out statuses and generated fields
func Match(ctx context.Context, obj1, obj2 *unstructured.Unstructured) bool {
	patch, err := GetPatch(obj1, obj2)
	if err != nil {
		return false
	}

	if string(patch) == "{}" {
		return true
	}

	contextutils.LoggerFrom(ctx).Infow("objects differ", "diff", string(patch), "original", Key(obj1), "desired", Key(obj2))

	return false
}

func GetPatch(obj1, obj2 *unstructured.Unstructured) ([]byte, error) {
	zeroGeneratedValues(obj1)
	zeroGeneratedValues(obj2)
	jsn1, err := json.Marshal(obj1)
	if err != nil {
		return nil, err
	}
	jsn2, err := json.Marshal(obj2)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreateMergePatch(jsn1, jsn2)
	if err != nil {
		return nil, err
	}

	return patch, nil
}

func Patch(obj *unstructured.Unstructured, patchJson []byte) error {
	doc, err := json.Marshal(obj.Object)
	if err != nil {
		return err
	}
	out, err := jsonpatch.MergePatch(doc, patchJson)
	if err != nil {
		return err
	}

	uncastObj, err := runtime.Decode(unstructured.UnstructuredJSONScheme, out)
	if err != nil {
		return err
	}
	res, ok := uncastObj.(*unstructured.Unstructured)
	if !ok {
		return errors.Errorf("%T expected to be type *unstructured.Unstructured")
	}

	*obj = *res

	return nil
}

func zeroGeneratedValues(obj *unstructured.Unstructured) {
	obj.SetUID("")
	obj.SetResourceVersion("")
	obj.SetGeneration(0)
	obj.SetSelfLink("")
	obj.SetCreationTimestamp(metav1.Time{})
	obj.SetDeletionTimestamp(nil)
	obj.SetDeletionGracePeriodSeconds(nil)
	obj.SetInitializers(nil)
	obj.SetFinalizers(nil)
	obj.SetOwnerReferences(nil)
	obj.SetClusterName("")
	delete(obj.Object, "status")
}
