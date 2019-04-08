package appmesh

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/appmesh"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
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
	// Handle appmesh routing rule
	ProcessRoutingRules(rule v1.RoutingRuleList) error
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
	VirtualNodes    []*appmesh.VirtualNodeData
	VirtualServices []*appmesh.VirtualServiceData
	VirtualRouters  []*appmesh.VirtualRouterData
	Routes          []*appmesh.RouteData
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

func (c *awsAppMeshConfiguration) ProcessRoutingRules(rule v1.RoutingRuleList) error {
	for _, us := range c.upstreamList {
		if err := c.createVirtualServiceForHost(us); err != nil {
			return err
		}
	}

	// matcher, err := createAppmeshMatcher(rule)
	// if err != nil {
	// 	return err
	// }
	// route := &appmesh.HttpRoute{
	// 	Match: matcher,
	// }
	//
	// switch typedRule := rule.GetSpec().GetRuleType().(type) {
	// case *v1.RoutingRuleSpec_TrafficShifting:
	// 	if err := processTrafficShiftingRule(c.upstreamList, c.VirtualNodes, typedRule.TrafficShifting, route); err != nil {
	// 		return err
	// 	}
	// default:
	// 	return fmt.Errorf("currently only traffic shifting rules are supported by appmesh, found %t", typedRule)
	// }
	//
	// virtualRouter := &appmesh.VirtualRouterData{
	// 	MeshName: &c.MeshName,
	// 	VirtualRouterName: appmeshVirtualRouterName(rule),
	// 	Spec:
	// }
	//
	// routeData := &appmesh.RouteData{
	// 	VirtualRouterName: virtualRouter.VirtualRouterName,
	// }

	return nil
}

func appmeshRouteName(rule *v1.RoutingRule) *string {
	name := fmt.Sprintf("%s.%s-route", rule.Metadata.Namespace, rule.Metadata.Name)
	return &name
}

func appmeshVirtualRouterName(rule *v1.RoutingRule) *string {
	name := fmt.Sprintf("%s.%s-vr", rule.Metadata.Namespace, rule.Metadata.Name)
	return &name
}

func (c *awsAppMeshConfiguration) AllowAll() error {

	// TODO: loop over podInfo, create a VirtualNode for every unique virtualNodeName, lookup (and validate) the upstreams for the pod
	//  to get the serviceDiscovery.dns.hostname and ports (need to validate these against the pod ports). Then create a VS for
	//  the Virtual node.
	//  Lastly, iterate over all vn/vs and add all VSs as back ends for all the VNs (excepts for the VS that maps to the VN)

	if err := c.connectAllVirtualNodes(); err != nil {
		return err
	}

	return nil
}

func (c *awsAppMeshConfiguration) connectAllVirtualNodes() error {

	for _, vn := range c.VirtualNodes {
		for _, vs := range c.VirtualServices {
			if vs.Spec.Provider.VirtualNode.VirtualNodeName == vn.VirtualNodeName {
				continue
			}
			backend := &appmesh.Backend{
				VirtualService: &appmesh.VirtualServiceBackend{
					VirtualServiceName: vs.VirtualServiceName,
				},
			}
			vn.Spec.Backends = append(vn.Spec.Backends, backend)
		}
	}

	return nil
}

func (c *awsAppMeshConfiguration) createVirtualServiceForHost(upstream *gloov1.Upstream) error {
	host, err := utils.GetHostForUpstream(upstream)
	if err != nil {
		return err
	}

	pods, ok := c.upstreamInfo[upstream]
	if !ok {
		return nil
	}

	for _, pod := range pods {
		info, ok := c.podInfo[pod]
		if !ok {
			continue
		}

		vn, vs := podAppmeshConfig(info, c.MeshName, host)
		c.VirtualNodes = append(c.VirtualNodes, vn)

		c.VirtualServices = append(c.VirtualServices, vs)
	}

	return nil
}

func podAppmeshConfig(info *podInfo, meshName, hostName string) (*appmesh.VirtualNodeData, *appmesh.VirtualServiceData) {
	var vn *appmesh.VirtualNodeData
	var vs *appmesh.VirtualServiceData
	vn = &appmesh.VirtualNodeData{
		MeshName:        &meshName,
		VirtualNodeName: &info.virtualNodeName,
		Spec: &appmesh.VirtualNodeSpec{
			Backends:  []*appmesh.Backend{},
			Listeners: []*appmesh.Listener{},
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
