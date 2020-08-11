package version

import (
	"context"
	"encoding/json"
	"fmt"
	extv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
	"github.com/solo-io/service-mesh-hub/pkg/common/version"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/dockerutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
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
	kubeClient, err := buildKubeClient(opts.kubeconfig, opts.kubecontext)
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

// TODO remove this when PR 860 is merged
func (o *options) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&o.kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
}

// TODO remove this when PR 860 is merged
func buildKubeClient(kubeconfig, kubecontext string) (client.Client, error) {
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubecontext}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}
	scheme := scheme.Scheme
	if err := schemes.SchemeBuilder.AddToScheme(scheme); err != nil {
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
