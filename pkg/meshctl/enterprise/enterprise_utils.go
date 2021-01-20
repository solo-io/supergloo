package enterprise

import (
	"context"
	"strings"

	extv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
)

// GetEnterpriseNetworkingVersion returns the Enterprise Extender version if the Gloo Mesh Enterprise Extender is
// detected without error and is determined to be installed.
func GetEnterpriseNetworkingVersion(ctx context.Context, kubeConfigPath, kubeContext string) (string, error) {
	kubeClient, err := utils.BuildClient(kubeConfigPath, kubeContext)
	if err != nil {
		return "", err
	}

	deploymentClient := extv1.NewDeploymentClient(kubeClient)
	deployments, err := deploymentClient.ListDeployment(ctx)
	if err != nil {
		return "", err
	}

	for _, deployment := range deployments.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if strings.Contains(container.Image, "enterprise-networking") {
				image, err := dockerutils.ParseImageName(container.Image)
				if err != nil {
					return "", err
				}
				return image.Tag, err
			}
		}
	}
	return "", nil
}
