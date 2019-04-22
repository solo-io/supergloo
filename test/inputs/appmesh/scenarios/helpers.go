package scenarios

import (
	"github.com/aws/aws-sdk-go/service/appmesh"
)

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

func createVirtualNode(name, host, mesh, protocol string, port int64, backendHosts []string) *appmesh.VirtualNodeData {
	var backends []*appmesh.Backend
	for _, vs := range backendHosts {
		vsName := vs
		backends = append(backends, &appmesh.Backend{
			VirtualService: &appmesh.VirtualServiceBackend{
				VirtualServiceName: &vsName,
			},
		})
	}
	return &appmesh.VirtualNodeData{
		MeshName:        &mesh,
		VirtualNodeName: &name,
		Spec: &appmesh.VirtualNodeSpec{
			ServiceDiscovery: &appmesh.ServiceDiscovery{
				Dns: &appmesh.DnsServiceDiscovery{
					Hostname: &host,
				},
			},
			Listeners: []*appmesh.Listener{
				{
					PortMapping: &appmesh.PortMapping{
						Port:     &port,
						Protocol: &protocol,
					},
				},
			},
			Backends: backends,
		},
	}
}

func createVirtualServiceWithVn(hostname, meshName, virtualNodeName string) *appmesh.VirtualServiceData {
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

func createVirtualServiceWithVr(hostname, meshName, virtualRouterName string) *appmesh.VirtualServiceData {
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

func createRoute(meshName, routeName, virtualRouterName, prefix string, action *appmesh.HttpRouteAction) *appmesh.RouteData {
	return &appmesh.RouteData{
		VirtualRouterName: &virtualRouterName,
		MeshName:          &meshName,
		RouteName:         &routeName,
		Spec: &appmesh.RouteSpec{
			HttpRoute: &appmesh.HttpRoute{
				Match: &appmesh.HttpRouteMatch{
					Prefix: &prefix,
				},
				Action: action,
			},
		},
	}
}

type vnWeightTuple struct {
	virtualNode string
	weight      int64
}

func createRouteAction(destinations []vnWeightTuple) *appmesh.HttpRouteAction {
	var targets []*appmesh.WeightedTarget
	for _, d := range destinations {
		dest := d
		targets = append(targets, &appmesh.WeightedTarget{
			VirtualNode: &dest.virtualNode,
			Weight:      &dest.weight,
		})
	}
	return &appmesh.HttpRouteAction{
		WeightedTargets: targets,
	}
}

// Returns a slice of all hosts except the given one
func allHostsMinus(excludedHost string) (out []string) {
	for _, host := range allHostnames {
		if host != excludedHost {
			out = append(out, host)
		}
	}
	return
}
