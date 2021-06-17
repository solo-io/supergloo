package utils

import (
	"context"
	"io"
	"strings"

	"github.com/rotisserie/eris"
	errors "github.com/rotisserie/eris"
	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/schemes"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	// required import to enable kube client-go auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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

func BuildClientset(kubeConfigPath, kubeContext string) (*kubernetes.Clientset, error) {
	cfg, err := kubeconfig.GetRestConfigWithContext(kubeConfigPath, kubeContext, "")
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube config")
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "getting kube clientset")
	}
	return kubeClient, nil
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

func Zip(fs afero.Fs, dir string, file string) error {
	tarball, err := fs.Create(file)
	if err != nil {
		return err
	}
	if err := tarutils.Tar(dir, fs, tarball); err != nil {
		return err
	}
	_, err = tarball.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}
