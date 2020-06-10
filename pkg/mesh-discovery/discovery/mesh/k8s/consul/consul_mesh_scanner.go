package consul

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	k8s_core_v1 "k8s.io/api/core/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
type ConsulConnectMeshScanner k8s.MeshScanner

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

func (c *consulMeshScanner) ScanDeployment(_ context.Context, clusterName string, deployment *k8s_apps_v1.Deployment, _ client.Client) (*smh_discovery.Mesh, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		isConsulInstallation, err := c.consulConnectInstallationScanner.IsConsulConnect(container)
		if err != nil {
			return nil, ErrorDetectingDeployment(err)
		}

		if !isConsulInstallation {
			continue
		}

		return c.buildConsulMeshObject(deployment, container, clusterName)
	}

	return nil, nil
}

// returns an error if the image name is un-parsable
func (c *consulMeshScanner) buildConsulMeshObject(
	deployment *k8s_apps_v1.Deployment,
	container k8s_core_v1.Container,
	clusterName string) (*smh_discovery.Mesh, error) {

	parsedImage, err := c.imageNameParser.Parse(container.Image)
	if err != nil {
		return nil, err
	}

	imageVersion := parsedImage.Tag
	if parsedImage.Digest != "" {
		imageVersion = parsedImage.Digest
	}

	return &smh_discovery.Mesh{
		ObjectMeta: k8s_meta_v1.ObjectMeta{
			Name:      buildMeshName(clusterName, deployment, container),
			Namespace: container_runtime.GetWriteNamespace(),
			Labels:    DiscoveryLabels,
		},
		Spec: smh_discovery_types.MeshSpec{
			MeshType: &smh_discovery_types.MeshSpec_ConsulConnect{
				ConsulConnect: &smh_discovery_types.MeshSpec_ConsulConnectMesh{
					Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
						InstallationNamespace: deployment.GetNamespace(),
						Version:               imageVersion,
					},
				},
			},
			Cluster: &smh_core_types.ResourceRef{
				Name:      clusterName,
				Namespace: container_runtime.GetWriteNamespace(),
			},
		},
	}, nil
}

// returns "consul(-$datacenterName)-$installNamespace(-$clusterName)"
func buildMeshName(clusterName string, deployment *k8s_apps_v1.Deployment, container k8s_core_v1.Container) string {
	meshName := meshNamePrefix

	cmd := strings.Join(container.Command, " ")
	datacenterNameMatch := datacenterRegex.FindStringSubmatch(cmd)

	if len(datacenterNameMatch) == 2 {
		meshName += fmt.Sprintf("-%s", datacenterNameMatch[1])
	}

	meshName += fmt.Sprintf("-%s", deployment.Namespace)

	if deployment.ClusterName != "" {
		meshName += fmt.Sprintf("-%s", clusterName)
	}

	return meshName
}
