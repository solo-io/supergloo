package kube

import (
	kuberbac "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kubemeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// If you change this interface, you have to rerun mockgen
type RbacClient interface {
	CreateCrbIfNotExist(crbName string, namespaceName string) error
	DeleteCrb(crbName string) error
}

type KubeRbacClient struct {
	kube kubernetes.Interface
}

func NewKubeRbacClient(kube kubernetes.Interface) *KubeRbacClient {
	return &KubeRbacClient{
		kube: kube,
	}
}

func (client *KubeRbacClient) CreateCrbIfNotExist(crbName string, namespaceName string) error {
	_, err := client.kube.RbacV1().ClusterRoleBindings().Create(GetCrb(crbName, namespaceName))
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func GetCrb(crbName string, namespaceName string) *kuberbac.ClusterRoleBinding {
	meta := kubemeta.ObjectMeta{
		Name: crbName,
	}
	subject := kuberbac.Subject{
		Kind:      "ServiceAccount",
		Namespace: namespaceName,
		Name:      "default",
	}
	roleRef := kuberbac.RoleRef{
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
		APIGroup: "rbac.authorization.k8s.io",
	}
	return &kuberbac.ClusterRoleBinding{
		ObjectMeta: meta,
		Subjects:   []kuberbac.Subject{subject},
		RoleRef:    roleRef,
	}
}

func (client *KubeRbacClient) DeleteCrb(crbName string) error {
	return client.kube.RbacV1().ClusterRoleBindings().Delete(crbName, &kubemeta.DeleteOptions{})
}
