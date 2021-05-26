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
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all Kubernetes cluster registered with Gloo Mesh",
	}

	cmd.AddCommand(
		communityCommand(ctx),
		enterpriseCommand(ctx),
	)

	return cmd
}

func communityCommand(ctx context.Context) *cobra.Command {
	opts := listOptions{}
	cmd := &cobra.Command{
		Use:     "community",
		Short:   "List registered clusters for Gloo Mesh community edition",
		Example: "  meshctl cluster list community",
		RunE: func(cmd *cobra.Command, args []string) error {
			secretClient := helpers.MustKubeClient().CoreV1().Secrets(opts.namespace)
			secrets, err := secretClient.List(ctx, metav1.ListOptions{})
			if err != nil {
				return errors.Wrapf(err, "Failed to list clusters.")
			}
			for _, s := range secrets.Items {
				if string(s.Type) == clusterSecretType {
					fmt.Printf("%s\n", s.GetName())
				}
			}
			return nil
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

func enterpriseCommand(ctx context.Context) *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:     "enterprise",
		Short:   "List registered clusters for Gloo Mesh enterprise edition",
		Example: " meshctl cluster list enterprise",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			kubernetesClusterList, err := multicluster_solo_io_v1alpha1.NewKubernetesClusterClient(client).ListKubernetesCluster(ctx)
			if err != nil {
				return errors.Wrapf(err, "Failed to list clusters.")
			}
			for _, c := range kubernetesClusterList.Items {
				fmt.Printf("%s\n", c.GetName())
			}
			return nil
		},
	}
	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true

	return cmd
}

func (o *listOptions) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Gloo Mesh is installed in")
}
