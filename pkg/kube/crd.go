package kube

import (
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/pkg/install/shared"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CrdsFromManifest(crdManifestYaml string) ([]*v1beta1.CustomResourceDefinition, error) {
	var crds []*v1beta1.CustomResourceDefinition
	crdRuntimeObjects, err := shared.ParseKubeManifest(crdManifestYaml)
	if err != nil {
		return nil, err
	}
	for _, obj := range crdRuntimeObjects {
		apiExtCrd, ok := obj.(*v1beta1.CustomResourceDefinition)
		if !ok {
			return nil, errors.Wrapf(err, "internal error: crd manifest must only contain CustomResourceDefinitions")
		}
		crds = append(crds, apiExtCrd)
	}
	return crds, nil
}

// If you change this interface, you have to rerun mockgen
type CrdClient interface {
	CreateCrds(crds ...*v1beta1.CustomResourceDefinition) error
	DeleteCrds(crdNames ...string) error
}

type KubeCrdClient struct {
	apiExts apiexts.Interface
}

func NewKubeCrdClient(apiExts apiexts.Interface) *KubeCrdClient {
	return &KubeCrdClient{
		apiExts: apiExts,
	}
}

func (client *KubeCrdClient) CreateCrds(crds ...*v1beta1.CustomResourceDefinition) error {
	for _, crd := range crds {
		if _, err := client.apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "failed to create crd: %v", crd)
		}
	}
	return nil
}

func (client *KubeCrdClient) DeleteCrds(crdNames ...string) error {
	for _, name := range crdNames {
		err := client.apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(name, &v1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to delete crd: %v", name)
		}
	}
	return nil
}
