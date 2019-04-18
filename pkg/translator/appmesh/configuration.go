package appmesh

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/pkg/errors"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
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

type AwsAppMeshPodInfo map[*v1.Pod]*podInfo
type AwsAppMeshUpstreamInfo map[*gloov1.Upstream][]*v1.Pod

type PodVirtualNode map[*v1.Pod]*appmesh.VirtualNodeData
type PodVirtualService map[*v1.Pod]*appmesh.VirtualServiceData

type VirtualNodeByHost map[string]*appmesh.VirtualNodeData
type VirtualServiceByHost map[string]*appmesh.VirtualServiceData

type AwsAppMeshConfiguration interface {
	// Configure resources to allow traffic from/to all services in the mesh
	AllowAll() error
	// Handle appmesh routing rule
	ProcessRoutingRules(rules v1.RoutingRuleList) error
}

// Represents the output of the App Mesh translator
type awsAppMeshConfiguration struct {
	// We build these objects once in the constructor. They are meant to help in all the translation operations where we
	// probably will need to look up pods by upstreams and vice-versa multiple times.
	podInfo         AwsAppMeshPodInfo
	podInfoByVnName map[string]AwsAppMeshPodInfo
	podList         v1.PodList
	upstreamInfo    AwsAppMeshUpstreamInfo
	upstreamList    gloov1.UpstreamList

	// These are the actual results of the translations
	MeshName           string
	OldVirtualNodes    PodVirtualNode
	OldVirtualServices PodVirtualService
	VirtualNodes       VirtualNodeByHost
	VirtualServices    VirtualServiceByHost
	VirtualRouters     []*appmesh.VirtualRouterData
	Routes             []*appmesh.RouteData
}

func NewAwsAppMeshConfiguration(meshName string, pods v1.PodList, upstreams gloov1.UpstreamList) (AwsAppMeshConfiguration, error) {

	// Get all pods that point to this mesh via the APPMESH_VIRTUAL_NODE_NAME env set on their AWS App Mesh sidecar.
	appMeshPodInfo, appMeshPodList, err := getPodsForMesh(meshName, pods)
	if err != nil {
		return nil, err
	}

	// Find all upstreams that are associated with the appmesh pods
	// Also updates each podInfo in appMeshPodInfo with the list of upstreams that match it
	appMeshUpstreamInfo, appMeshUpstreamList := getUpstreamsForMesh(upstreams, appMeshPodInfo, appMeshPodList)

	// Group pods by the virtual nodes they belong to
	podInfoByVirtualNode := groupByVirtualNodeName(appMeshPodInfo)

	// Create the virtual node objects. These will be updated later.
	virtualNodes, err := initializeVirtualNodes(meshName, podInfoByVirtualNode)
	if err != nil {
		return nil, err
	}

	config := &awsAppMeshConfiguration{
		podInfo:         appMeshPodInfo,
		podInfoByVnName: podInfoByVirtualNode,
		podList:         appMeshPodList,
		upstreamInfo:    appMeshUpstreamInfo,
		upstreamList:    appMeshUpstreamList,

		MeshName:     meshName,
		VirtualNodes: virtualNodes,
	}

	return config, nil
}

func (c *awsAppMeshConfiguration) ProcessRoutingRules(rules v1.RoutingRuleList) error {
	for _, rule := range rules {
		if err := c.processRoutingRule(rule); err != nil {
			return err
		}
	}
	return nil
}

func (c *awsAppMeshConfiguration) processRoutingRule(rule *v1.RoutingRule) error {

	matcher, err := createAppmeshMatcher(rule)
	if err != nil {
		return err
	}
	route := &appmesh.HttpRoute{
		Match: matcher,
	}

	// create appmesh routes based on v1 rule
	switch typedRule := rule.GetSpec().GetRuleType().(type) {
	case *v1.RoutingRuleSpec_TrafficShifting:
		if err := processTrafficShiftingRule(c.upstreamList, c.OldVirtualNodes, typedRule.TrafficShifting, route); err != nil {
			return err
		}
	default:
		return fmt.Errorf("currently only traffic shifting rules are supported by appmesh, found %t", typedRule)
	}

	// apply the appmesh routes to the relevant pods
	for _, pod := range c.podList {
		vn := c.OldVirtualNodes[pod]
		matches, err := podMatchesRule(vn, rule.DestinationSelector, c.upstreamList)
		if err != nil {
			return err
		}
		if !matches {
			continue
		}

		virtualRouter := &appmesh.VirtualRouterData{
			MeshName:          &c.MeshName,
			VirtualRouterName: appmeshVirtualRouterName(rule, c.podInfo[pod]),
			Spec: &appmesh.VirtualRouterSpec{
				Listeners: listenerToRouterListener(vn.Spec.Listeners),
			},
		}

		routeData := &appmesh.RouteData{
			VirtualRouterName: virtualRouter.VirtualRouterName,
			MeshName:          &c.MeshName,
			RouteName:         appmeshRouteName(rule, c.podInfo[pod]),
			Spec: &appmesh.RouteSpec{
				HttpRoute: route,
			},
		}
		c.VirtualRouters = append(c.VirtualRouters, virtualRouter)
		c.Routes = append(c.Routes, routeData)

		vs := c.OldVirtualServices[pod]
		vs.Spec = &appmesh.VirtualServiceSpec{
			Provider: &appmesh.VirtualServiceProvider{
				VirtualRouter: &appmesh.VirtualRouterServiceProvider{
					VirtualRouterName: appmeshVirtualRouterName(rule, c.podInfo[pod]),
				},
			},
		}
	}

	return nil
}

func listenerToRouterListener(listeners []*appmesh.Listener) []*appmesh.VirtualRouterListener {
	vrListeners := make([]*appmesh.VirtualRouterListener, len(listeners))
	for i, listener := range listeners {
		vrListeners[i] = &appmesh.VirtualRouterListener{
			PortMapping: listener.PortMapping,
		}
	}
	return vrListeners
}

func podMatchesRule(vnode *appmesh.VirtualNodeData, selector *v1.PodSelector, upstreams gloov1.UpstreamList) (bool, error) {
	selectorType := selector.GetSelectorType()
	if selectorType == nil {
		return false, errors.Errorf("pod selector type cannot be nil")
	}

	destinationHost := *vnode.Spec.ServiceDiscovery.Dns.Hostname
	return utils.RuleAppliesToDestination(destinationHost, selector, upstreams)
}

func appmeshRouteName(rule *v1.RoutingRule, pod *podInfo) *string {
	name := fmt.Sprintf("%s.%s.%s-route", rule.Metadata.Namespace, rule.Metadata.Name,
		pod.virtualNodeName)
	return &name
}

func appmeshVirtualRouterName(rule *v1.RoutingRule, pod *podInfo) *string {
	name := fmt.Sprintf("%s.%s.%s-vr", rule.Metadata.Namespace, rule.Metadata.Name,
		pod.virtualNodeName)
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

	for _, vn := range c.OldVirtualNodes {
		for _, vs := range c.OldVirtualServices {
			// check if Virtual node is nil, which means virtual router is set and that name is set
			if vs.Spec.Provider.VirtualNode != nil && vs.Spec.Provider.VirtualNode.VirtualNodeName == vn.VirtualNodeName {
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
