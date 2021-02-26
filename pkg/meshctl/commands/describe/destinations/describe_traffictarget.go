package destinations

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
		Use:     "destination",
		Short:   "Description of discovered Destinations",
		Aliases: []string{"destinations"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeDestinations(ctx, c, opts.searchTerms)
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
	flags.StringSliceVarP(&o.searchTerms, "search", "s", []string{}, "A list of terms to match traffic target names against")
}

func describeDestinations(ctx context.Context, c client.Client, searchTerms []string) (string, error) {
	destinationClient := discoveryv1.NewDestinationClient(c)
	destinationList, err := destinationClient.ListDestination(ctx)
	if err != nil {
		return "", err
	}
	var destinationDescriptions []destinationDescription
	for _, destination := range destinationList.Items {
		destination := destination // pike
		if matchDestination(destination, searchTerms) {
			destinationDescriptions = append(destinationDescriptions, describeDestination(&destination))
		}
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Traffic_Policies", "Access_Policies"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range destinationDescriptions {
		table.Append([]string{
			description.Metadata.string(),
			printing.FormattedObjectRefs(description.TrafficPolicies),
			printing.FormattedObjectRefs(description.AccessPolicies),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (m destinationMetadata) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Name", m.Name))
	s.WriteString(printing.FormattedField("Namespace", m.Namespace))
	s.WriteString(printing.FormattedField("Cluster", m.Cluster))
	s.WriteString(printing.FormattedField("Type", m.Type))
	s.WriteString(printing.FormattedField("Federated DNS Name", m.FederatedDnsName))
	return s.String()
}

type destinationDescription struct {
	Metadata        *destinationMetadata
	TrafficPolicies []*v1.ObjectRef
	AccessPolicies  []*v1.ObjectRef
}

type destinationMetadata struct {
	Type              string
	Name              string
	Namespace         string
	Cluster           string
	FederatedDnsName  string
	FederatedToMeshes []*v1.ObjectRef
}

func matchDestination(destination discoveryv1.Destination, searchTerms []string) bool {
	// do not apply matching when there are no search strings
	if len(searchTerms) == 0 {
		return true
	}

	for _, s := range searchTerms {
		if strings.Contains(destination.Name, s) {
			return true
		}
	}

	return false
}

func describeDestination(destination *discoveryv1.Destination) destinationDescription {
	meshMeta := getDestinationMetadata(destination)
	var trafficPolicies []*v1.ObjectRef
	for _, fs := range destination.Status.AppliedTrafficPolicies {
		trafficPolicies = append(trafficPolicies, fs.Ref)
	}

	var accessPolicies []*v1.ObjectRef
	for _, vm := range destination.Status.AppliedAccessPolicies {
		accessPolicies = append(accessPolicies, vm.Ref)
	}

	return destinationDescription{
		Metadata:        &meshMeta,
		TrafficPolicies: trafficPolicies,
		AccessPolicies:  accessPolicies,
	}
}

func getDestinationMetadata(destination *discoveryv1.Destination) destinationMetadata {
	switch destination.Spec.GetType().(type) {
	case *discoveryv1.DestinationSpec_KubeService_:
		kubeServiceRef := destination.Spec.GetKubeService().Ref
		return destinationMetadata{
			Type:              "kubernetes service",
			Name:              kubeServiceRef.Name,
			Namespace:         kubeServiceRef.Namespace,
			Cluster:           kubeServiceRef.ClusterName,
			FederatedDnsName:  destination.Status.GetAppliedFederation().GetFederatedHostname(),
			FederatedToMeshes: destination.Status.GetAppliedFederation().GetFederatedToMeshes(),
		}
	}
	return destinationMetadata{}
}
