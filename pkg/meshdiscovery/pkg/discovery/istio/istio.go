package istio

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	istio      = "istio"
	istioPilot = istio + "-pilot"
)

type istioMeshDiscovery struct {
	ctx            context.Context
	pods           v1.PodList
	existingMeshes v1.MeshList
}

func NewIstioMeshDiscovery(ctx context.Context, pods v1.PodList, meshes v1.MeshList) *istioMeshDiscovery {
	existingMeshes := filterIstioMeshes(meshes)
	return &istioMeshDiscovery{ctx: ctx, pods: pods, existingMeshes: existingMeshes}
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

func (imd *istioMeshDiscovery) DiscoverMeshes() (v1.MeshList, error) {
	logger := contextutils.LoggerFrom(imd.ctx)
	possibleIstioPods := filterIstioPods(imd.pods)
	if len(possibleIstioPods) == 0 {
		logger.Debugf("no possible istio pods found")
		return nil, nil
	}

	pilotPods := findIstioPilots(possibleIstioPods)
	if len(pilotPods) == 0 {
		logger.Debugf("no pilot pods found in istio pod list")
		return nil, nil
	}

	return v1.MeshList{}, nil
}

func filterIstioPods(pods v1.PodList) v1.PodList {
	var result v1.PodList
	for _, pod := range pods {
		if strings.Contains(pod.Name, istio) {
			result = append(result, pod)
		}
	}
	return result
}

func findIstioPilots(pods v1.PodList) v1.PodList {
	var result v1.PodList
	for _, pod := range pods {
		if strings.Contains(pod.Name, istioPilot) {
			result = append(result, pod)
		}
	}
	return result
}

//
// func getVersionFromPod(pod v1.Pod) (string, error) {
// 	containers := pod.Spec.Containers
// 	for _, container := range containers {
//
// 	}
// 	return "", nil
// }
//
