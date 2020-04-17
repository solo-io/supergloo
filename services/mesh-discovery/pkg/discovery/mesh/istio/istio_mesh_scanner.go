package istio

import (
	"context"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh"
	"istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/util/gogoprotomarshal"
	k8s_apps_types "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	IstiodDeploymentName  = "istiod"
	CitadelDeploymentName = "istio-citadel"
	IstioContainerKeyword = "istio"
	PilotContainerKeyword = "pilot"
	IstioConfigMapName    = "istio"
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

func NewIstioMeshScanner(
	imageNameParser docker.ImageNameParser,
	configMapClientFactory k8s_core.ConfigMapClientFactory,
) IstioMeshScanner {
	return &istioMeshScanner{
		imageNameParser:        imageNameParser,
		configMapClientFactory: configMapClientFactory,
	}
}

type istioMeshScanner struct {
	imageNameParser        docker.ImageNameParser
	configMapClientFactory k8s_core.ConfigMapClientFactory
}

func (i *istioMeshScanner) ScanDeployment(ctx context.Context, clusterName string, deployment *k8s_apps_types.Deployment, clusterScopedClient client.Client) (*zephyr_discovery.Mesh, error) {
	istioDeployment, err := i.detectIstioDeployment(clusterName, deployment)
	if err != nil {
		return nil, err
	} else if istioDeployment == nil {
		return nil, nil
	}
	trustDomain, err := i.getTrustDomain(ctx, clusterScopedClient, deployment.GetNamespace())
	if err != nil {
		return nil, err
	}
	return &zephyr_discovery.Mesh{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      istioDeployment.Name(),
			Namespace: env.GetWriteNamespace(),
			Labels:    DiscoveryLabels,
		},
		Spec: zephyr_discovery_types.MeshSpec{
			MeshType: &zephyr_discovery_types.MeshSpec_Istio{
				Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{
					Installation: &zephyr_discovery_types.MeshSpec_MeshInstallation{
						InstallationNamespace: deployment.GetNamespace(),
						Version:               istioDeployment.Version,
					},
					CitadelInfo: &zephyr_discovery_types.MeshSpec_IstioMesh_CitadelInfo{
						TrustDomain:      trustDomain,
						CitadelNamespace: deployment.GetNamespace(),
						// This assumes that the istiod deployment is the cert provider
						CitadelServiceAccount: deployment.Spec.Template.Spec.ServiceAccountName,
					},
				},
			},
			Cluster: &zephyr_core_types.ResourceRef{
				Name:      clusterName,
				Namespace: env.GetWriteNamespace(),
			},
		},
	}, nil
}

// TODO: Delete once fully migrated to Istio 1.5
// This is to support Istio 1.4 where a separate istio-citadel deployment provides certificates.
func isCitadelDeployment(deploymentName string) bool {
	return deploymentName == CitadelDeploymentName
}

func (i *istioMeshScanner) detectIstioDeployment(clusterName string, deployment *k8s_apps_types.Deployment) (*istioDeployment, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		// Detect either istiod deployment (for Istio versions >= 1.5) or istio-citadel deployment (for Istio versions < 1.5)
		if isIstiod(deployment, &container) || isCitadelDeployment(deployment.GetName()) {
			parsedImage, err := i.imageNameParser.Parse(container.Image)
			if err != nil {
				return nil, UnexpectedPilotImageName(err, container.Image)
			}
			version := parsedImage.Tag
			if parsedImage.Digest != "" {
				version = parsedImage.Digest
			}
			return &istioDeployment{Version: version, Namespace: deployment.Namespace, Cluster: clusterName}, nil
		}
	}

	return nil, nil
}

func (i *istioMeshScanner) getTrustDomain(
	ctx context.Context,
	clusterScopedClient client.Client,
	namespace string,
) (string, error) {
	configMapClient := i.configMapClientFactory(clusterScopedClient)
	configMap, err := configMapClient.GetConfigMap(ctx, client.ObjectKey{Name: IstioConfigMapName, Namespace: namespace})
	if err != nil {
		return "", err
	}
	meshConfigString, ok := configMap.Data["mesh"]
	if !ok {
		return "", eris.Errorf("Failed to find 'mesh' entry in ConfigMap with name/namespace %s/%s", IstioConfigMapName, namespace)
	}
	var meshConfig v1alpha1.MeshConfig
	err = gogoprotomarshal.ApplyYAML(meshConfigString, &meshConfig)
	if err != nil {
		return "", eris.Wrapf(err, "Failed to parse yaml from ConfigMap with name/namespace %s/%s", IstioConfigMapName, namespace)
	}
	return meshConfig.TrustDomain, nil
}

// Return true if deployment is inferred to be an Istiod deployment
func isIstiod(deployment *k8s_apps_types.Deployment, container *k8s_core_types.Container) bool {
	return deployment.GetName() == IstiodDeploymentName &&
		strings.Contains(container.Image, IstioContainerKeyword) &&
		strings.Contains(container.Image, PilotContainerKeyword)
}

type istioDeployment struct {
	Version, Namespace, Cluster string
}

// TODO merge with linkerd controller type
func (i istioDeployment) Name() string {
	if i.Cluster == "" {
		return "istio-" + i.Namespace
	}
	// TODO: https://github.com/solo-io/service-mesh-hub/issues/141
	return "istio-" + i.Namespace + "-" + i.Cluster
}
