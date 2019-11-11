package config

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/version"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func CreateRootContext(customCtx context.Context, name string) context.Context {
	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	rootCtx = contextutils.WithLogger(rootCtx, name)
	loggingContext := []interface{}{"version", version.Version}
	rootCtx = contextutils.WithLoggerValues(rootCtx, loggingContext...)
	return rootCtx
}

func CrdsExist(cfg *rest.Config, crds ...crd.Crd) bool {
	for _, resource := range crds {
		if !CrdExists(resource, cfg) {
			return false
		}
	}
	return true
}

func CrdExists(resource crd.Crd, cfg *rest.Config) bool {
	client := apiexts.NewForConfigOrDie(cfg)
	_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(resource.FullName(), v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		return false
	}
	return true
}
