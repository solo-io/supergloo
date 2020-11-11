package workload

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/printing"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Command(ctx context.Context) *cobra.Command {
	opts := new(options)
	cmd := &cobra.Command{
		Use:     "workload",
		Short:   "Description of managed workloads",
		Aliases: []string{"workloads"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeWorkloads(ctx, c, opts.searchTerms)
			if err != nil {
				return err
			}
			fmt.Println(description)
			return nil
		},
	}

	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	searchTerms []string
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringSliceVarP(&o.searchTerms, "search", "s", []string{}, "A list of terms to match workload names against")
}

func describeWorkloads(ctx context.Context, c client.Client, searchTerms []string) (string, error) {
	workloadClient := discoveryv1alpha2.NewWorkloadClient(c)
	workloadList, err := workloadClient.ListWorkload(ctx)
	if err != nil {
		return "", err
	}
	var workloadDescriptions []workloadDescription
	for _, workload := range workloadList.Items {
		workload := workload // pike
		if matchWorkload(workload, searchTerms) {
			workloadDescriptions = append(workloadDescriptions, describeWorkload(&workload))
		}
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Mesh", "Kubernetes_Controller"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range workloadDescriptions {
		table.Append([]string{
			description.Metadata.string(),
			printing.FormattedObjectRef(description.Mesh),
			printing.FormattedClusterObjectRef(description.KubernetesController),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (m workloadMetadata) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Name", m.Name))
	s.WriteString(printing.FormattedField("Namespace", m.Namespace))
	s.WriteString(printing.FormattedField("Cluster", m.Cluster))
	s.WriteString(printing.FormattedField("Kubernetes Service Account", m.KubernetesServiceAccount))
	return s.String()
}

type workloadDescription struct {
	Metadata             *workloadMetadata
	Mesh                 *v1.ObjectRef
	KubernetesController *v1.ClusterObjectRef
}

type workloadMetadata struct {
	Name                     string
	Namespace                string
	Cluster                  string
	KubernetesServiceAccount string
}

func matchWorkload(workload discoveryv1alpha2.Workload, searchTerms []string) bool {
	// do not apply matching when there are no search strings
	if len(searchTerms) == 0 {
		return true
	}

	for _, s := range searchTerms {
		if strings.Contains(workload.Name, s) {
			return true
		}
	}

	return false
}

func describeWorkload(workload *discoveryv1alpha2.Workload) workloadDescription {
	workloadMeta := getWorkloadMetadata(workload)
	return workloadDescription{
		Metadata:             &workloadMeta,
		Mesh:                 workload.Spec.Mesh,
		KubernetesController: workload.Spec.GetKubernetes().Controller,
	}
}

func getWorkloadMetadata(workload *discoveryv1alpha2.Workload) workloadMetadata {
	return workloadMetadata{
		Name:                     workload.Name,
		Namespace:                workload.Namespace,
		Cluster:                  workload.ClusterName,
		KubernetesServiceAccount: workload.Spec.GetKubernetes().ServiceAccountName,
	}
}
