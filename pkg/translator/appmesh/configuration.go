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
)

type podInfo struct {
	// These come from the APPMESH_APP_PORTS envs on pods that have been injected
	ports []uint32
	// These come from the APPMESH_VIRTUAL_NODE_NAME envs on pods that have been injected
	virtualNodeName string
	// All the upstreams that match this pod
	upstreams gloov1.UpstreamList
}

type AwsAppMeshPodInfo map[*customkube.Pod]*podInfo
type AwsAppMeshUpstreamInfo map[*gloov1.Upstream][]*customkube.Pod

type AwsAppMeshConfiguration interface {
	// Configure resources to allow traffic from/to all services in the mesh
	AllowAll() error
}

// Represents the output of the App Mesh translator
type awsAppMeshConfiguration struct {
	// We build these objects once in the constructor. They are meant to help in all the translation operations where we
	// probably will need to look up pods by upstreams and vice-versa multiple times.
	podInfo      AwsAppMeshPodInfo
	podList      customkube.PodList
	upstreamInfo AwsAppMeshUpstreamInfo
	upstreamList gloov1.UpstreamList

	// These are the actual results of the translations
	MeshName        string
	VirtualNodes    []VirtualNode
	VirtualServices []VirtualService
	VirtualRouters  []VirtualRouter
	Routes          []Route
}

// TODO(marco): to Eitan: I have not tested the util methods used in here, sorry in advance if they do not work as expected
func NewAwsAppMeshConfiguration(mesh *v1.Mesh, pods customkube.PodList, upstreams gloov1.UpstreamList) (AwsAppMeshConfiguration, error) {

	// Get all pods that point to this mesh via the APPMESH_VIRTUAL_NODE_NAME env set on their AWS App Mesh sidecar.
	appMeshPodInfo, appMeshPodList, err := getPodsForMesh(mesh, pods)
	if err != nil {
		return nil, err
	}

	// Find all upstreams that are associated with the appmesh pods
	// Also updates each podInfo in appMeshPodInfo with the list of upstreams that match it
	appMeshUpstreamInfo, appMeshUpstreamList, err := getUpstreamsForMesh(upstreams, appMeshPodInfo, appMeshPodList)
	if err != nil {
		return nil, err
	}

	return &awsAppMeshConfiguration{
		podInfo:      appMeshPodInfo,
		podList:      appMeshPodList,
		upstreamInfo: appMeshUpstreamInfo,
		upstreamList: appMeshUpstreamList,

		MeshName: mesh.Metadata.Name,
	}, nil
}

func (c *awsAppMeshConfiguration) AllowAll() error {

	// TODO: loop over podInfo, create a VirtualNode for every unique virtualNodeName, lookup (and validate) the upstreams for the pod
	//  to get the serviceDiscovery.dns.hostname and ports (need to validate these against the pod ports). Then create a VS for
	//  the Virtual node.
	//  Lastly, iterate over all vn/vs and add all VSs as back ends for all the VNs (excepts for the VS that maps to the VN)
	return nil
}

func getPodsForMesh(mesh *v1.Mesh, pods customkube.PodList) (AwsAppMeshPodInfo, customkube.PodList, error) {
	var appMeshPodList customkube.PodList
	appMeshPods := make(AwsAppMeshPodInfo, 0)
	for _, pod := range pods {
		kubePod, err := kubernetes.ToKube(pod)
		if err != nil {
			return nil, nil, err
		}

		var info *podInfo
		for _, container := range kubePod.Spec.Containers {
			for _, env := range container.Env {
				if env.Name == PodVirtualNodeEnvName {

					// Env value is expected to have the following format
					// - name: "APPMESH_VIRTUAL_NODE_NAME"
					//   value: "mesh/meshName/virtualNode/virtualNodeName"
					vnNameParts := strings.Split(env.Value, "/")
					if len(vnNameParts) != 4 {
						return nil, nil, errors.Errorf("unexpected format for %s env for pod %s.%s. Expected format is [%s] but found [%s]",
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
								return nil, nil, errors.Wrapf(err, "failed to parse [%s] (value of %s env) to int array",
									env.Value, PodPortsEnvName)
							}
							info.ports = append(info.ports, uint32(ui64))
						}
					}
				}
			}

			appMeshPodList = append(appMeshPodList, pod)
			appMeshPods[pod] = info
		}
	}
	return appMeshPods, appMeshPodList, nil
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

// NOTE: these types do not attempt to recreate the structure of the App Mesh resources. They have a more flat structure
// but still contain all the information needed to build the correspondent App Mesh API types.
// TODO: whether to keep them this way depends on how we decide to implement the reconciler

type VirtualNode struct {
	// Name of the VirtualNode
	Name string
	// Name of the Mesh this VirtualNode belongs to
	//MeshName string TODO: come back and see whether these references are useful
	// DNS name used to discover the task group for the VirtualNode
	HostName string
	// List of protocol/ports the VirtualNode expects to receive inbound traffic on
	Listeners []Listener
	// Names of the VirtualServices that the VirtualNode expects to direct outbound traffic to
	Backends []string
}

type VirtualService struct {
	// Name of the VirtualService
	Name string
	// Name of the VirtualRouter that directs incoming requests for this VirtualService
	// to different VirtualNodes via Routes
	VirtualRouter string
}

type VirtualRouter struct {
	// Name of the VirtualRouter
	Name string
	// List of protocol/ports the VirtualRouter expects to receive inbound traffic on.
	Listeners []Listener
}

type Route struct {
	// Name of the Route
	Name string
	// Name of the VirtualRouter that directs incoming requests for this VirtualService to different VirtualNodes via Routes
	VirtualRouter string
	// Define how traffic is distributed to one or more target virtual nodes with relative weighting.
	WeightedTargets []WeightedTarget
	// Path that the route should match
	Prefix string
}

type Listener struct {
	Port     uint32
	Protocol string
}

type WeightedTarget struct {
	// Name of the VirtualNode
	VirtualNode string
	// Relative weight of this target with respect to the other targets defined for this Route.
	// All weight for the route will be normalized to add up to 100.
	Weight uint32
}
