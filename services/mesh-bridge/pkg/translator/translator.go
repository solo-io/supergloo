package translator

import (
	"context"
	"fmt"

	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup/config"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type Translator interface {
	Translate(ctx context.Context, meshBridgesByNamespace MeshBridgesByNamespace) (ServiceEntriesByNamespace, error)
}

func NewMeshBridgeTranslator(clientset config.ClientSet) Translator {
	return &translator{
		clientset: clientset,
	}
}

type translator struct {
	clientset config.ClientSet
}

type MeshBridgesByNamespace map[string]v1.MeshBridgeList
type ServiceEntriesByNamespace map[string]v1alpha3.ServiceEntryList

func (t *translator) Translate(ctx context.Context, meshBridgesByNamespace MeshBridgesByNamespace) (ServiceEntriesByNamespace, error) {
	result := make(ServiceEntriesByNamespace)
	ipGen := newIpGenerator()
	for namespace, meshBridges := range meshBridgesByNamespace {
		serviceEntries, err := t.meshBridgesToServiceEntry(ctx, namespace, meshBridges, ipGen)
		if err != nil {
			return nil, err
		}
		result[namespace] = append(result[namespace], serviceEntries...)
	}
	return result, nil
}

func (t *translator) meshBridgesToServiceEntry(ctx context.Context, namespace string, meshBridges v1.MeshBridgeList,
	generator *ipGenerator) ([]*v1alpha3.ServiceEntry, error) {

	var result []*v1alpha3.ServiceEntry

	uniqueAddresses := make(map[string]*BridgesWithExits)
	for _, meshBridge := range meshBridges {
		mesh, err := t.clientset.Mesh().Read(meshBridge.GetTargetMesh().GetResource().GetNamespace(),
			meshBridge.GetTargetMesh().GetResource().GetName(), clients.ReadOpts{})
		if err != nil {
			return nil, err
		}

		entryPoint, err := t.clientset.MeshIngress().Read(mesh.GetEntryPoint().GetResource().GetNamespace(),
			mesh.GetEntryPoint().GetResource().GetName(), clients.ReadOpts{})
		if err != nil {
			return nil, err
		}

		if bridgeWithExit, ok := uniqueAddresses[entryPoint.GetAddress()]; ok {
			bridgeWithExit.meshBridges = append(bridgeWithExit.meshBridges, meshBridge)
			bridgeWithExit.ports = append(bridgeWithExit.ports, entryPoint.GetPort())
		} else {
			uniqueAddresses[entryPoint.GetAddress()] = &BridgesWithExits{
				meshBridges: v1.MeshBridgeList{meshBridge},
				ports:       []uint32{entryPoint.GetPort()},
			}
		}
	}

	for address, bridges := range uniqueAddresses {
		info, err := t.infoFromTargetServices(bridges.meshBridges)
		if err != nil {
			return nil, err
		}
		ports := make(map[string]uint32)
		for i, v := range bridges.ports {
			ports[fmt.Sprintf("http%d", i+1)] = v
		}
		serviceEntry := &v1alpha3.ServiceEntry{
			Metadata: core.Metadata{
				Name:      addressToServiceEntryName(address, namespace),
				Namespace: namespace,
			},
			Hosts:      info.hosts,
			Addresses:  []string{generator.nextIp()},
			Ports:      info.ports,
			Location:   v1alpha3.ServiceEntry_MESH_INTERNAL,
			Resolution: v1alpha3.ServiceEntry_DNS,
			Endpoints: []*v1alpha3.ServiceEntry_Endpoint{
				{
					Address: address,
					Ports:   ports,
				},
			},
		}
		result = append(result, serviceEntry)
	}
	return result, nil
}

type BridgesWithExits struct {
	meshBridges v1.MeshBridgeList
	ports       []uint32
}

type UpstreamInfo struct {
	hosts []string
	ports []*v1alpha3.Port
}

func (t *translator) infoFromTargetServices(list v1.MeshBridgeList) (*UpstreamInfo, error) {
	hosts := make(map[string]bool)
	var endpoints []*v1alpha3.Port
	for i, bridge := range list {
		upstream, err := t.clientset.Upstreams().Read(bridge.GetTarget().GetResource().GetNamespace(),
			bridge.GetTarget().GetResource().GetName(), clients.ReadOpts{
				Cluster: bridge.GetTarget().GetCluster(),
			})
		if err != nil {
			return nil, err
		}
		kubeUpstream := upstream.GetUpstreamSpec().GetKube()
		if kubeUpstream == nil {
			return nil, fmt.Errorf("currently only kube upstreams are supported, %s supplied",
				bridge.GetTarget().String())
		}
		hostName := fmt.Sprintf("%s.%s.global", kubeUpstream.GetServiceName(), kubeUpstream.GetServiceNamespace())
		hosts[hostName] = true
		endpoints = append(endpoints, &v1alpha3.Port{
			Number:   kubeUpstream.GetServicePort(),
			Protocol: "http",
			Name:     fmt.Sprintf("http%d", i+1),
		})
	}
	var result []string
	for service, _ := range hosts {
		result = append(result, service)
	}
	return &UpstreamInfo{
		hosts: result,
		ports: endpoints,
	}, nil
}

func addressToServiceEntryName(address string, namespace string) string {
	return fmt.Sprintf("%s-mesh-bridge-%s", namespace, address)
}
