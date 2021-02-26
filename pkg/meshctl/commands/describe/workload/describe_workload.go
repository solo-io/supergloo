package workload

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/describe/printing"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:     "workload",
		Short:   "Description of workloads",
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
	opts.addToFlags(cmd.Flags())

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
	workloadClient := discoveryv1.NewWorkloadClient(c)
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
	table.SetHeader([]string{"Metadata", "Kubernetes", "Mesh"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range workloadDescriptions {
		table.Append([]string{
			printing.FormattedClusterObjectRef(description.Metadata),
			description.Kubernetes.string(),
			printing.FormattedObjectRef(description.Mesh),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (k workloadKubernetes) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Service Account", k.ServiceAccount))

	if len(k.PodLabels) > 0 {
		s.WriteString("\nPOD LABELS\n")
		for label, value := range k.PodLabels {
			s.WriteString(printing.FormattedField(label, value))
		}
	}

	s.WriteString("\nCONTROLLER\n")
	s.WriteString(printing.FormattedClusterObjectRef(k.Controller))

	return s.String()
}

type workloadDescription struct {
	Metadata   *v1.ClusterObjectRef
	Kubernetes *workloadKubernetes
	Mesh       *v1.ObjectRef
}

type workloadKubernetes struct {
	ServiceAccount string
	PodLabels      map[string]string
	Controller     *v1.ClusterObjectRef
}

func matchWorkload(workload discoveryv1.Workload, searchTerms []string) bool {
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

func describeWorkload(workload *discoveryv1.Workload) workloadDescription {
	workloadMeta := getWorkloadMetadata(workload)
	workloadKubernetes := getWorkloadKubernetes(workload)
	return workloadDescription{
		Metadata:   &workloadMeta,
		Kubernetes: &workloadKubernetes,
		Mesh:       workload.Spec.Mesh,
	}
}

func getWorkloadMetadata(workload *discoveryv1.Workload) v1.ClusterObjectRef {
	return v1.ClusterObjectRef{
		Name:        workload.Name,
		Namespace:   workload.Namespace,
		ClusterName: workload.ClusterName,
	}
}

func getWorkloadKubernetes(workload *discoveryv1.Workload) workloadKubernetes {
	return workloadKubernetes{
		ServiceAccount: workload.Spec.GetKubernetes().ServiceAccountName,
		PodLabels:      workload.Spec.GetKubernetes().GetPodLabels(),
		Controller:     workload.Spec.GetKubernetes().Controller,
	}
}
