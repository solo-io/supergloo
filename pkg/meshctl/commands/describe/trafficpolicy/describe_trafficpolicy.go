package trafficpolicy

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
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
		Use:     "trafficpolicy",
		Short:   "Description of managed traffic policies",
		Aliases: []string{"trafficpolicys"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeTrafficPolicys(ctx, c, opts.searchTerms)
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

func describeTrafficPolicys(ctx context.Context, c client.Client, searchTerms []string) (string, error) {
	trafficPolicyClient := networkingv1alpha2.NewTrafficPolicyClient(c)
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
	table.SetHeader([]string{"Metadata", "Source_Service_Accounts", "Destination_Services"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range trafficPolicyDescriptions {
		table.Append([]string{
			description.Metadata.string(),
			formattedWorkloadSelectors(description.SourceWorkloads),
			printing.FormattedClusterObjectRefs(description.DestinationServices),
			formattedHttpMatchers(description.HttpMatchers),
		})
	}
	table.Render()

	return buf.String(), nil
}

func formattedWorkloadSelectors(sels []*networkingv1alpha2.WorkloadSelector) string {
	if len(sels) < 1 {
		return ""
	}
	var s strings.Builder
	for i, sel := range sels {
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

func formattedHttpMatchers(sels []*networkingv1alpha2.TrafficPolicySpec_HttpMatcher) string {
	if len(sels) < 1 {
		return ""
	}
	var s strings.Builder
	for i, matcher := range sels {
		s.WriteString(printing.FormattedField("Prefix", matcher.GetPrefix()))
		s.WriteString(printing.FormattedField("Exact", matcher.GetExact()))
		s.WriteString(printing.FormattedField("Regex", matcher.GetRegex()))
		s.WriteString(printing.FormattedField("Method", matcher.GetMethod().Method.String()))
		s.WriteString("HEADERS\n")
		for _, header := range matcher.GetHeaders() {
			val := header.Value
			if header.Regex {
				val += " (regex)"
			}
			s.WriteString(printing.FormattedField(header.Name, val))
		}
		s.WriteString("QUERY PARAMETERS")
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

func (m trafficPolicyMetadata) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Name", m.Name))
	s.WriteString(printing.FormattedField("Namespace", m.Namespace))
	s.WriteString(printing.FormattedField("Cluster", m.Cluster))
	return s.String()
}

type trafficPolicyDescription struct {
	Metadata            *trafficPolicyMetadata
	SourceWorkloads     []*networkingv1alpha2.WorkloadSelector
	DestinationServices []*v1.ClusterObjectRef
	HttpMatchers        []*networkingv1alpha2.TrafficPolicySpec_HttpMatcher
}

type trafficPolicyMetadata struct {
	Name      string
	Namespace string
	Cluster   string
}

func matchTrafficPolicy(trafficPolicy networkingv1alpha2.TrafficPolicy, searchTerms []string) bool {
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

func describeTrafficPolicy(trafficPolicy *networkingv1alpha2.TrafficPolicy) trafficPolicyDescription {
	trafficPolicyMeta := getTrafficPolicyMetadata(trafficPolicy)
	var sourceWorkloads []*networkingv1alpha2.WorkloadSelector
	for _, wl := range trafficPolicy.Spec.GetSourceSelector() {
		sourceWorkloads = append(sourceWorkloads, wl)
	}

	var destinationServices []*v1.ClusterObjectRef
	for _, sel := range trafficPolicy.Spec.GetDestinationSelector() {
		if svcs := sel.GetKubeServiceRefs(); svcs != nil {
			destinationServices = append(destinationServices, svcs.Services...)
		}
	}

	return trafficPolicyDescription{
		Metadata:            &trafficPolicyMeta,
		SourceWorkloads:     sourceWorkloads,
		DestinationServices: destinationServices,
		HttpMatchers:        trafficPolicy.Spec.GetHttpRequestMatchers(),
	}
}

func getTrafficPolicyMetadata(trafficPolicy *networkingv1alpha2.TrafficPolicy) trafficPolicyMetadata {
	return trafficPolicyMetadata{
		Name:      trafficPolicy.Name,
		Namespace: trafficPolicy.Namespace,
		Cluster:   trafficPolicy.ClusterName,
	}
}
