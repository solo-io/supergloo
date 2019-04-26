package appmesh

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/translator/utils"
)

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

func groupByVirtualNodeName(podInfo awsAppMeshPodInfo) map[string]awsAppMeshPodInfo {
	byVirtualNodeName := make(map[string]awsAppMeshPodInfo)
	for pod, info := range podInfo {

		vnName := info.virtualNodeName

		if _, ok := byVirtualNodeName[vnName]; !ok {
			byVirtualNodeName[vnName] = make(awsAppMeshPodInfo, 0)
		}

		byVirtualNodeName[vnName][pod] = info
	}

	return byVirtualNodeName
}

func initializeVirtualNodes(meshName string, podInfoByVnName map[string]awsAppMeshPodInfo) (virtualNodeByHost, error) {
	virtualNodes := make(virtualNodeByHost)

	for virtualNodeName, podInfo := range podInfoByVnName {
		services := make(map[string]bool)
		upstreamPorts := make(map[uint32]bool)
		appmeshAppPorts := make(map[uint32]bool)

		// For each pod that belongs to the virtual node
		for _, info := range podInfo {

			// Collect upstream service names and ports
			for _, us := range info.upstreams {
				host, err := utils.GetHostForUpstream(us)
				if err != nil {
					return nil, err
				}
				port, err := utils.GetPortForUpstream(us)
				if err != nil {
					return nil, err
				}

				services[host] = true
				upstreamPorts[port] = true
			}

			// Collect all ports specified via the APPMESH_APP_PORTS env on the pod
			for _, port := range info.ports {
				appmeshAppPorts[port] = true
			}
		}

		// Validate hostname
		var hostname string
		switch len(services) {
		case 0:
			return nil, errors.Errorf("failed to determine service discovery information for virtual node %s: "+
				"no services match its pods", virtualNodeName)
		case 1:
			for name := range services {
				hostname = name
				break
			}
		default:
			names := make([]string, 0)
			for name := range services {
				names = append(names, name)
			}
			return nil, errors.Errorf("failed to determine service discovery information for virtual node %s: "+
				"multiple services match its pods: %v", virtualNodeName, fmt.Sprintf(strings.Join(names, ",")))
		}

		// TODO: virtual nodes can only have one listener, but a) a pod can expose multiple ports and b) APPMESH_APP_PORTS
		//  can contain multiple ports. What are we supposed to do here? Enforce only a single port?
		// Note: gloo UDS creates an upstream for every port defined on a service. This is why we collect all the ports
		// for the upstreams that match the Virtual Node pods and verify that the resulting set contains all the ports
		// specified in the APPMESH_APP_PORTS envs.
		var ports []uint32
		for requiredPort := range appmeshAppPorts {
			if _, ok := upstreamPorts[requiredPort]; !ok {
				return nil, errors.Errorf("service %s does not define mapping for port %d required by virtual node %s",
					hostname, requiredPort, virtualNodeName)
			}
			ports = append(ports, requiredPort)
		}

		virtualNodes[hostname] = createVirtualNode(ports, virtualNodeName, meshName, hostname)
	}

	return virtualNodes, nil
}
