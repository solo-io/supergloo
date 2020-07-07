package linkerd

import (
	"strings"

	linkerdconfig "github.com/linkerd/linkerd2/controller/gen/config"
	"github.com/linkerd/linkerd2/pkg/config"
	linkerdk8s "github.com/linkerd/linkerd2/pkg/k8s"
	"github.com/rotisserie/eris"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/common/docker"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/smh/pkg/mesh-discovery/translation/utils"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	linkerdConfigMapName = linkerdk8s.ConfigConfigMapName
	linkerdImageName     = "linkerd-io/controller"
)

// detects Linkerd if a deployment contains the linkerd container.
type meshDetector struct {
	configMaps corev1sets.ConfigMapSet
}

func NewMeshDetector(configMaps corev1sets.ConfigMapSet) detector.MeshDetector {
	return &meshDetector{configMaps: configMaps}
}

// returns nil, nil of the deployment does not contain the linkerd image
func (d *meshDetector) DetectMesh(deployment *appsv1.Deployment) (*v1alpha1.Mesh, error) {
	version, err := d.getLinkerdVersion(deployment)
	if err != nil {
		return nil, err
	}

	if version == "" {
		return nil, nil
	}

	linkerdConfig, err := getLinkerdConfig(d.configMaps, deployment.ClusterName, deployment.Namespace)
	if err != nil {
		return nil, err
	}

	clusterDomain := linkerdConfig.GetGlobal().GetClusterDomain()

	mesh := &v1alpha1.Mesh{
		ObjectMeta: utils.DiscoveredObjectMeta(deployment),
		Spec: v1alpha1.MeshSpec{
			MeshType: &v1alpha1.MeshSpec_Linkerd{
				Linkerd: &v1alpha1.MeshSpec_LinkerdMesh{
					Installation: &v1alpha1.MeshSpec_MeshInstallation{
						Namespace: deployment.Namespace,
						Cluster:   deployment.ClusterName,
						Version:   version,
					},
					ClusterDomain: clusterDomain,
				},
			},
		},
	}

	return mesh, nil
}

func (d *meshDetector) getLinkerdVersion(deployment *appsv1.Deployment) (string, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if strings.Contains(container.Image, linkerdImageName) {
			// TODO there can be > 1 controller image per pod, do we care?
			parsedImage, err := docker.ParseImageName(container.Image)
			if err != nil {
				return "", eris.Wrapf(err, "failed to parse linkerd image tag: %s", container.Image)
			}

			version := parsedImage.Tag
			if parsedImage.Digest != "" {
				version = parsedImage.Digest
			}
			return version, nil
		}
	}

	return "", nil
}

func getLinkerdConfig(
	configMaps corev1sets.ConfigMapSet,
	cluster,
	namespace string,
) (*linkerdconfig.All, error) {
	linkerdConfigMap, err := configMaps.Find(&skv1.ClusterObjectRef{
		Name:        linkerdConfigMapName,
		Namespace:   namespace,
		ClusterName: cluster,
	})
	if err != nil {
		return nil, err
	}

	cfg, err := config.FromConfigMap(linkerdConfigMap.Data)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
