package linkerd

import (
	"context"
	"strings"

	"github.com/google/wire"
	linkerdconfig "github.com/linkerd/linkerd2/controller/gen/config"
	"github.com/linkerd/linkerd2/pkg/config"
	linkerdk8s "github.com/linkerd/linkerd2/pkg/k8s"
	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh"
	k8s_apps_v1 "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	WireProviderSet = wire.NewSet(
		NewLinkerdMeshScanner,
	)
	DiscoveryLabels = map[string]string{
		"discovered_by": "linkerd-mesh-discovery",
	}
	UnexpectedControllerImageName = func(err error, imageName string) error {
		return eris.Wrapf(err, "invalid or unexpected image name format for linkerd controller: %s", imageName)
	}
	LinkerdConfigMapName = linkerdk8s.ConfigConfigMapName
	DefaultClusterDomain = "cluster.local"
)

// disambiguates this MeshScanner from the other MeshScanner implementations so that wire stays happy
type LinkerdMeshScanner mesh.MeshScanner

func NewLinkerdMeshScanner(imageNameParser docker.ImageNameParser) LinkerdMeshScanner {
	return &linkerdMeshScanner{
		imageNameParser: imageNameParser,
	}
}

type linkerdMeshScanner struct {
	imageNameParser docker.ImageNameParser
}

func getLinkerdConfig(ctx context.Context, name, namespace string, kube client.Client) (*linkerdconfig.All, error) {
	cm := &k8s_core_types.ConfigMap{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := kube.Get(ctx, key, cm); err != nil {
		return nil, err
	}
	cfg, err := config.FromConfigMap(cm.Data)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (l *linkerdMeshScanner) ScanDeployment(ctx context.Context, clusterName string, deployment *k8s_apps_v1.Deployment, clusterScopedClient client.Client) (*zephyr_discovery.Mesh, error) {

	linkerdController, err := l.detectLinkerdController(clusterName, deployment)

	if err != nil {
		return nil, err
	}

	if linkerdController == nil {
		return nil, nil
	}

	linkerdConfig, err := getLinkerdConfig(ctx, LinkerdConfigMapName, linkerdController.namespace, clusterScopedClient)
	if err != nil {
		return nil, err
	}

	clusterDomain := linkerdConfig.GetGlobal().GetClusterDomain()
	if clusterDomain == "" {
		clusterDomain = DefaultClusterDomain
	}

	return &zephyr_discovery.Mesh{
		ObjectMeta: k8s_meta_v1.ObjectMeta{
			Name:      linkerdController.name(),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels,
		},
		Spec: zephyr_discovery_types.MeshSpec{
			MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{
				Linkerd: &zephyr_discovery_types.MeshSpec_LinkerdMesh{
					Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
						InstallationNamespace: deployment.GetNamespace(),
						Version:               linkerdController.version,
					},
					ClusterDomain: clusterDomain,
				},
			},
			Cluster: &zephyr_core_types.ResourceRef{
				Name:      clusterName,
				Namespace: env.GetWriteNamespace(),
			},
		},
	}, nil
}

func (l *linkerdMeshScanner) detectLinkerdController(clusterName string, deployment *k8s_apps_v1.Deployment) (*linkerdControllerDeployment, error) {
	var linkerdController *linkerdControllerDeployment

	for _, container := range deployment.Spec.Template.Spec.Containers {
		if strings.Contains(container.Image, "linkerd-io/controller") {
			// TODO there can be > 1 controller image per pod, do we care?
			parsedImage, err := l.imageNameParser.Parse(container.Image)
			if err != nil {
				return nil, UnexpectedControllerImageName(err, container.Image)
			}

			version := parsedImage.Tag
			if parsedImage.Digest != "" {
				version = parsedImage.Digest
			}
			linkerdController = &linkerdControllerDeployment{version: version, namespace: deployment.Namespace, cluster: clusterName}
		}
	}

	return linkerdController, nil
}

type linkerdControllerDeployment struct {
	version, namespace, cluster string
}

func (c linkerdControllerDeployment) name() string {
	if c.cluster == "" {
		return "linkerd-" + c.namespace
	}
	// TODO cluster is not restricted to kube name scheme, kebab it
	return "linkerd-" + c.namespace + "-" + c.cluster
}
