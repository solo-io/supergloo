package appmesh

import (
	"strconv"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	customkube "github.com/solo-io/supergloo/pkg/api/external/kubernetes/core/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
	kubev1 "k8s.io/api/core/v1"
)

func getPodInfo(mesh *v1.Mesh, kubePod *kubev1.Pod) (*podInfo, error) {
	var info *podInfo
	for _, container := range kubePod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == PodVirtualNodeEnvName {

				// Env value is expected to have the following format
				// - name: "APPMESH_VIRTUAL_NODE_NAME"
				//   value: "mesh/meshName/virtualNode/virtualNodeName"
				vnNameParts := strings.Split(env.Value, "/")
				if len(vnNameParts) != 4 {
					return nil, errors.Errorf("unexpected format for %s env for pod %s.%s. Expected format is [%s] but found [%s]",
						PodVirtualNodeEnvName, kubePod.Namespace, kubePod.Name, "mesh/meshName/virtualNode/virtualNodeName", env.Value)
				}

				if meshName := vnNameParts[1]; meshName == mesh.Metadata.Name {
					info = &podInfo{
						virtualNodeName: vnNameParts[3],
					}
				}
			}
		}
	}

	if info != nil {
		for _, initContainer := range kubePod.Spec.InitContainers {
			for _, env := range initContainer.Env {
				if env.Name == PodPortsEnvName {
					for _, portStr := range strings.Split(env.Value, ",") {
						ui64, err := strconv.ParseUint(strings.Trim(portStr, " "), 10, 32)
						if err != nil {
							return nil, errors.Wrapf(err, "failed to parse [%s] (value of %s env) to int array",
								env.Value, PodPortsEnvName)
						}
						info.ports = append(info.ports, uint32(ui64))
					}
				}
			}
		}
	}

	return info, nil
}

func getUpstreamsForMesh(upstreams gloov1.UpstreamList, podInfo AwsAppMeshPodInfo, appMeshPodList customkube.PodList) (AwsAppMeshUpstreamInfo, gloov1.UpstreamList, error) {
	appMeshUpstreamInfo := make(AwsAppMeshUpstreamInfo, 0)
	for _, us := range upstreams {

		// Get all the appMesh pods for this upstream
		pods, err := utils.PodsForUpstreams(gloov1.UpstreamList{us}, appMeshPodList)
		if err != nil {
			return nil, nil, err
		}
		if len(pods) > 0 {
			appMeshUpstreamInfo[us] = pods
		}

		// Add this upstream to the info the pod it belongs to
		for _, pod := range pods {
			podInfo[pod].upstreams = append(podInfo[pod].upstreams, us)
		}
	}

	var appMeshUpstreamList gloov1.UpstreamList
	for us := range appMeshUpstreamInfo {
		appMeshUpstreamList = append(appMeshUpstreamList, us)
	}

	return appMeshUpstreamInfo, appMeshUpstreamList, nil

}

func getPodsForMesh(mesh *v1.Mesh, pods customkube.PodList) (AwsAppMeshPodInfo, customkube.PodList, error) {
	var appMeshPodList customkube.PodList
	appMeshPods := make(AwsAppMeshPodInfo, 0)
	for _, pod := range pods {
		kubePod, err := kubernetes.ToKube(pod)
		if err != nil {
			return nil, nil, err
		}

		info, err := getPodInfo(mesh, kubePod)
		if err != nil {
			return nil, nil, err
		}

		if info != nil {
			appMeshPodList = append(appMeshPodList, pod)
			appMeshPods[pod] = info
		}
	}
	return appMeshPods, appMeshPodList, nil
}
