package istio

import (
	"context"
	"regexp"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	istio      = "istio"
	pilot      = "pilot"
	istioPilot = istio + "-" + pilot

	injectionConst = "injection-enabled"
)

type istioMeshDiscovery struct {
}
type IstioMeshDiscoveryContext struct {
	pods           v1.PodList
	existingMeshes v1.MeshList
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

	discoveryCtx := IstioMeshDiscoveryContext{
		pods:           snapshot.Pods.List(),
		existingMeshes: filterIstioMeshes(snapshot.Meshes.List()),
	}
	logger := contextutils.LoggerFrom(ctx)

	pilotPods := findIstioPods(discoveryCtx.pods)
	if len(pilotPods) == 0 {
		logger.Debugf("no pilot pods found in istio pod list")
		return nil, nil
	}

	var meshes v1.MeshList
	for _, pilotPod := range pilotPods {
		if strings.Contains(pilotPod.Name, istioPilot) {
			mesh, err := constructDiscoveryData(pilotPod)
			if err != nil {
				return nil, err
			}
			meshes = append(meshes, mesh)
		}
	}

	return meshes, nil
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

func constructDiscoveryData(istioPilotPod *v1.Pod) (*v1.Mesh, error) {
	mesh := &v1.Mesh{}

	istioVersion, err := getVersionFromPod(istioPilotPod)
	if err != nil {
		return nil, err
	}

	discoveryData := &v1.DiscoveryMetadata{
		InstallationNamespace: istioPilotPod.Namespace,
		MeshVersion:           istioVersion,
	}
	mesh.DiscoveryMetadata = discoveryData
	return mesh, nil
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
