package operator

import (
	"context"

	"github.com/hashicorp/go-multierror"
	k8s_apps_v1_clients "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/kube/unstructured"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewOperatorDao(
	ctx context.Context,
	unstructuredKubeClient unstructured.UnstructuredKubeClient,
	deploymentClient k8s_apps_v1_clients.DeploymentClient,
	imageNameParser docker.ImageNameParser,
) OperatorDao {
	return &operatorDao{
		ctx:                    ctx,
		unstructuredKubeClient: unstructuredKubeClient,
		deploymentClient:       deploymentClient,
		imageNameParser:        imageNameParser,
	}
}

type operatorDao struct {
	ctx                    context.Context
	unstructuredKubeClient unstructured.UnstructuredKubeClient
	deploymentClient       k8s_apps_v1_clients.DeploymentClient
	imageNameParser        docker.ImageNameParser
}

func (o *operatorDao) ApplyManifest(installationNamespace, manifest string) error {
	resources, err := o.unstructuredKubeClient.BuildResources(installationNamespace, manifest)
	if err != nil {
		return FailedToParseInstallManifest(err)
	}

	createdResources, installErr := o.unstructuredKubeClient.Create(installationNamespace, resources)
	if installErr != nil {
		_, deleteErr := o.unstructuredKubeClient.Delete(installationNamespace, createdResources)
		if deleteErr != nil {
			var multiErr *multierror.Error
			multiErr = multierror.Append(multiErr, FailedToInstallOperator(installErr))
			multiErr = multierror.Append(multiErr, FailedToCleanFailedInstallation(deleteErr))

			return multiErr.ErrorOrNil()
		}

		return FailedToInstallOperator(installErr)
	}

	return nil
}

func (o *operatorDao) FindOperatorDeployment(name, namespace string) (*k8s_apps_v1.Deployment, error) {
	deployment, err := o.deploymentClient.GetDeployment(o.ctx, client.ObjectKey{Name: name, Namespace: namespace})
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return deployment, nil
}
