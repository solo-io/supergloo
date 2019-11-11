package appmesh

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/solo-io/mesh-projects/services/internal/utils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

const (
	PodVirtualNodeEnvName = "APPMESH_VIRTUAL_NODE_NAME"
	PodPortsEnvName       = "APPMESH_APP_PORTS"
)

type podInfo struct {
	// These come from the APPMESH_APP_PORTS envs on pods that have been injected
	ports []uint32
	// These come from the APPMESH_VIRTUAL_NODE_NAME envs on pods that have been injected
	virtualNodeName string
	// All the upstreams that match this pod
	upstreams gloov1.UpstreamList
}

type awsAppMeshPodInfo map[*kubernetes.Pod]*podInfo
type awsAppMeshUpstreamInfo map[*gloov1.Upstream][]*kubernetes.Pod

type AwsAppMeshConfiguration interface {
	// Return all injected Upstreams
	InjectedUpstreams() gloov1.UpstreamList
}

// Represents the output of the App Mesh translator
type AwsAppMeshConfigurationImpl struct {
	// We build these objects once in the constructor. They are meant to help in all the translation operations where we
	// probably will need to look up pods by upstreams and vice-versa multiple times.
	PodList      kubernetes.PodList
	upstreamInfo awsAppMeshUpstreamInfo
	UpstreamList gloov1.UpstreamList

	// These are the actual results of the translations
	MeshName string
}

func NewAwsAppMeshConfiguration(meshName string, pods kubernetes.PodList, upstreams gloov1.UpstreamList) (AwsAppMeshConfiguration, error) {

	// Get all pods that point to this mesh via the APPMESH_VIRTUAL_NODE_NAME env set on their AWS App Mesh sidecar.
	appMeshPodInfo, appMeshPodList, err := getPodsForMesh(meshName, pods)
	if err != nil {
		return nil, err
	}

	// Find all upstreams that are associated with the appmesh pods
	// Also updates each podInfo in appMeshPodInfo with the list of upstreams that match it
	appMeshUpstreamInfo, appMeshUpstreamList := getUpstreamsForMesh(upstreams, appMeshPodInfo, appMeshPodList)

	return &AwsAppMeshConfigurationImpl{
		PodList:      appMeshPodList,
		upstreamInfo: appMeshUpstreamInfo,
		UpstreamList: appMeshUpstreamList,

		MeshName: meshName,
	}, nil
}

func (c *AwsAppMeshConfigurationImpl) InjectedUpstreams() gloov1.UpstreamList {
	return c.UpstreamList
}

func getPodsForMesh(meshName string, pods kubernetes.PodList) (awsAppMeshPodInfo, kubernetes.PodList, error) {
	var appMeshPodList kubernetes.PodList
	appMeshPods := make(awsAppMeshPodInfo, 0)
	for _, pod := range pods {
		info, err := getPodInfo(meshName, pod)
		if err != nil {
			return nil, nil, err
		}

		if info != nil {
			appMeshPodList = append(appMeshPodList, pod)
			appMeshPods[pod] = info
		}
	}

	sort.SliceStable(appMeshPodList, func(i, j int) bool {
		iPod := appMeshPodList[i]
		jPod := appMeshPodList[j]
		return fmt.Sprintf("%s.%s", iPod.Namespace, iPod.Name) < fmt.Sprintf("%s.%s", jPod.Namespace, jPod.Name)
	})

	return appMeshPods, appMeshPodList, nil
}

func getPodInfo(meshName string, pod *kubernetes.Pod) (*podInfo, error) {
	var info *podInfo
	for _, container := range pod.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == PodVirtualNodeEnvName {

				// Env value is expected to have the following format
				// - name: "APPMESH_VIRTUAL_NODE_NAME"
				//   value: "mesh/meshName/virtualNode/virtualNodeName"
				vnNameParts := strings.Split(env.Value, "/")
				if len(vnNameParts) != 4 {
					return nil, errors.Errorf("unexpected format for %s env for pod %s.%s. Expected format is [%s] but found [%s]",
						PodVirtualNodeEnvName, pod.Namespace, pod.Name, "mesh/meshName/virtualNode/virtualNodeName", env.Value)
				}

				if podMeshName := vnNameParts[1]; podMeshName == meshName {
					info = &podInfo{
						virtualNodeName: vnNameParts[3],
					}
				}
			}
		}
	}

	// info is non-nil is this pod has the appmesh sidecar
	if info != nil {
		for _, initContainer := range pod.Spec.InitContainers {
			for _, env := range initContainer.Env {
				if env.Name == PodPortsEnvName {
					for _, portStr := range strings.Split(env.Value, ",") {
						ui64, err := strconv.ParseUint(strings.TrimSpace(portStr), 10, 32)
						if err != nil {
							return nil, errors.Wrapf(err, "failed to parse [%s] (value of %s env) to int array",
								env.Value, PodPortsEnvName)
						}
						info.ports = append(info.ports, uint32(ui64))
					}
				}
			}
		}
		if len(info.ports) == 0 {
			return nil, errors.Errorf("could not find %s env on any initContainer for pod %s", PodPortsEnvName, pod.Name)
		}
	}

	return info, nil
}

func namespaceId(meta core.Metadata) string {
	return fmt.Sprintf("%s.%s", meta.Namespace, meta.Name)
}

func getUpstreamsForMesh(upstreams gloov1.UpstreamList, podInfo awsAppMeshPodInfo, appMeshPodList kubernetes.PodList) (awsAppMeshUpstreamInfo, gloov1.UpstreamList) {
	appMeshUpstreamInfo := make(awsAppMeshUpstreamInfo, 0)
	for _, us := range upstreams {

		// Get all the appMesh pods for this upstream
		pods := utils.PodsForUpstreams(gloov1.UpstreamList{us}, appMeshPodList)
		if len(pods) > 0 {
			appMeshUpstreamInfo[us] = pods
		}

		// Add this upstream to the info of the pod it belongs to
		for _, pod := range pods {
			podInfo[pod].upstreams = append(podInfo[pod].upstreams, us)
		}
	}

	var appMeshUpstreamList gloov1.UpstreamList
	for us := range appMeshUpstreamInfo {
		appMeshUpstreamList = append(appMeshUpstreamList, us)
	}

	sort.SliceStable(appMeshUpstreamList, func(i, j int) bool {
		return namespaceId(appMeshUpstreamList[i].Metadata) < namespaceId(appMeshUpstreamList[j].Metadata)
	})

	return appMeshUpstreamInfo, appMeshUpstreamList

}
