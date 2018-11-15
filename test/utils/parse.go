package utils

import (
	"context"
	"github.com/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"gopkg.in/yaml.v2"
	"k8s.io/api/admissionregistration/v1beta1"
	apps "k8s.io/api/apps/v1beta2"
	autoscaling "k8s.io/api/autoscaling/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type UntypedKubeObject map[string]interface{}
type KubeObjectCollection []runtime.Object

func ParseKubeManifest(ctx context.Context, manifest string) (KubeObjectCollection, error) {
	snippets := strings.Split(manifest, "---")
	var objs KubeObjectCollection
	for _, objectYaml := range snippets {
		parsedObjs, err:= parseobjectYaml(ctx, objectYaml)
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

func parseobjectYaml(ctx context.Context, objectYaml string) (KubeObjectCollection, error) {
	var untyped UntypedKubeObject
	if err := yaml.Unmarshal([]byte(objectYaml), &untyped); err != nil {
		return nil, errors.Wrapf(err, "unmarshalling %v", objectYaml)
	}
	// yaml was empty
	if untyped == nil {
		return nil, nil
	}

	kindVal, ok := untyped["kind"]
	if !ok {
		return nil, errors.Errorf("%v missing key 'kind'", untyped)
	}
	kind, ok := kindVal.(string)
	if !ok {
		return nil, errors.Errorf("%v unexpected value for 'kind' in %v", kindVal, untyped)
	}
	obj, err := convertUntypedToResource(kind, objectYaml, untyped)
	if err != nil {
		// TODO (ilackarms): handle this error somewhere
		contextutils.LoggerFrom(ctx).Errorf("%v", errors.Wrapf(err, "parsing resource %v", untyped["metadata"]))
		return nil, nil
	}
	list, ok := obj.(*core.List)
	if !ok {
		for _, item := range list.Items {
			objs = append(objs, item.Object)
		}
	} else {
	}
}

func convertUntypedToResource(kind, rawYaml string, untyped UntypedKubeObject) (runtime.Object, error) {
	var obj runtime.Object
	switch kind {
	case "List":
		items, ok := untyped["items"]
		if !ok {
			return nil, errors.Errorf("list must contain items key")
		}
		itemList, ok := items.([]interface{})
		if !ok {
			return nil, errors.Errorf("list must be an array")
		}
		for _, item := range itemList {
			itemObj, ok := item.(map[string]interface{})
			if !ok {
				if !ok {
					return nil, errors.Errorf("list items must be objects")
				}
			}
			yam, err :=
		}
		obj = &core.List{}
	case "Namespace":
		obj = &core.Namespace{}
	case "ServiceAccount":
		obj = &core.ServiceAccount{}
	case "ClusterRole":
		obj = &rbac.ClusterRole{}
	case "ClusterRoleBinding":
		obj = &rbac.ClusterRoleBinding{}
	case "Job":
		obj = &batch.Job{}
	case "ConfigMap":
		obj = &core.ConfigMap{}
	case "Service":
		obj = &core.Service{}
	case "Deployment":
		obj = &apps.Deployment{}
	case "DaemonSet":
		obj = &apps.DaemonSet{}
	case "CustomResourceDefinition":
		obj = &apiextensions.CustomResourceDefinition{}
	case "MutatingWebhookConfiguration":
		obj = &v1beta1.MutatingWebhookConfiguration{}
	case "HorizontalPodAutoscaler":
		obj = &autoscaling.HorizontalPodAutoscaler{}
	default:
		return nil, errors.Errorf("unsupported kind %v", kind)
	}
	if err := yaml.Unmarshal([]byte(rawYaml), obj); err != nil {
		return nil, errors.Wrapf(err, "parsing raw yaml as %+v", obj)
	}
	return obj, nil
}

type infraSyncer struct {
	kube kubernetes.Interface
	exts apiexts.Interface
}

func (s *infraSyncer) create(obj runtime.Object) error {
	kube := s.kube
	exts := s.exts
	switch obj := obj.(type) {
	case *core.Namespace:
		_, err := kube.CoreV1().Namespaces().Create(obj)
		return err
	case *core.ConfigMap:
		_, err := kube.CoreV1().ConfigMaps(obj.Namespace).Create(obj)
		return err
	case *core.ServiceAccount:
		_, err := kube.CoreV1().ServiceAccounts(obj.Namespace).Create(obj)
		return err
	case *core.Service:
		_, err := kube.CoreV1().Services(obj.Namespace).Create(obj)
		return err
	case *rbac.ClusterRole:
		_, err := kube.RbacV1().ClusterRoles().Create(obj)
		return err
	case *rbac.ClusterRoleBinding:
		_, err := kube.RbacV1().ClusterRoleBindings().Create(obj)
		return err
	case *batch.Job:
		_, err := kube.BatchV1().Jobs(obj.Namespace).Create(obj)
		return err
	case *apps.Deployment:
		_, err := kube.AppsV1beta2().Deployments(obj.Namespace).Create(obj)
		return err
	case *apps.DaemonSet:
		_, err := kube.AppsV1beta2().DaemonSets(obj.Namespace).Create(obj)
		return err
	case *apiextensions.CustomResourceDefinition:
		_, err := exts.ApiextensionsV1beta1().CustomResourceDefinitions().Create(obj)
		return err
	case *v1beta1.MutatingWebhookConfiguration:
		_, err := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(obj)
		return err
	case *autoscaling.HorizontalPodAutoscaler:
		_, err := kube.AutoscalingV1().HorizontalPodAutoscalers(obj.Namespace).Create(obj)
		return err
	}
	return errors.Errorf("no implementation for type %v", obj)
}

// resource version should be ignored / not matter
func (s *infraSyncer) update(obj runtime.Object) error {
	kube := s.kube
	exts := s.exts
	switch obj := obj.(type) {
	case *core.Namespace:
		client := kube.CoreV1().Namespaces()
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *core.ConfigMap:
		client := kube.CoreV1().ConfigMaps(obj.Namespace)
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *core.ServiceAccount:
		client := kube.CoreV1().ServiceAccounts(obj.Namespace)
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *core.Service:
		client := kube.CoreV1().Services(obj.Namespace)
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *rbac.ClusterRole:
		client := kube.RbacV1().ClusterRoles()
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *rbac.ClusterRoleBinding:
		client := kube.RbacV1().ClusterRoleBindings()
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *batch.Job:
		client := kube.BatchV1().Jobs(obj.Namespace)
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *apps.Deployment:
		client := kube.AppsV1beta2().Deployments(obj.Namespace)
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *apps.DaemonSet:
		client := kube.AppsV1beta2().DaemonSets(obj.Namespace)
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *apiextensions.CustomResourceDefinition:
		client := exts.ApiextensionsV1beta1().CustomResourceDefinitions()
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *v1beta1.MutatingWebhookConfiguration:
		client := kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	case *autoscaling.HorizontalPodAutoscaler:
		client := kube.AutoscalingV1().HorizontalPodAutoscalers(obj.Namespace)
		obj2, err := client.Get(obj.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.ResourceVersion = obj2.ResourceVersion
		_, err = client.Update(obj)
		return err
	}
	return errors.Errorf("no implementation for type %v", obj)
}

// this can be just an empty object of the correct type w/ the name and namespace (if applicable) set
func (s *infraSyncer) delete(obj runtime.Object) error {
	kube := s.kube
	exts := s.exts
	switch obj := obj.(type) {
	case *core.Namespace:
		return kube.CoreV1().Namespaces().Delete(obj.Name, nil)
	case *core.ConfigMap:
		return kube.CoreV1().ConfigMaps(obj.Namespace).Delete(obj.Name, nil)
	case *core.ServiceAccount:
		return kube.CoreV1().ServiceAccounts(obj.Namespace).Delete(obj.Name, nil)
	case *core.Service:
		return kube.CoreV1().Services(obj.Namespace).Delete(obj.Name, nil)
	case *rbac.ClusterRole:
		return kube.RbacV1().ClusterRoles().Delete(obj.Name, nil)
	case *rbac.ClusterRoleBinding:
		return kube.RbacV1().ClusterRoleBindings().Delete(obj.Name, nil)
	case *batch.Job:
		return kube.BatchV1().Jobs(obj.Namespace).Delete(obj.Name, nil)
	case *apps.Deployment:
		return kube.AppsV1beta2().Deployments(obj.Namespace).Delete(obj.Name, nil)
	case *apps.DaemonSet:
		return kube.AppsV1beta2().DaemonSets(obj.Namespace).Delete(obj.Name, nil)
	case *apiextensions.CustomResourceDefinition:
		return exts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(obj.Name, nil)
	case *v1beta1.MutatingWebhookConfiguration:
		return kube.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Delete(obj.Name, nil)
	case *autoscaling.HorizontalPodAutoscaler:
		return kube.AutoscalingV1().HorizontalPodAutoscalers(obj.Namespace).Delete(obj.Name, nil)
	}
	return errors.Errorf("no implementation for type %v", obj)
}
