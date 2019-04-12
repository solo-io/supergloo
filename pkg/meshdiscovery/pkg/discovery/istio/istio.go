package istio

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	istio      = "istio"
	pilot      = "pilot"
	istioPilot = istio + "-" + pilot

	injectionLabel = "istio-injection"
)

type istioMeshDiscovery struct {
}

func NewIstioMeshDiscovery() *istioMeshDiscovery {
	return &istioMeshDiscovery{}
}

func filterIstioMeshes(meshes v1.MeshList) v1.MeshList {
	var result v1.MeshList
	for _, mesh := range meshes {
		if istioMesh := mesh.GetIstio(); istioMesh != nil {
			result = append(result, mesh)
		}
	}
	return result
}

func (imd *istioMeshDiscovery) DiscoverMeshes(ctx context.Context, snapshot *v1.DiscoverySnapshot) (v1.MeshList, error) {
	pods := snapshot.Pods.List()
	existingMeshes := filterIstioMeshes(snapshot.Meshes.List())
	logger := contextutils.LoggerFrom(ctx)

	pilotPods := findIstioPods(pods)
	if len(pilotPods) == 0 {
		logger.Debugf("no pilot pods found in istio pod list")
		return nil, nil
	}

	var discoveredMeshes v1.MeshList
	for _, pilotPod := range pilotPods {
		if strings.Contains(pilotPod.Name, istioPilot) {
			mesh, err := constructDiscoveryData(ctx, pilotPod)
			if err != nil {
				return nil, err
			}
			logger.Debugf("successfully discovered mesh data for %v", mesh)
			discoveredMeshes = append(discoveredMeshes, mesh)
		}
	}

	mergedMeshes, err := mergeMeshes(discoveredMeshes, existingMeshes)
	if err != nil {
		return nil, err
	}

	return mergedMeshes, nil
}

func getWriteNamespace() string {
	if writeNamespace := os.Getenv("POD_NAMESPACE"); writeNamespace != "" {
		return writeNamespace
	}
	return "supergloo-system"
}

func mergeMeshes(discoveredMeshes, existingMeshes v1.MeshList) (v1.MeshList, error) {
	var mergedMeshes v1.MeshList
	for _, discoveredMesh := range discoveredMeshes {
		meshExists := false
		for _, existingMesh := range existingMeshes {
			istioMesh := existingMesh.GetIstio()
			if istioMesh == nil {
				continue
			}

			// This discovered mesh already exists, update the discovery data
			if istioMesh.InstallationNamespace == discoveredMesh.DiscoveryMetadata.InstallationNamespace {
				existingMesh.DiscoveryMetadata = discoveredMesh.DiscoveryMetadata
				meshExists = true
				break
			}
		}
		if meshExists {
			continue
		}

		discoveredMesh.MeshType = &v1.Mesh_Istio{
			Istio: &v1.IstioMesh{
				InstallationNamespace: discoveredMesh.DiscoveryMetadata.InstallationNamespace,
				IstioVersion:          discoveredMesh.DiscoveryMetadata.MeshVersion,
			},
		}
		discoveredMesh.MtlsConfig = discoveredMesh.DiscoveryMetadata.MtlsConfig

		mergedMeshes = append(mergedMeshes, discoveredMesh)
	}
	mergedMeshes = append(mergedMeshes, existingMeshes...)
	return mergedMeshes, nil
}

func constructDiscoveryData(ctx context.Context, istioPilotPod *v1.Pod) (*v1.Mesh, error) {
	logger := contextutils.LoggerFrom(ctx)
	mesh := &v1.Mesh{
		Metadata: core.Metadata{
			Namespace: getWriteNamespace(),
			Name:      fmt.Sprintf("istio-%s", istioPilotPod.Namespace),
		},
	}

	istioVersion, err := getVersionFromPod(istioPilotPod)
	if err != nil {
		logger.Debugf("unable to find version from pod %v", istioPilotPod)
		return nil, err
	}

	discoveryData := &v1.DiscoveryMetadata{
		InstallationNamespace:  istioPilotPod.Namespace,
		MeshVersion:            istioVersion,
		InjectedNamespaceLabel: injectionLabel,
	}
	mesh.DiscoveryMetadata = discoveryData
	return mesh, nil
}

func findIstioPods(pods v1.PodList) v1.PodList {
	var result v1.PodList
	for _, pod := range pods {
		if strings.Contains(pod.Name, istio) {
			result = append(result, pod)
		}
	}
	return result
}

func imageVersion(image string) (string, error) {
	regex := regexp.MustCompile("([0-9]+[.][0-9]+[.][0-9]+$)")
	imageTag := regex.FindString(image)
	if imageTag == "" {
		return "", errors.Errorf("unable to find image version for image: %s", image)
	}
	return imageTag, nil
}

func getVersionFromPod(pod *v1.Pod) (string, error) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		if strings.Contains(container.Image, istio) && strings.Contains(container.Image, pilot) {
			return imageVersion(container.Image)
		}
	}
	return "", errors.Errorf("unable to find pilot container from pod")
}
