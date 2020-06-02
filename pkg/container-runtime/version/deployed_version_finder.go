package version

import (
	"context"
	"strings"

	"github.com/rotisserie/eris"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/docker"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	NoOpenSourceDeployment             = eris.New("Could not find any open-source components of Service Mesh Hub")
	FailedToLookUpOpenSourceDeployment = func(err error) error {
		return eris.Wrap(err, "Failed to find any open-source components of Service Mesh Hub")
	}
	FailedToParseImageName = func(err error, imageName string) error {
		return eris.Wrapf(err, "Failed to parse image name: '%s'", imageName)
	}
	FailedToFindContainer = func(containerName, deploymentName string) error {
		return eris.Errorf("Failed to find open-source deployment; could not find container %s in deployment %s", containerName, deploymentName)
	}
)

const meshNetworkingName = "mesh-networking"

func NewDeployedVersionFinder(
	deploymentClient kubernetes_apps.DeploymentClient,
	imageNameParser docker.ImageNameParser,
) DeployedVersionFinder {
	return &deployedVersionFinder{
		deploymentClient: deploymentClient,
		imageNameParser:  imageNameParser,
	}
}

type deployedVersionFinder struct {
	deploymentClient kubernetes_apps.DeploymentClient
	imageNameParser  docker.ImageNameParser
}

func (d *deployedVersionFinder) OpenSourceVersion(ctx context.Context, installNamespace string) (string, error) {
	deployment, err := d.deploymentClient.GetDeployment(ctx, client.ObjectKey{
		Name:      meshNetworkingName,
		Namespace: installNamespace,
	})
	if errors.IsNotFound(err) {
		return "", NoOpenSourceDeployment
	} else if err != nil {
		return "", FailedToLookUpOpenSourceDeployment(err)
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == meshNetworkingName {
			parsedImageName, err := d.imageNameParser.Parse(container.Image)
			if err != nil {
				return "", FailedToParseImageName(err, container.Image)
			}

			version := parsedImageName.Tag
			if version == "" {
				version = parsedImageName.Digest
			}

			return strings.TrimPrefix(version, "v"), nil
		}
	}

	return "", FailedToFindContainer(meshNetworkingName, meshNetworkingName)
}
