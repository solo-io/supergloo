package kubeinstall

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CallbackOptions interface {
	PreInstall() error
	PostInstall() error
	PreCreate(res *unstructured.Unstructured) error
	PostCreate(res *unstructured.Unstructured) error
	PreUpdate(res *unstructured.Unstructured) error
	PostUpdate(res *unstructured.Unstructured) error
	PreDelete(res *unstructured.Unstructured) error
	PostDelete(res *unstructured.Unstructured) error
}
