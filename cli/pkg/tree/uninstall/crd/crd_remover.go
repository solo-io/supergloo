package crd_uninstall

import (
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	kubernetes_apiext "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apiext"
	"k8s.io/client-go/rest"
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
	crdClientFactory kubernetes_apiext.CrdClientFactory,
) CrdRemover {
	return &crdRemover{
		crdClientFactory: crdClientFactory,
	}
}

type crdRemover struct {
	crdClientFactory kubernetes_apiext.CrdClientFactory
}

func (c *crdRemover) RemoveZephyrCrds(clusterName string, remoteKubeConfig *rest.Config) (crdsDeleted bool, err error) {
	crdClient, err := c.crdClientFactory(remoteKubeConfig)
	if err != nil {
		return false, FailedToBuildCrdClient(err, clusterName)
	}

	crds, err := crdClient.List()
	if err != nil {
		return false, FailedToListCrds(err, clusterName)
	}

	foundZephyrResources := false
	for _, crd := range crds.Items {
		if strings.HasSuffix(crd.GetName(), cliconstants.ServiceMeshHubApiGroupSuffix) {
			foundZephyrResources = true
			err := crdClient.Delete(crd.GetName())
			if err != nil {
				return foundZephyrResources, FailedToDeleteCrd(err, clusterName, crd.GetName())
			}
		}
	}

	return foundZephyrResources, nil
}
