package translator

import (
	"context"
	"fmt"

	"github.com/solo-io/mesh-projects/pkg/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type Translator interface {
	Translate(ctx context.Context, meshBridges v1.MeshBridgeList) ([]*v1alpha3.ServiceEntry, error)
}

func NewMeshBridgeTranslator() Translator {
	return &translator{}
}

type translator struct {
}

type MeshBridgesByNamespace map[string]v1.MeshBridgeList

func (t *translator) Translate(ctx context.Context, meshBridges v1.MeshBridgeList) ([]*v1alpha3.ServiceEntry, error) {
	var result []*v1alpha3.ServiceEntry
	meshBridgesByNamespace := t.getMeshBridgesByNamespace(ctx, meshBridges)
	for namespace, meshBridges := range meshBridgesByNamespace {
		serviceEntries, err := t.meshBridgesToServiceEntry(ctx, namespace, meshBridges)
		if err != nil {
			return nil, err
		}
		result = append(result, serviceEntries...)
	}
	return result, nil
}

func (t *translator) meshBridgesToServiceEntry(ctx context.Context, namespace string, meshBridges v1.MeshBridgeList) (
	[]*v1alpha3.ServiceEntry, error) {

	var result []*v1alpha3.ServiceEntry

	uniqueAddresses := make(map[string]v1.MeshBridgeList)
	// for _, meshBridge := range meshBridges {
	// 	uniqueAddresses[meshBridge.GetAddress()] = append(uniqueAddresses[meshBridge.GetAddress()], meshBridge)
	// }

	for address, _ := range uniqueAddresses {
		serviceEntry := &v1alpha3.ServiceEntry{
			Metadata: core.Metadata{
				Name:      addressToServiceEntryName(address),
				Namespace: namespace,
			},
			Hosts:     []string{address},
			Addresses: []string{"240.0.0.2"},
			Ports: []*v1alpha3.Port{
				{
					Number:   9080,
					Protocol: "http",
					Name:     "http1",
				},
			},
			Location:   v1alpha3.ServiceEntry_MESH_INTERNAL,
			Resolution: v1alpha3.ServiceEntry_DNS,
			Endpoints: []*v1alpha3.ServiceEntry_Endpoint{
				{
					Ports: map[string]uint32{
						"http1": 15443,
					},
				},
			},
		}
		result = append(result, serviceEntry)
	}
	return result, nil
}

func addressToServiceEntryName(address string) string {
	return fmt.Sprintf("%s-mesh-bridge", address)
}

func (t *translator) getMeshBridgesByNamespace(ctx context.Context, meshBridges v1.MeshBridgeList) MeshBridgesByNamespace {
	meshBridgeMap := make(MeshBridgesByNamespace)
	for _, v := range meshBridges {
		meshBridgeMap[v.Metadata.GetNamespace()] = append(meshBridgeMap[v.Metadata.GetNamespace()], v)
	}
	return meshBridgeMap
}
