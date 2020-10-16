package mesh

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/internal/flags"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/printing"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Command(ctx context.Context, opts *flags.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mesh",
		Short:   "Description of managed meshes",
		Aliases: []string{"meshes"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.Kubeconfig, opts.Kubecontext)
			if err != nil {
				return err
			}
			description, err := describeMeshes(ctx, c)
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

func describeMeshes(ctx context.Context, c client.Client) (string, error) {
	meshClient := discoveryv1alpha2.NewMeshClient(c)
	meshList, err := meshClient.ListMesh(ctx)
	if err != nil {
		return "", err
	}
	var meshDescriptions []meshDescription
	for _, mesh := range meshList.Items {
		mesh := mesh // pike
		meshDescriptions = append(meshDescriptions, describeMesh(&mesh))
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Virtual_Mesh", "Failover_Services"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range meshDescriptions {
		table.Append([]string{
			description.Metadata.string(),
			printing.FormattedObjectRef(description.VirtualMesh),
			printing.FormattedObjectRefs(description.FailoverServices),
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
	Metadata         *meshMetadata
	VirtualMesh      *v1.ObjectRef
	FailoverServices []*v1.ObjectRef
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

func describeMesh(mesh *discoveryv1alpha2.Mesh) meshDescription {
	meshMeta := getMeshMetadata(mesh)

	var failoverServices []*v1.ObjectRef
	for _, fs := range mesh.Status.AppliedFailoverServices {
		failoverServices = append(failoverServices, fs.Ref)
	}

	return meshDescription{
		Metadata:         &meshMeta,
		VirtualMesh:      mesh.Status.GetAppliedVirtualMesh().GetRef(),
		FailoverServices: failoverServices,
	}
}

func getMeshMetadata(mesh *discoveryv1alpha2.Mesh) meshMetadata {
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
	var meshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation
	switch mesh.Spec.GetMeshType().(type) {
	case *discoveryv1alpha2.MeshSpec_Istio_:
		meshType = "istio"
		meshInstallation = mesh.Spec.GetIstio().Installation
	case *discoveryv1alpha2.MeshSpec_Linkerd:
		meshType = "linkerd"
		meshInstallation = mesh.Spec.GetLinkerd().Installation
	case *discoveryv1alpha2.MeshSpec_ConsulConnect:
		meshType = "consulconnect"
		meshInstallation = mesh.Spec.GetConsulConnect().Installation
	case *discoveryv1alpha2.MeshSpec_Osm:
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
