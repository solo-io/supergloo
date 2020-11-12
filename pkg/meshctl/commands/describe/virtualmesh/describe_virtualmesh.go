package virtualmesh

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
		Use:     "virtualmesh",
		Short:   "Description of virtual meshes",
		Aliases: []string{"virtualmeshes", "vmesh", "vmeshes"},
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
	table.SetHeader([]string{"Metadata", "Net", "Meshes"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range virtualMeshDescriptions {
		table.Append([]string{
			printing.FormattedClusterObjectRef(description.Metadata),
			description.Net.string(),
			printing.FormattedObjectRefs(description.Meshes),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (n virtualMeshNet) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Global Access Policy", n.GlobalAccessPolicy))
	return s.String()
}

type virtualMeshDescription struct {
	Metadata *v1.ClusterObjectRef
	Net      *virtualMeshNet
	Meshes   []*v1.ObjectRef
}

type virtualMeshNet struct {
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
	virtualMeshNet := getVirtualMeshNet(virtualMesh)

	var meshes []*v1.ObjectRef
	for _, m := range virtualMesh.Spec.GetMeshes() {
		meshes = append(meshes, m)
	}

	return virtualMeshDescription{
		Metadata: &virtualMeshMeta,
		Net:      &virtualMeshNet,
		Meshes:   meshes,
	}
}

func getVirtualMeshMetadata(virtualMesh *networkingv1alpha2.VirtualMesh) v1.ClusterObjectRef {
	return v1.ClusterObjectRef{
		Name:        virtualMesh.Name,
		Namespace:   virtualMesh.Namespace,
		ClusterName: virtualMesh.ClusterName,
	}
}

func getVirtualMeshNet(virtualMesh *networkingv1alpha2.VirtualMesh) virtualMeshNet {
	return virtualMeshNet{
		GlobalAccessPolicy: virtualMesh.Spec.GlobalAccessPolicy.String(),
	}
}
