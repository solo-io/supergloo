package kubernetes_apiext

import (
	"context"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/version"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type ServerVersionClient interface {
	Get() (*version.Info, error)
}

type CustomResourceDefinitionClient interface {
	Get(ctx context.Context, name string) (*v1beta1.CustomResourceDefinition, error)
	List(ctx context.Context) (*v1beta1.CustomResourceDefinitionList, error)
	Delete(ctx context.Context, crd *v1beta1.CustomResourceDefinition) error
}
