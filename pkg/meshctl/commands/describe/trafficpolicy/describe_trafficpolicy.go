package trafficpolicy

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
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
		Use:     "trafficpolicy",
		Short:   "Description of traffic policies",
		Aliases: []string{"trafficpolicies"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeTrafficPolicies(ctx, c, opts.searchTerms)
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
	flags.StringSliceVarP(&o.searchTerms, "search", "s", []string{}, "A list of terms to match traffic policy names against")
}

func describeTrafficPolicies(ctx context.Context, c client.Client, searchTerms []string) (string, error) {
	trafficPolicyClient := networkingv1.NewTrafficPolicyClient(c)
	trafficPolicyList, err := trafficPolicyClient.ListTrafficPolicy(ctx)
	if err != nil {
		return "", err
	}
	var trafficPolicyDescriptions []trafficPolicyDescription
	for _, trafficPolicy := range trafficPolicyList.Items {
		trafficPolicy := trafficPolicy // pike
		if matchTrafficPolicy(trafficPolicy, searchTerms) {
			trafficPolicyDescriptions = append(trafficPolicyDescriptions, describeTrafficPolicy(&trafficPolicy))
		}
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Source_Workloads", "Destination_Services", "HTTP_Matchers"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range trafficPolicyDescriptions {
		table.Append([]string{
			printing.FormattedClusterObjectRef(description.Metadata),
			formattedWorkloadSelectors(description.SourceWorkloads),
			printing.FormattedClusterObjectRefs(description.DestinationServices),
			formattedHttpMatchers(description.HttpMatchers),
		})
	}
	table.Render()

	return buf.String(), nil
}

func formattedWorkloadSelectors(sels []*commonv1.WorkloadSelector) string {
	if len(sels) < 1 {
		return ""
	}
	var s strings.Builder
	for i, sel := range sels {
		sel := sel.GetKubeWorkloadMatcher()
		s.WriteString(printing.FormattedField("Namespaces", strings.Join(sel.Namespaces, ", ")))
		s.WriteString(printing.FormattedField("Clusters", strings.Join(sel.Clusters, ", ")))
		s.WriteString("LABELS\n")
		for label, val := range sel.GetLabels() {
			s.WriteString(printing.FormattedField(label, val))
		}
		if i < len(sels)-1 {
			s.WriteString("\n")
		}
	}
	return s.String()
}

func formattedHttpMatchers(sels []*networkingv1.HttpMatcher) string {
	if len(sels) < 1 {
		return ""
	}
	var s strings.Builder
	for i, matcher := range sels {
		uri := matcher.GetUri()
		s.WriteString(printing.FormattedField("Prefix", uri.GetPrefix()))
		s.WriteString(printing.FormattedField("Exact", uri.GetExact()))
		s.WriteString(printing.FormattedField("Regex", uri.GetRegex()))
		s.WriteString(printing.FormattedField("Method", matcher.GetMethod()))
		s.WriteString("HEADERS\n")
		for _, header := range matcher.GetHeaders() {
			val := header.Value
			if header.Regex {
				val += " (regex)"
			}
			s.WriteString(printing.FormattedField(header.Name, val))
		}
		s.WriteString("QUERY PARAMETERS\n")
		for _, param := range matcher.GetQueryParameters() {
			val := param.Value
			if param.Regex {
				val += " (regex)"
			}
			s.WriteString(printing.FormattedField(param.Name, val))
		}
		if i < len(sels)-1 {
			s.WriteString("\n")
		}
	}
	return s.String()
}

type trafficPolicyDescription struct {
	Metadata            *v1.ClusterObjectRef
	SourceWorkloads     []*commonv1.WorkloadSelector
	DestinationServices []*v1.ClusterObjectRef
	HttpMatchers        []*networkingv1.HttpMatcher
}

func matchTrafficPolicy(trafficPolicy networkingv1.TrafficPolicy, searchTerms []string) bool {
	// do not apply matching when there are no search strings
	if len(searchTerms) == 0 {
		return true
	}

	for _, s := range searchTerms {
		if strings.Contains(trafficPolicy.Name, s) {
			return true
		}
	}

	return false
}

func describeTrafficPolicy(trafficPolicy *networkingv1.TrafficPolicy) trafficPolicyDescription {
	trafficPolicyMeta := getTrafficPolicyMetadata(trafficPolicy)

	var destinationServices []*v1.ClusterObjectRef
	for _, sel := range trafficPolicy.Spec.GetDestinationSelector() {
		if svcs := sel.GetKubeServiceRefs(); svcs != nil {
			destinationServices = append(destinationServices, svcs.Services...)
		}
	}

	return trafficPolicyDescription{
		Metadata:            &trafficPolicyMeta,
		SourceWorkloads:     trafficPolicy.Spec.GetSourceSelector(),
		DestinationServices: destinationServices,
		HttpMatchers:        trafficPolicy.Spec.GetHttpRequestMatchers(),
	}
}

func getTrafficPolicyMetadata(trafficPolicy *networkingv1.TrafficPolicy) v1.ClusterObjectRef {
	return v1.ClusterObjectRef{
		Name:        trafficPolicy.Name,
		Namespace:   trafficPolicy.Namespace,
		ClusterName: trafficPolicy.ClusterName,
	}
}
