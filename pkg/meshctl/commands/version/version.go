package version

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	extv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/apps/v1"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display the version of meshctl and installed Service Mesh Hub components",
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
	ComponentName string         `json:"componentName"`
	Image         componentImage `json:"image"`
}
type componentImage struct {
	Domain  string `json:"domain"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

const (
	appLabelKey         = "app"
	imageMatchSubstring = "service-mesh-hub"
)

func getImage(deployment *v1.Deployment) (*componentImage, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if strings.Contains(container.Image, imageMatchSubstring) {
			parsedImage, err := dockerutils.ParseImageName(container.Image)
			if err != nil {
				return nil, err
			}
			imageVersion := parsedImage.Tag
			if parsedImage.Digest != "" {
				imageVersion = parsedImage.Digest
			}
			return &componentImage{Domain: parsedImage.Domain, Path: parsedImage.Path, Version: imageVersion}, nil
		}
	}
	return nil, nil
}

func printVersion(ctx context.Context, opts *options) error {
	kubeClient, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
	if err != nil {
		return err
	}
	deploymentClient := extv1.NewDeploymentClient(kubeClient)
	deployments, err := deploymentClient.ListDeployment(ctx)
	if err != nil {
		return err
	}

	// map of namespace to list of components
	componentMap := make(map[string][]component)
	for _, deployment := range deployments.Items {
		image, err := getImage(&deployment)
		if err != nil {
			return err
		}
		if image != nil {
			namespace := deployment.GetObjectMeta().GetNamespace()
			componentName := deployment.GetObjectMeta().GetLabels()[appLabelKey]
			componentMap[namespace] = append(componentMap[namespace], component{ComponentName: componentName, Image: *image})
		}
	}

	// convert to output format
	var serverVersions []serverVersion
	for namespace, components := range componentMap {
		serverVersions = append(serverVersions, serverVersion{Namespace: namespace, Components: components})
	}
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

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
}
