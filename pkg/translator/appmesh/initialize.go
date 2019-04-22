package appmesh

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/appmesh"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
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
	}

	return info, nil
}

func namespaceId(meta core.Metadata) string {
	return fmt.Sprintf("%s.%s", meta.Namespace, meta.Name)
}

func getUpstreamsForMesh(upstreams gloov1.UpstreamList, podInfo AwsAppMeshPodInfo, appMeshPodList v1.PodList) (AwsAppMeshUpstreamInfo, gloov1.UpstreamList, error) {
	appMeshUpstreamInfo := make(AwsAppMeshUpstreamInfo, 0)
	for _, us := range upstreams {

		// Get all the appMesh pods for this upstream
		pods := utils.PodsForUpstreams(gloov1.UpstreamList{us}, appMeshPodList)
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

	sort.SliceStable(appMeshUpstreamList, func(i, j int) bool {
		return namespaceId(appMeshUpstreamList[i].Metadata) < namespaceId(appMeshUpstreamList[j].Metadata)
	})

	return appMeshUpstreamInfo, appMeshUpstreamList, nil

}

func getPodsForMesh(mesh *v1.Mesh, pods v1.PodList) (AwsAppMeshPodInfo, v1.PodList, error) {
	var appMeshPodList v1.PodList
	appMeshPods := make(AwsAppMeshPodInfo, 0)
	for _, pod := range pods {
		kubePod, err := kubernetes.ToKubePod(pod)
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

	sort.SliceStable(appMeshPodList, func(i, j int) bool {
		iPod := appMeshPodList[i]
		jPod := appMeshPodList[j]
		return fmt.Sprintf("%s.%s", iPod.Namespace, iPod.Name) < fmt.Sprintf("%s.%s", jPod.Namespace, jPod.Name)
	})

	return appMeshPods, appMeshPodList, nil
}

func upstreamForVNLabel(upstreams gloov1.UpstreamList, vnLabel, vnName string) (*gloov1.Upstream, error) {
	for _, us := range upstreams {
		labels := utils.GetLabelsForUpstream(us)
		for key, value := range labels {
			if key == vnLabel && value == vnName {
				return us, nil
			}
		}
	}
	return nil, errors.Errorf("unable to find upstream with selector %s", vnLabel)
}

func podAppmeshConfig(info *podInfo, meshName, hostName string) (*appmesh.VirtualNodeData, *appmesh.VirtualServiceData) {
	var vn *appmesh.VirtualNodeData
	var vs *appmesh.VirtualServiceData
	listeners := make([]*appmesh.Listener, len(info.ports))
	for i, v := range info.ports {
		port := int64(v)
		protocol := "http"
		listeners[i] = &appmesh.Listener{
			PortMapping: &appmesh.PortMapping{
				Protocol: &protocol,
				Port:     &port,
			},
		}
	}
	vn = &appmesh.VirtualNodeData{
		MeshName:        &meshName,
		VirtualNodeName: &info.virtualNodeName,
		Spec: &appmesh.VirtualNodeSpec{
			// Backends get added back in later
			// configuration.go:222
			Backends:  []*appmesh.Backend{},
			Listeners: listeners,
			ServiceDiscovery: &appmesh.ServiceDiscovery{
				Dns: &appmesh.DnsServiceDiscovery{
					Hostname: &hostName,
				},
			},
		},
	}

	vs = &appmesh.VirtualServiceData{
		MeshName:           &meshName,
		VirtualServiceName: &hostName,
		Spec: &appmesh.VirtualServiceSpec{
			Provider: &appmesh.VirtualServiceProvider{
				VirtualNode: &appmesh.VirtualNodeServiceProvider{
					VirtualNodeName: &info.virtualNodeName,
				},
			},
		},
	}
	return vn, vs
}
