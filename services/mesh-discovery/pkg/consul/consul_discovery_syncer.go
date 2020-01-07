package consul

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	meshprojectsapi "github.com/solo-io/mesh-projects/pkg/api/v1"
	coreapi "github.com/solo-io/mesh-projects/pkg/api/v1/core"
	globalcommon "github.com/solo-io/mesh-projects/services/common"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	kubev1 "k8s.io/api/core/v1"
)

const (
	consulServerArg           = "-server"
	normalizedConsulImagePath = "library/consul"
	meshNamePrefix            = "consul"
)

var (
	DiscoveryLabels = map[string]string{
		"discovered_by": "consul-mesh-discovery",
	}

	datacenterRegex = regexp.MustCompile("-datacenter=([a-zA-Z0-9]*)")
)

type consulDiscoveryPlugin struct {
	consulConnectInstallationFinder ConsulConnectInstallationFinder
	imageNameParser                 globalcommon.ImageNameParser
}

func NewConsulDiscoveryPlugin(writeNamespace string,
	meshReconciler meshprojectsapi.MeshReconciler,
	reconciler meshprojectsapi.MeshIngressReconciler,
	consulConnectInstallationFinder ConsulConnectInstallationFinder,
	imageNameParser globalcommon.ImageNameParser) meshprojectsapi.DiscoverySyncer {

	return common.NewDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		reconciler,
		&consulDiscoveryPlugin{
			consulConnectInstallationFinder,
			imageNameParser,
		},
	)
}

func (c *consulDiscoveryPlugin) MeshType() string {
	return common.ConsulMeshID
}

func (c *consulDiscoveryPlugin) DiscoveryLabels() map[string]string {
	return DiscoveryLabels
}

func (c *consulDiscoveryPlugin) DesiredMeshes(ctx context.Context, writeNamespace string, snap *meshprojectsapi.DiscoverySnapshot) (consulMeshes meshprojectsapi.MeshList, err error) {
	logger := contextutils.LoggerFrom(ctx)
	namespaceToMesh := map[string]*meshprojectsapi.Mesh{}

	for _, deployment := range snap.Deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			// if we have already discovered an installation of consul, don't attempt to keep discovering in this namespace
			if _, ok := namespaceToMesh[deployment.Namespace]; ok {
				continue
			}

			isConsulInstallation, err := c.consulConnectInstallationFinder.IsConsulConnect(container)
			if err != nil {
				logger.Errorf("Error while detecting consul deployment: %v", err)
				continue
			}

			if isConsulInstallation {
				mesh, err := c.buildConsulMeshObject(deployment, container, writeNamespace)
				if err != nil {
					logger.Errorf("Error while building consul mesh object: %v", err)
					continue
				}

				namespaceToMesh[deployment.Namespace] = mesh
			}
		}
	}

	// flatten the values of the map into one array
	for _, mesh := range namespaceToMesh {
		consulMeshes = append(consulMeshes, mesh)
	}

	return
}

// returns an error if the image name is un-parsable
func (c *consulDiscoveryPlugin) buildConsulMeshObject(deployment *kubernetes.Deployment, container kubev1.Container, writeNamespace string) (*meshprojectsapi.Mesh, error) {
	parsedImage, err := c.imageNameParser.Parse(container.Image)
	if err != nil {
		return nil, err
	}

	imageVersion := parsedImage.Tag
	if parsedImage.Digest != "" {
		imageVersion = parsedImage.Digest
	}

	meshMetadata := core.Metadata{
		Name:      buildMeshName(deployment, container),
		Namespace: writeNamespace,
		Labels:    DiscoveryLabels,
	}
	meshRef := meshMetadata.Ref()
	ingressRef := meshprojectsapi.BuildMeshIngress(&meshRef, meshMetadata.Labels).Metadata.Ref()

	return &meshprojectsapi.Mesh{
		Metadata: meshMetadata,
		MeshType: &meshprojectsapi.Mesh_ConsulConnect{
			ConsulConnect: &meshprojectsapi.ConsulConnectMesh{
				Installation: &meshprojectsapi.MeshInstallation{
					InstallationNamespace: deployment.Namespace,
					Version:               imageVersion,
				},
			},
		},
		EntryPoint: &coreapi.ClusterResourceRef{
			Resource: ingressRef,
		},
	}, nil
}

// returns "consul(-$datacenterName)-$installNamespace(-$clusterName)"
func buildMeshName(deployment *kubernetes.Deployment, container kubev1.Container) string {
	meshName := meshNamePrefix

	cmd := strings.Join(container.Command, " ")
	datacenterNameMatch := datacenterRegex.FindStringSubmatch(cmd)

	if len(datacenterNameMatch) == 2 {
		meshName += fmt.Sprintf("-%s", datacenterNameMatch[1])
	}

	meshName += fmt.Sprintf("-%s", deployment.Namespace)

	if deployment.ClusterName != "" {
		meshName += fmt.Sprintf("-%s", deployment.ClusterName)
	}

	return meshName
}
