package list

import (
	"context"
	"fmt"

	errors "github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	multicluster_solo_io_v1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterSecretType = "solo.io/kubeconfig"
)

type listOptions struct {
	kubeconfig  string
	kubecontext string
	namespace   string
}

func Command(ctx context.Context) *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all Kubernetes cluster registered with Gloo Mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			communityClusters, err := listCommunityClusters(ctx, opts)
			if err != nil {
				return err
			}
			if len(communityClusters) > 0 {
				fmt.Println("Registered community clusters:")
				for _, cluster := range communityClusters {
					fmt.Printf("- %s\n", cluster)
				}
				fmt.Println()
			}

			enterpriseClusters, err := listEnterpriseClusters(ctx, opts)
			if err != nil {
				return err
			}
			if len(enterpriseClusters) > 0 {
				fmt.Println("Registered enterprise clusters:")
				for _, cluster := range enterpriseClusters {
					fmt.Printf("- %s\n", cluster)
				}
				fmt.Println()
			}
			return nil
		},
	}

	opts.addToFlags(cmd.Flags())
	return cmd
}

func listCommunityClusters(ctx context.Context, opts *listOptions) ([]string, error) {
	var communityClusters []string
	secretClient := helpers.MustKubeClient().CoreV1().Secrets(opts.namespace)
	secrets, err := secretClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return communityClusters, errors.Wrapf(err, "Failed to list clusters.")
	}
	for _, s := range secrets.Items {
		if string(s.Type) == clusterSecretType {
			communityClusters = append(communityClusters, s.GetName())
		}
	}
	return communityClusters, nil
}

func listEnterpriseClusters(ctx context.Context, opts *listOptions) ([]string, error) {
	var enterpriseClusters []string
	client, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
	if err != nil {
		return enterpriseClusters, err
	}
	kubernetesClusterList, err := multicluster_solo_io_v1alpha1.NewKubernetesClusterClient(client).ListKubernetesCluster(ctx)
	if err != nil {
		return enterpriseClusters, errors.Wrapf(err, "Failed to list clusters.")
	}
	for _, c := range kubernetesClusterList.Items {
		enterpriseClusters = append(enterpriseClusters, c.GetName())
	}
	return enterpriseClusters, nil
}

func (o *listOptions) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Gloo Mesh is installed in")
}
