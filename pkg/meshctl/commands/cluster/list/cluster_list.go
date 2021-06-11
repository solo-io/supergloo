package list

import (
	"context"
	"os"

	"github.com/olekukonko/tablewriter"
	errors "github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	multicluster_solo_io_v1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Cluster", "Namespace"})
			table.SetRowLine(true)
			table.SetAutoWrapText(false)

			kubeClient, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			kubernetesClusterList, err := multicluster_solo_io_v1alpha1.NewKubernetesClusterClient(kubeClient).
				ListKubernetesCluster(ctx, &client.ListOptions{Namespace: opts.namespace})
			if err != nil {
				return errors.Wrapf(err, "Failed to list clusters.")
			}
			for _, c := range kubernetesClusterList.Items {
				table.Append([]string{
					c.GetName(),
					c.GetNamespace(),
				})
			}
			table.Render()
			return nil
		},
	}

	opts.addToFlags(cmd.Flags())
	return cmd
}

func (o *listOptions) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Gloo Mesh is installed in")
}
