package istio

import (
	"context"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/common/docker"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	WireProviderSet = wire.NewSet(
		NewIstioMeshScanner,
	)
	DiscoveryLabels = map[string]string{
		"discovered_by": "istio-mesh-discovery",
	}
	UnexpectedPilotImageName = func(err error, imageName string) error {
		return eris.Wrapf(err, "invalid or unexpected image name format for istio pilot: %s", imageName)
	}
)

// disambiguates this MeshScanner from the other MeshScanner implementations so that wire stays happy
type IstioMeshScanner mesh.MeshScanner

func NewIstioMeshScanner(imageNameParser docker.ImageNameParser) IstioMeshScanner {
	return &istioMeshScanner{
		imageNameParser: imageNameParser,
	}
}

type istioMeshScanner struct {
	imageNameParser docker.ImageNameParser
}

func (i *istioMeshScanner) ScanDeployment(_ context.Context, deployment *k8s_apps_v1.Deployment) (*discoveryv1alpha1.Mesh, error) {

	pilot, err := i.detectPilotDeployment(deployment)
	if err != nil {
		return nil, err
	} else if pilot == nil {
		return nil, nil
	}

	return &discoveryv1alpha1.Mesh{
		ObjectMeta: k8s_meta_v1.ObjectMeta{
			Name:      pilot.Name(),
			Namespace: env.DefaultWriteNamespace,
			Labels:    DiscoveryLabels,
		},
		Spec: discovery_types.MeshSpec{
			MeshType: &discovery_types.MeshSpec_Istio{
				Istio: &discovery_types.IstioMesh{
					Installation: &discovery_types.MeshInstallation{
						InstallationNamespace: deployment.GetNamespace(),
						Version:               pilot.Version,
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

func (i *istioMeshScanner) detectPilotDeployment(deployment *k8s_apps_v1.Deployment) (*PilotDeployment, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if strings.Contains(container.Image, "istio") && strings.Contains(container.Image, "pilot") {
			parsedImage, err := i.imageNameParser.Parse(container.Image)
			if err != nil {
				return nil, UnexpectedPilotImageName(err, container.Image)
			}

			version := parsedImage.Tag
			if parsedImage.Digest != "" {
				version = parsedImage.Digest
			}

			return &PilotDeployment{Version: version, Namespace: deployment.Namespace, Cluster: deployment.ClusterName}, nil
		}
	}

	return nil, nil
}

type PilotDeployment struct {
	Version, Namespace, Cluster string
}

// TODO merge with linkerd controller type
func (c PilotDeployment) Name() string {
	if c.Cluster == "" {
		return "istio-" + c.Namespace
	}
	// TODO: https://github.com/solo-io/mesh-projects/issues/141
	return "istio-" + c.Namespace + "-" + c.Cluster
}
