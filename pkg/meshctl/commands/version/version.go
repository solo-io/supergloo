package version

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display the version of meshctl and installed Gloo Mesh components",
		RunE: func(cmd *cobra.Command, args []string) error {
			return printVersion(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", "gloo-mesh", "namespace that glooo mesh components are deployed to")
}

type versionInfo struct {
	Client clientVersion   `json:"client"`
	Server []serverVersion `json:"server"`
}
type clientVersion struct {
	Version string `json:"version"`
}
type serverVersion struct {
	Namespace  string      `json:"namespace"`
	Components []component `json:"components"`
}
type component struct {
	ComponentName string           `json:"componentName"`
	Images        []componentImage `json:"images"`
}
type componentImage struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

func printVersion(ctx context.Context, opts *options) error {
	serverVersions := makeServerVersions(ctx, opts)
	versions := versionInfo{
		Client: clientVersion{Version: version.Version},
		Server: serverVersions,
	}

	bytes, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

func makeServerVersions(ctx context.Context, opts *options) []serverVersion {
	kubeClient, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to Kubernetes: %s\n", err.Error())
		return nil
	}
	deploymentClient := appsv1.NewDeploymentClient(kubeClient)
	deployments, err := deploymentClient.ListDeployment(ctx, &client.ListOptions{Namespace: opts.namespace})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to list deployments: %s\n", err.Error())
		return nil
	}

	// map of namespace to list of components
	componentMap := make(map[string][]component)
	for _, deployment := range deployments.Items {
		images, err := getImages(&deployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to pull image information for %s: %s\n", deployment.Name, err.Error())
			continue
		}
		if len(images) == 0 {
			continue
		}

		namespace := deployment.GetObjectMeta().GetNamespace()
		componentMap[namespace] = append(
			componentMap[namespace],
			component{
				ComponentName: deployment.GetName(),
				Images:        images,
			},
		)
	}

	// convert to output format
	var serverVersions []serverVersion
	for namespace, components := range componentMap {
		serverVersions = append(serverVersions, serverVersion{Namespace: namespace, Components: components})
	}

	return serverVersions
}

func getImages(deployment *v1.Deployment) ([]componentImage, error) {
	images := make([]componentImage, len(deployment.Spec.Template.Spec.Containers))
	for i, container := range deployment.Spec.Template.Spec.Containers {
		parsedImage, err := dockerutils.ParseImageName(container.Image)
		if err != nil {
			return nil, err
		}
		imageVersion := parsedImage.Tag
		if parsedImage.Digest != "" {
			imageVersion = parsedImage.Digest
		}

		images[i] = componentImage{
			Name:    container.Name,
			Domain:  parsedImage.Domain,
			Path:    parsedImage.Path,
			Version: imageVersion,
		}
	}

	return images, nil
}
