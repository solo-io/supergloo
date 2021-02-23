package accesspolicy

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	networkingv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
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
		Use:     "accesspolicy",
		Short:   "Description of access policies",
		Aliases: []string{"accesspolicies"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeAccessPolicies(ctx, c, opts.searchTerms)
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
	flags.StringSliceVarP(&o.searchTerms, "search", "s", []string{}, "A list of terms to match access policy names against")
}

func describeAccessPolicies(ctx context.Context, c client.Client, searchTerms []string) (string, error) {
	accessPolicyClient := networkingv1alpha2.NewAccessPolicyClient(c)
	accessPolicyList, err := accessPolicyClient.ListAccessPolicy(ctx)
	if err != nil {
		return "", err
	}
	var accessPolicyDescriptions []accessPolicyDescription
	for _, accessPolicy := range accessPolicyList.Items {
		accessPolicy := accessPolicy // pike
		if matchAccessPolicy(accessPolicy, searchTerms) {
			accessPolicyDescriptions = append(accessPolicyDescriptions, describeAccessPolicy(&accessPolicy))
		}
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Filters", "Source_Service_Accounts", "Destination_Services"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range accessPolicyDescriptions {
		table.Append([]string{
			printing.FormattedClusterObjectRef(description.Metadata),
			description.Filters.string(),
			printing.FormattedClusterObjectRefs(description.SourceServiceAccounts),
			printing.FormattedClusterObjectRefs(description.DestinationServices),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (f accessPolicyFilters) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Allowed Paths", strings.Join(f.AllowedPaths, ", ")))
	s.WriteString(printing.FormattedField("Allowed Methods", strings.Join(f.AllowedMethods, ", ")))
	s.WriteString(printing.FormattedField("Allowed Ports", strings.Join(f.AllowedPorts, ", ")))
	return s.String()
}

type accessPolicyDescription struct {
	Metadata              *v1.ClusterObjectRef
	Filters               *accessPolicyFilters
	SourceServiceAccounts []*v1.ClusterObjectRef
	DestinationServices   []*v1.ClusterObjectRef
}

type accessPolicyFilters struct {
	AllowedPaths   []string
	AllowedMethods []string
	AllowedPorts   []string
}

func matchAccessPolicy(accessPolicy networkingv1alpha2.AccessPolicy, searchTerms []string) bool {
	// do not apply matching when there are no search strings
	if len(searchTerms) == 0 {
		return true
	}

	for _, s := range searchTerms {
		if strings.Contains(accessPolicy.Name, s) {
			return true
		}
	}

	return false
}

func describeAccessPolicy(accessPolicy *networkingv1alpha2.AccessPolicy) accessPolicyDescription {
	accessPolicyMeta := getAccessPolicyMetadata(accessPolicy)
	accessPolicyFilters := getAccessPolicyFilters(accessPolicy)

	var sourceServiceAccounts []*v1.ClusterObjectRef
	for _, sel := range accessPolicy.Spec.GetSourceSelector() {
		if svcAccs := sel.GetKubeServiceAccountRefs(); svcAccs != nil {
			sourceServiceAccounts = append(sourceServiceAccounts, svcAccs.ServiceAccounts...)
		}
	}

	var destinationServices []*v1.ClusterObjectRef
	for _, sel := range accessPolicy.Spec.GetDestinationSelector() {
		if svcs := sel.GetKubeServiceRefs(); svcs != nil {
			destinationServices = append(destinationServices, svcs.Services...)
		}
	}

	return accessPolicyDescription{
		Metadata:              &accessPolicyMeta,
		Filters:               &accessPolicyFilters,
		SourceServiceAccounts: sourceServiceAccounts,
		DestinationServices:   destinationServices,
	}
}

func getAccessPolicyMetadata(accessPolicy *networkingv1alpha2.AccessPolicy) v1.ClusterObjectRef {
	return v1.ClusterObjectRef{
		Name:        accessPolicy.Name,
		Namespace:   accessPolicy.Namespace,
		ClusterName: accessPolicy.ClusterName,
	}
}

func getAccessPolicyFilters(accessPolicy *networkingv1alpha2.AccessPolicy) accessPolicyFilters {
	filters := accessPolicyFilters{
		AllowedPaths:   accessPolicy.Spec.AllowedPaths,
		AllowedMethods: make([]string, len(accessPolicy.Spec.GetAllowedMethods())),
		AllowedPorts:   make([]string, len(accessPolicy.Spec.GetAllowedPorts())),
	}

	for i, method := range accessPolicy.Spec.GetAllowedMethods() {
		filters.AllowedMethods[i] = method
	}

	for i, port := range accessPolicy.Spec.GetAllowedPorts() {
		filters.AllowedMethods[i] = fmt.Sprint(port)
	}

	return filters
}
