package virtualmesh

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
		Use:     "virtualmesh",
		Short:   "Description of managed virtual meshes",
		Aliases: []string{"virtualmeshes"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeVirtualMeshes(ctx, c, opts.searchTerms)
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
	flags.StringSliceVarP(&o.searchTerms, "search", "s", []string{}, "A list of terms to match virtual mesh names against")
}

func describeVirtualMeshes(ctx context.Context, c client.Client, searchTerms []string) (string, error) {
	virtualMeshClient := networkingv1alpha2.NewVirtualMeshClient(c)
	virtualMeshList, err := virtualMeshClient.ListVirtualMesh(ctx)
	if err != nil {
		return "", err
	}
	var virtualMeshDescriptions []virtualMeshDescription
	for _, virtualMesh := range virtualMeshList.Items {
		virtualMesh := virtualMesh // pike
		if matchVirtualMesh(virtualMesh, searchTerms) {
			virtualMeshDescriptions = append(virtualMeshDescriptions, describeVirtualMesh(&virtualMesh))
		}
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Meshes"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range virtualMeshDescriptions {
		table.Append([]string{
			description.Metadata.string(),
			printing.FormattedObjectRefs(description.Meshes),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (m virtualMeshMetadata) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Name", m.Name))
	s.WriteString(printing.FormattedField("Namespace", m.Namespace))
	s.WriteString(printing.FormattedField("Cluster", m.Cluster))
	s.WriteString(printing.FormattedField("Global Access Policy", m.GlobalAccessPolicy))
	return s.String()
}

type virtualMeshDescription struct {
	Metadata *virtualMeshMetadata
	Meshes   []*v1.ObjectRef
}

type virtualMeshMetadata struct {
	Name               string
	Namespace          string
	Cluster            string
	GlobalAccessPolicy string
}

func matchVirtualMesh(virtualMesh networkingv1alpha2.VirtualMesh, searchTerms []string) bool {
	// do not apply matching when there are no search strings
	if len(searchTerms) == 0 {
		return true
	}

	for _, s := range searchTerms {
		if strings.Contains(virtualMesh.Name, s) {
			return true
		}
	}

	return false
}

func describeVirtualMesh(virtualMesh *networkingv1alpha2.VirtualMesh) virtualMeshDescription {
	virtualMeshMeta := getVirtualMeshMetadata(virtualMesh)

	var meshes []*v1.ObjectRef
	for _, m := range virtualMesh.Spec.GetMeshes() {
		meshes = append(meshes, m)
	}

	return virtualMeshDescription{
		Metadata: &virtualMeshMeta,
		Meshes:   meshes,
	}
}

func getVirtualMeshMetadata(virtualMesh *networkingv1alpha2.VirtualMesh) virtualMeshMetadata {
	return virtualMeshMetadata{
		Name:               virtualMesh.Name,
		Namespace:          virtualMesh.Namespace,
		Cluster:            virtualMesh.ClusterName,
		GlobalAccessPolicy: virtualMesh.Spec.GlobalAccessPolicy.String(),
	}
}
