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

const (
	appLabelKey         = "app"
	imageMatchSubstring = "service-mesh-hub"
)

func getSMHImageVersion(deployment *v1.Deployment) (string, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if strings.Contains(container.Image, imageMatchSubstring) {
			parsedImage, err := dockerutils.ParseImageName(container.Image)
			if err != nil {
				return "", err
			}
			return parsedImage.Tag, nil
		}
	}
	return "", nil
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

	serverComponents := make(map[string]string)
	for _, deployment := range deployments.Items {
		version, err := getSMHImageVersion(&deployment)
		if err != nil {
			return err
		}
		if version != "" {
			component := deployment.GetObjectMeta().GetLabels()[appLabelKey]
			serverComponents[component] = version
		}
	}

	versions := map[string]interface{}{
		"Client Version": version.Version,
		"Server Version": serverComponents,
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
