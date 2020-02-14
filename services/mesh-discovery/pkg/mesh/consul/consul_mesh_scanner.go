package consul

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	k8s_core_v1 "k8s.io/api/core/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	consulServerArg           = "-server"
	normalizedConsulImagePath = "library/consul"
	meshNamePrefix            = "consul"
)

var (
	WireProviderSet = wire.NewSet(
		NewConsulConnectInstallationScanner,
		NewConsulMeshScanner,
	)
	DiscoveryLabels = map[string]string{
		"discovered_by": "consul-mesh-discovery",
	}
	ErrorDetectingDeployment = func(err error) error {
		return eris.Wrap(err, "Error while detecting consul deployment")
	}

	datacenterRegex = regexp.MustCompile("-datacenter=([a-zA-Z0-9]*)")
)

// disambiguates this MeshScanner from the other MeshScanner implementations so that wire stays happy
type ConsulConnectMeshScanner mesh.MeshScanner

func NewConsulMeshScanner(
	imageNameParser docker.ImageNameParser,
	consulConnectInstallationScanner ConsulConnectInstallationScanner) ConsulConnectMeshScanner {
	return &consulMeshScanner{
		consulConnectInstallationScanner,
		imageNameParser,
	}
}

type consulMeshScanner struct {
	consulConnectInstallationScanner ConsulConnectInstallationScanner
	imageNameParser                  docker.ImageNameParser
}

func (c *consulMeshScanner) ScanDeployment(_ context.Context,
	deployment *k8s_apps_v1.Deployment) (*discoveryv1alpha1.Mesh, error) {

	for _, container := range deployment.Spec.Template.Spec.Containers {
		isConsulInstallation, err := c.consulConnectInstallationScanner.IsConsulConnect(container)
		if err != nil {
			return nil, ErrorDetectingDeployment(err)
		}

		if !isConsulInstallation {
			continue
		}

		return c.buildConsulMeshObject(deployment, container, env.DefaultWriteNamespace)
	}

	return nil, nil
}

// returns an error if the image name is un-parsable
func (c *consulMeshScanner) buildConsulMeshObject(
	deployment *k8s_apps_v1.Deployment,
	container k8s_core_v1.Container,
	writeNamespace string) (*discoveryv1alpha1.Mesh, error) {

	parsedImage, err := c.imageNameParser.Parse(container.Image)
	if err != nil {
		return nil, err
	}

	imageVersion := parsedImage.Tag
	if parsedImage.Digest != "" {
		imageVersion = parsedImage.Digest
	}

	return &discoveryv1alpha1.Mesh{
		ObjectMeta: k8s_meta_v1.ObjectMeta{
			Name:      buildMeshName(deployment, container),
			Namespace: env.DefaultWriteNamespace,
			Labels:    DiscoveryLabels,
		},
		Spec: discovery_types.MeshSpec{
			MeshType: &discovery_types.MeshSpec_ConsulConnect{
				ConsulConnect: &discovery_types.ConsulConnectMesh{
					Installation: &discovery_types.MeshInstallation{
						InstallationNamespace: deployment.GetNamespace(),
						Version:               imageVersion,
					},
				},
			},
			Cluster: &core_types.ResourceRef{
				Name:      deployment.GetClusterName(),
				Namespace: env.DefaultWriteNamespace,
			},
		},
	}, nil
}

// returns "consul(-$datacenterName)-$installNamespace(-$clusterName)"
func buildMeshName(deployment *k8s_apps_v1.Deployment, container k8s_core_v1.Container) string {
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
