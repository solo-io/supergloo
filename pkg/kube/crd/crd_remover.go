package crd_uninstall

import (
	"context"
	"strings"

	"github.com/rotisserie/eris"
	kubernetes_apiext "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apiextensions.k8s.io/v1beta1"
	"github.com/solo-io/service-mesh-hub/pkg/constants"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	FailedToBuildCrdClient = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to build CRD client for cluster %s", clusterName)
	}
	FailedToListCrds = func(err error, clusterName string) error {
		return eris.Wrapf(err, "Failed to list CRDs for cluster %s", clusterName)
	}
	FailedToDeleteCrd = func(err error, clusterName, crdName string) error {
		return eris.Wrapf(err, "Failed to delete CRD %s on cluster %s", crdName, clusterName)
	}
)

func NewCrdRemover(
	crdClientFactory kubernetes_apiext.CustomResourceDefinitionClientFromConfigFactory,
) CrdRemover {
	return &crdRemover{
		crdClientFactory: crdClientFactory,
	}
}

type crdRemover struct {
	crdClientFactory kubernetes_apiext.CustomResourceDefinitionClientFromConfigFactory
}

func (c *crdRemover) RemovesmhCrds(ctx context.Context, clusterName string, remoteKubeConfig *rest.Config) (crdsDeleted bool, err error) {
	return c.removeCrdsWithNameSuffix(ctx, clusterName, remoteKubeConfig, constants.ServiceMeshHubApiGroupSuffix)
}

func (c *crdRemover) RemoveCrdGroup(ctx context.Context, clusterName string, remoteKubeConfig *rest.Config, groupVersion schema.GroupVersion) (crdsDeleted bool, err error) {
	return c.removeCrdsWithNameSuffix(ctx, clusterName, remoteKubeConfig, groupVersion.Group)
}

// remove all CRDs with the given name suffix
func (c *crdRemover) removeCrdsWithNameSuffix(ctx context.Context, clusterName string, remoteKubeConfig *rest.Config, crdNameSuffix string) (crdsDeleted bool, err error) {
	crdClient, err := c.crdClientFactory(remoteKubeConfig)
	if err != nil {
		return false, FailedToBuildCrdClient(err, clusterName)
	}

	crds, err := crdClient.ListCustomResourceDefinition(ctx)
	if err != nil {
		return false, FailedToListCrds(err, clusterName)
	}

	for _, crd := range crds.Items {
		if strings.HasSuffix(crd.GetName(), crdNameSuffix) {
			crdsDeleted = true
			existing, err := crdClient.GetCustomResourceDefinition(ctx, client.ObjectKey{Name: crd.GetName()})
			if err != nil {
				if errors.IsNotFound(err) {
					// may be a race condition elsewhere; the CRD was removed between when we listed it and when we came here to actually delete it
					continue
				}
				return false, FailedToDeleteCrd(err, clusterName, crd.GetName())
			}
			err = crdClient.DeleteCustomResourceDefinition(ctx, client.ObjectKey{Name: existing.GetName()})
			if err != nil {
				return false, FailedToDeleteCrd(err, clusterName, crd.GetName())
			}
		}
	}

	return crdsDeleted, nil
}
