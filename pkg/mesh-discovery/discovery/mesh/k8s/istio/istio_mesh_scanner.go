package istio

import (
	"context"
	"strings"

	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	"github.com/solo-io/service-mesh-hub/pkg/common/metadata"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/container-runtime/docker"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh/k8s"
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
type IstioMeshScanner k8s.MeshScanner

func NewIstioMeshScanner(
	imageNameParser docker.ImageNameParser,
	configMapClientFactory k8s_core_providers.ConfigMapClientFactory,
) IstioMeshScanner {
	return &istioMeshScanner{
		imageNameParser:        imageNameParser,
		configMapClientFactory: configMapClientFactory,
	}
}

type istioMeshScanner struct {
	imageNameParser        docker.ImageNameParser
	configMapClientFactory k8s_core_providers.ConfigMapClientFactory
}

func (i *istioMeshScanner) ScanDeployment(ctx context.Context, clusterName string, deployment *k8s_apps_types.Deployment, configMapClient k8s_core.ConfigMapClient) (*smh_discovery.Mesh, error) {
	istioDeployment, err := i.detectIstioDeployment(clusterName, deployment)
	if err != nil {
		return nil, err
	} else if istioDeployment == nil {
		return nil, nil
	}
	trustDomain, err := i.getTrustDomain(ctx, configMapClient, deployment.GetNamespace())
	if err != nil {
		return nil, err
	}

	meshSpec, meshType, err := i.buildMeshSpec(
		istioDeployment,
		clusterName,
		trustDomain,
		deployment.GetNamespace(),
		deployment.Spec.Template.Spec.ServiceAccountName,
	)
	if err != nil {
		return nil, err
	}

	return &smh_discovery.Mesh{
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      metadata.BuildMeshName(meshType, istioDeployment.Namespace, istioDeployment.Cluster),
			Namespace: container_runtime.GetWriteNamespace(),
			Labels:    DiscoveryLabels,
		},
		Spec: *meshSpec,
	}, nil
}

func (*istioMeshScanner) buildMeshSpec(
	deployment *istioDeployment,
	clusterName string,
	trustDomain string,
	citadelNamespace string,
	citadelServiceAccountName string,
) (*smh_discovery_types.MeshSpec, smh_core_types.MeshType, error) {
	cluster := &smh_core_types.ResourceRef{
		Name:      clusterName,
		Namespace: container_runtime.GetWriteNamespace(),
	}
	istioMetadata := &smh_discovery_types.MeshSpec_IstioMesh{
		Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
			InstallationNamespace: deployment.Namespace,
			Version:               deployment.Version,
		},
		CitadelInfo: &smh_discovery_types.MeshSpec_IstioMesh_CitadelInfo{
			TrustDomain:      trustDomain,
			CitadelNamespace: citadelNamespace,
			// This assumes that the istiod deployment is the cert provider
			CitadelServiceAccount: citadelServiceAccountName,
		},
	}

	if strings.HasPrefix(deployment.Version, "1.5") {
		return &smh_discovery_types.MeshSpec{
			Cluster: cluster,
			MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
				Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
					Metadata: istioMetadata,
				},
			},
		}, smh_core_types.MeshType_ISTIO1_5, nil
	} else if strings.HasPrefix(deployment.Version, "1.6") {
		return &smh_discovery_types.MeshSpec{
			Cluster: cluster,
			MeshType: &smh_discovery_types.MeshSpec_Istio1_6_{
				Istio1_6: &smh_discovery_types.MeshSpec_Istio1_6{
					Metadata: istioMetadata,
				},
			},
		}, smh_core_types.MeshType_ISTIO1_6, nil
	} else {
		return nil, smh_core_types.MeshType_ISTIO1_5, eris.Errorf("Unrecognized Istio version: %s", deployment.Version)
	}
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
	configMapClient k8s_core.ConfigMapClient,
	namespace string,
) (string, error) {
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
