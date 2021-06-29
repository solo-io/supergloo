package mesh

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
		Use:     "mesh",
		Short:   "Description of managed meshes",
		Aliases: []string{"meshes"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeMeshes(ctx, c, opts.searchTerms)
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
	flags.StringSliceVarP(&o.searchTerms, "search", "s", []string{}, "A list of terms to match mesh names against")
}

func describeMeshes(ctx context.Context, c client.Client, searchTerms []string) (string, error) {
	meshClient := discoveryv1.NewMeshClient(c)
	meshList, err := meshClient.ListMesh(ctx)
	if err != nil {
		return "", err
	}
	var meshDescriptions []meshDescription
	for _, mesh := range meshList.Items {
		mesh := mesh // pike
		if matchMesh(mesh, searchTerms) {
			meshDescriptions = append(meshDescriptions, describeMesh(&mesh))
		}
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Virtual_Mesh", "Virtual_Destinations"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range meshDescriptions {
		table.Append([]string{
			description.Metadata.string(),
			printing.FormattedObjectRef(description.VirtualMesh),
			printing.FormattedObjectRefs(description.VirtualDestinations),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (m meshMetadata) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Name", m.Name))
	s.WriteString(printing.FormattedField("Namespace", m.Namespace))
	s.WriteString(printing.FormattedField("Cluster", m.Cluster))
	s.WriteString(printing.FormattedField("Clusters", strings.Join(m.Clusters, ", ")))
	s.WriteString(printing.FormattedField("Type", m.Type))
	s.WriteString(printing.FormattedField("Region", m.Region))
	s.WriteString(printing.FormattedField("AwsAccountId", m.AwsAccountId))
	s.WriteString(printing.FormattedField("Version", m.Version))
	return s.String()
}

type meshDescription struct {
	Metadata            *meshMetadata
	VirtualMesh         *v1.ObjectRef
	VirtualDestinations []*v1.ObjectRef
}

type meshMetadata struct {
	Name         string
	Namespace    string
	Cluster      string
	Clusters     []string
	Type         string
	Region       string
	AwsAccountId string
	Version      string
}

func matchMesh(mesh discoveryv1.Mesh, searchTerms []string) bool {
	// do not apply matching when there are no search strings
	if len(searchTerms) == 0 {
		return true
	}

	for _, s := range searchTerms {
		if strings.Contains(mesh.Name, s) {
			return true
		}
	}

	return false
}

func describeMesh(mesh *discoveryv1.Mesh) meshDescription {
	meshMeta := getMeshMetadata(mesh)

	var virtualDestinations []*v1.ObjectRef
	for _, fs := range mesh.Status.AppliedVirtualDestinations {
		virtualDestinations = append(virtualDestinations, fs.Ref)
	}

	return meshDescription{
		Metadata:            &meshMeta,
		VirtualMesh:         mesh.Status.GetAppliedVirtualMesh().GetRef(),
		VirtualDestinations: virtualDestinations,
	}
}

func getMeshMetadata(mesh *discoveryv1.Mesh) meshMetadata {
	var meshType string
	if mesh.Spec.GetAwsAppMesh() != nil {
		appmesh := mesh.Spec.GetAwsAppMesh()
		return meshMetadata{
			Type:         "appmesh",
			Name:         appmesh.AwsName,
			Region:       appmesh.Region,
			AwsAccountId: appmesh.AwsAccountId,
			Clusters:     appmesh.Clusters,
		}
	}
	var meshInstallation *discoveryv1.MeshInstallation
	switch mesh.Spec.GetType().(type) {
	case *discoveryv1.MeshSpec_Istio_:
		meshType = "istio"
		meshInstallation = mesh.Spec.GetIstio().Installation
	case *discoveryv1.MeshSpec_Linkerd:
		meshType = "linkerd"
		meshInstallation = mesh.Spec.GetLinkerd().Installation
	case *discoveryv1.MeshSpec_ConsulConnect:
		meshType = "consulconnect"
		meshInstallation = mesh.Spec.GetConsulConnect().Installation
	case *discoveryv1.MeshSpec_Osm:
		meshType = "osm"
		meshInstallation = mesh.Spec.GetOsm().Installation
	}
	return meshMetadata{
		Type:      meshType,
		Namespace: meshInstallation.Namespace,
		Cluster:   meshInstallation.Cluster,
		Version:   meshInstallation.Version,
	}
}
