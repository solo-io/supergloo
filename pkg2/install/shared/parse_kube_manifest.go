package shared

import (
	"strings"

	"github.com/ghodss/yaml"
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
	"k8s.io/apimachinery/pkg/runtime"
)

type UntypedKubeObject map[string]interface{}
type KubeObjectList []runtime.Object

func ParseKubeManifest(manifest string) (KubeObjectList, error) {
	snippets := strings.Split(manifest, "---")
	var objs KubeObjectList
	for _, objectYaml := range snippets {
		parsedObjs, err := parseobjectYaml(objectYaml)
		if err != nil {
			return nil, err
		}
		if parsedObjs == nil {
			continue
		}
		objs = append(objs, parsedObjs...)
	}
	return objs, nil
}

func parseobjectYaml(objectYaml string) (KubeObjectList, error) {
	obj, err := convertYamlToResource(objectYaml)
	if err != nil {
		return nil, errors.Wrapf(err, "unsupported object type: %v", objectYaml)
	}

	return obj, nil
}

func convertYamlToResource(objectYaml string) (KubeObjectList, error) {
	var untyped UntypedKubeObject
	if err := yaml.Unmarshal([]byte(objectYaml), &untyped); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling %v", objectYaml)
	}
	// yaml was empty
	if untyped == nil {
		return nil, nil
	}

	// need to be done manually as the go structs are embedded
	var typeMeta metav1.TypeMeta
	if err := yaml.Unmarshal([]byte(objectYaml), &typeMeta); err != nil {
		return nil, errors.Wrapf(err, "parsing raw yaml as %+v", typeMeta)
	}

	kind := typeMeta.Kind

	var obj runtime.Object
	switch kind {
	case "List":
		return convertUntypedList(untyped)
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
		return nil, errors.Errorf("unsupported kind %v", kind)
	}
	if err := yaml.Unmarshal([]byte(objectYaml), obj); err != nil {
		return nil, errors.Wrapf(err, "parsing raw yaml as %+v", obj)
	}
	return KubeObjectList{obj}, nil
}
func convertUntypedList(untyped UntypedKubeObject) (KubeObjectList, error) {
	itemsValue, ok := untyped["items"]
	if !ok {
		return nil, errors.Errorf("list object missing items")
	}
	items, ok := itemsValue.([]interface{})
	if !ok {
		return nil, errors.Errorf("items must be an array")
	}

	var returnList KubeObjectList
	for _, item := range items {
		itemYaml, err := yaml.Marshal(item)
		if err != nil {
			return nil, errors.Wrapf(err, "marshalling item yaml")
		}
		s := string(itemYaml)
		obj, err := convertYamlToResource(s)
		if err != nil {
			return nil, errors.Wrapf(err, "converting resource in list")
		}
		returnList = append(returnList, obj...)
	}
	return returnList, nil
}
