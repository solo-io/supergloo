package shared

import (
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CrdsFromManifest(crdManifestYaml string) ([]*v1beta1.CustomResourceDefinition, error) {
	var crds []*v1beta1.CustomResourceDefinition
	crdRuntimeObjects, err := ParseKubeManifest(crdManifestYaml)
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

func CreateCrds(apiExts apiexts.Interface, crds ...*v1beta1.CustomResourceDefinition) error {
	for _, crd := range crds {
		if _, err := apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "failed to create crd: %v", crd)
		}
	}
	return nil
}

func DeleteCrds(apiExts apiexts.Interface, crdNames ...string) error {
	for _, name := range crdNames {
		err := apiExts.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(name, &v1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to delete crd: %v", name)
		}
	}
	return nil
}
