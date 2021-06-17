package utils

import (
	"context"
	"strings"

	"github.com/rotisserie/eris"
	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/schemes"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildClient(kubeConfigPath, kubeContext string) (client.Client, error) {
	cfg, err := kubeconfig.GetRestConfigWithContext(kubeConfigPath, kubeContext, "")
	if err != nil {
		return nil, err
	}

	scheme := scheme.Scheme
	if err := schemes.AddToScheme(scheme); err != nil {
		return nil, err
	}

	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetGlooMeshVersion(ctx context.Context, kubeConfigPath, kubeContext, namespace string) (string, error) {
	kubeClient, err := BuildClient(kubeConfigPath, kubeContext)
	if err != nil {
		return "", err
	}

	deploymentClient := appsv1.NewDeploymentClient(kubeClient)
	deployments, err := deploymentClient.ListDeployment(ctx, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return "", err
	}

	// Find the networking deployment and return the tag of the networking image
	for _, deployment := range deployments.Items {
		if strings.Contains(deployment.Name, "networking") {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				if strings.Contains(container.Name, "networking") {
					image, err := dockerutils.ParseImageName(container.Image)
					if err != nil {
						return "", err
					}

					return image.Tag, err
				}
			}
		}
	}

	return "", eris.New("unable to find Gloo Mesh deployment in management cluster")
}
