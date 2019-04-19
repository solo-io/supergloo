package appmesh

import (
	"fmt"
	"strings"

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

type awsAppMeshPodInfo map[*v1.Pod]*podInfo
type awsAppMeshUpstreamInfo map[*gloov1.Upstream][]*v1.Pod

type virtualNodeByHost map[string]*appmesh.VirtualNodeData
type virtualServiceByHost map[string]*appmesh.VirtualServiceData

type routesByPort map[uint32][]*appmesh.HttpRoute
type routesByDestinationAndPort map[string]routesByPort

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
	upstreamInfo awsAppMeshUpstreamInfo
	upstreamList gloov1.UpstreamList

	// These are the actual results of the translations
	MeshName        string
	VirtualNodes    virtualNodeByHost
	VirtualServices virtualServiceByHost
	VirtualRouters  []*appmesh.VirtualRouterData
	Routes          []*appmesh.RouteData
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

	return &awsAppMeshConfiguration{
		upstreamInfo: appMeshUpstreamInfo,
		upstreamList: appMeshUpstreamList,

		MeshName:        meshName,
		VirtualNodes:    virtualNodes,
		VirtualServices: make(virtualServiceByHost),
	}, nil
}

func (c *awsAppMeshConfiguration) ProcessRoutingRules(rules v1.RoutingRuleList) error {

	routeMap := make(routesByDestinationAndPort)
	for _, rule := range rules {

		routesByDestination, listenerPort, err := processRoutingRule(rule, c.upstreamList, c.VirtualNodes)
		if err != nil {
			return errors.Wrapf(err, "failed to process routing role %s", rule.Metadata.Ref())
		}

		// merge the results for this route into the map
		for destinationHost, routes := range routesByDestination {
			if _, ok := routeMap[destinationHost]; !ok {
				routeMap[destinationHost] = make(routesByPort)
			}
			routeMap[destinationHost][listenerPort] = append(routeMap[destinationHost][listenerPort], routes...)
		}
	}

	// For each destination host:
	//   1. create a Virtual Service with a Virtual Router that groups all the Routes for that host
	//   2. set the Virtual Services as a Backends on all Virtual Nodes (excepts the one with the same hostname as the VS)
	for destinationHost, routesByPort := range routeMap {

		port, routes, err := validate(destinationHost, routesByPort)
		if err != nil {
			return err
		}

		// Create Virtual Router
		virtualRouter := createVirtualRouter(c.MeshName, destinationHost, port)
		c.VirtualRouters = append(c.VirtualRouters, virtualRouter)

		// Create Routes
		c.Routes = append(c.Routes, createRoutes(c.MeshName, destinationHost, *virtualRouter.VirtualRouterName, routes)...)

		// Create Virtual Service
		virtualService := createVirtualServiceWithVirtualRouterProvider(c.MeshName, destinationHost, *virtualRouter.VirtualRouterName)
		c.VirtualServices[destinationHost] = virtualService

		// Add the Virtual Service as Backend for all the relevant Virtual Nodes
		for host, virtualNode := range c.VirtualNodes {

			// Don't add a VN as a backend to itself
			if host == destinationHost {
				continue
			}

			virtualNode.Spec.Backends = append(virtualNode.Spec.Backends, &appmesh.Backend{
				VirtualService: &appmesh.VirtualServiceBackend{
					VirtualServiceName: virtualService.VirtualServiceName,
				},
			})
		}
	}

	return nil
}

func (c *awsAppMeshConfiguration) AllowAll() error {

	// Create missing Virtual Services
	for host := range c.VirtualNodes {
		if _, ok := c.VirtualServices[host]; !ok {
			c.VirtualServices[host] = createVirtualServiceWithVirtualNodeProvider(c.MeshName, host, host)
		}
	}

	// Add all Virtual Services as backends (upstream dependencies) for all Virtual Nodes
	for vnHost, vn := range c.VirtualNodes {

		// For faster lookup
		vnBackends := make(map[string]bool)
		for _, backend := range vn.Spec.Backends {
			vnBackends[*backend.VirtualService.VirtualServiceName] = true
		}

		for vsHost, vs := range c.VirtualServices {

			// Don't add Virtual Services that have the same host as the Virtual Node to the Backends
			if vnHost == vsHost {
				continue
			}

			// Skip if the Virtual Node already has a Backend for this hostname.
			// This means that it was created as part of a Routing Rule.
			if _, ok := vnBackends[*vs.VirtualServiceName]; ok {
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

// TODO: handle source selectors
// Returns:
//  - a map where each entry represents a routing destination and the correspondent value all the routes to that destination
//  - a port that will be used when building the Virtual Routers(s) that the Routes will be assigned to. We could have
//    multiple Virtual Routers if this RoutingRule matches multiple destinations; in that case each destination will
//    yield a Virtual Service and a Virtual Router (which will be associated with a copy of the Route set)
func processRoutingRule(rule *v1.RoutingRule, upstreams gloov1.UpstreamList, virtualNodes virtualNodeByHost) (
	map[string][]*appmesh.HttpRoute, uint32, error) {

	matchers, err := buildAppmeshMatchers(rule)
	if err != nil {
		return nil, 0, err
	}

	// Create the route action. It will potentially be reused for several routes.
	var routeAction *appmesh.HttpRouteAction
	var port uint32
	switch typedRule := rule.GetSpec().GetRuleType().(type) {
	case *v1.RoutingRuleSpec_TrafficShifting:
		routeAction, port, err = processTrafficShiftingRule(upstreams, virtualNodes, typedRule.TrafficShifting)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, errors.Errorf("found unsupported rule type %T. Currently only traffic shifting rules are"+
			" supported for AWS AppMesh ", typedRule)
	}

	// Create a route for each matcher. They will all be bound to the same virtual router.
	var routes []*appmesh.HttpRoute
	for _, matcher := range matchers {
		routes = append(routes, &appmesh.HttpRoute{
			Match:  matcher,
			Action: routeAction,
		})
	}

	// Get all upstreams matching the destination selectors
	destUpstreams, err := utils.UpstreamsForSelector(rule.DestinationSelector, upstreams)
	if err != nil {
		// Error is only thrown if it's an upstream selector and the upstream could not be found in the given list
		return nil, 0, errors.Wrapf(err, "the destination selector for routing rule %s does not match any pod injected "+
			"with the App Mesh sidecar", rule.Metadata.Ref())
	}
	uniqueHostnames := make(map[string]bool)
	for _, destUpstream := range destUpstreams {
		host, err := utils.GetHostForUpstream(destUpstream)
		if err != nil {
			return nil, 0, err
		}
		uniqueHostnames[host] = true
	}

	result := make(map[string][]*appmesh.HttpRoute)
	for destinationHost := range uniqueHostnames {
		result[destinationHost] = routes
	}

	return result, port, nil
}

func buildAppmeshMatchers(rule *v1.RoutingRule) ([]*appmesh.HttpRouteMatch, error) {
	var awsMatchers []*appmesh.HttpRouteMatch
	for _, matcher := range rule.GetRequestMatchers() {
		pathSpecifier := matcher.GetPathSpecifier()
		if pathSpecifier == nil {
			return nil, errors.Errorf("path specifier for routing rule %s cannot be nil", rule.Metadata.Ref())
		}
		switch matchType := pathSpecifier.(type) {
		case *gloov1.Matcher_Prefix:
			awsMatchers = append(awsMatchers, &appmesh.HttpRouteMatch{Prefix: &matchType.Prefix})
		default:
			return nil, errors.Errorf("unsupported matcher type %T on routing rule %s. AppMesh currently "+
				"supports only path prefix matchers", matcher.GetPathSpecifier(), rule.Metadata.Ref())
		}
	}
	return awsMatchers, nil
}

// Fails if multiple rules target the given host on different ports.
func validate(host string, routeMap routesByPort) (port uint32, routes []*appmesh.HttpRoute, err error) {
	if len(routeMap) > 1 {
		var ports []string
		for port := range routeMap {
			ports = append(ports, fmt.Sprint(port))
		}
		err = errors.Errorf("the Routing Rules resulted in multiple Routes to different ports (%s) on host %s. "+
			"Supergloo cannot translate this configuration as AWS App Mesh currently requires a single listener "+
			"to be specified on the Virtual Router for a DNS service name (Virtual Service)",
			strings.Join(ports, ","), host)
		return
	}

	for p, r := range routeMap {
		port, routes = p, r
	}
	return
}

func createVirtualNode(ports []uint32, virtualNodeName, meshName, hostName string) *appmesh.VirtualNodeData {
	var vn *appmesh.VirtualNodeData
	listeners := make([]*appmesh.Listener, len(ports))
	for i, v := range ports {
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
		VirtualNodeName: &virtualNodeName,
		Spec: &appmesh.VirtualNodeSpec{
			// Backends get added back in later
			Backends:  []*appmesh.Backend{},
			Listeners: listeners,
			ServiceDiscovery: &appmesh.ServiceDiscovery{
				Dns: &appmesh.DnsServiceDiscovery{
					Hostname: &hostName,
				},
			},
		},
	}
	return vn
}

func createVirtualRouter(meshName, hostname string, port uint32) *appmesh.VirtualRouterData {
	portInt64 := int64(port)
	return &appmesh.VirtualRouterData{
		MeshName:          &meshName,
		VirtualRouterName: &hostname,
		Spec: &appmesh.VirtualRouterSpec{
			Listeners: []*appmesh.VirtualRouterListener{
				{
					PortMapping: &appmesh.PortMapping{
						Port: &portInt64,
					},
				},
			},
		},
	}
}

func createVirtualServiceWithVirtualRouterProvider(meshName, hostname, virtualRouterName string) *appmesh.VirtualServiceData {
	return &appmesh.VirtualServiceData{
		MeshName:           &meshName,
		VirtualServiceName: &hostname,
		Spec: &appmesh.VirtualServiceSpec{
			Provider: &appmesh.VirtualServiceProvider{
				VirtualRouter: &appmesh.VirtualRouterServiceProvider{
					VirtualRouterName: &virtualRouterName,
				},
			},
		},
	}
}

func createVirtualServiceWithVirtualNodeProvider(meshName, hostname, virtualNodeName string) *appmesh.VirtualServiceData {
	return &appmesh.VirtualServiceData{
		MeshName:           &meshName,
		VirtualServiceName: &hostname,
		Spec: &appmesh.VirtualServiceSpec{
			Provider: &appmesh.VirtualServiceProvider{
				VirtualNode: &appmesh.VirtualNodeServiceProvider{
					VirtualNodeName: &virtualNodeName,
				},
			},
		},
	}
}

func createRoutes(meshName, hostname, virtualRouterName string, routes []*appmesh.HttpRoute) (out []*appmesh.RouteData) {
	for i, route := range routes {
		name := fmt.Sprintf("%s-%d", hostname, i)
		routeData := &appmesh.RouteData{
			VirtualRouterName: &virtualRouterName,
			MeshName:          &meshName,
			RouteName:         &name,
			Spec: &appmesh.RouteSpec{
				HttpRoute: route,
			},
		}
		out = append(out, routeData)
	}
	return
}
