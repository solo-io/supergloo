package extensions

import (
	"context"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/traffictargetutils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	EchoServerPort     = uint32(1234)
	EchoServerHostname = "echo.example.com"
	DockerHostAddress = "host.docker.internal"
)

// testExtensionsServer is an e2e implementation of a grpc extensions service for Networking
// that adds a route to an echo server running on the local machine (reachable via `host.docker.internal` from inside KinD)
func createMeshPatches(ctx context.Context, mesh *v1alpha2.MeshSpec) (istio.Builder, error) {

	istioMesh := mesh.GetIstio()
	if istioMesh == nil {
		return nil, nil
	}

	resourceCluster := istioMesh.GetInstallation().GetCluster()
	resourceNamespace := istioMesh.GetInstallation().GetNamespace()
	resourceName := "echo-server"

	resourceMeta := metav1.ObjectMeta{
		Name:        resourceName,
		Namespace:   resourceNamespace,
		ClusterName: resourceCluster,
		Labels:      metautils.TranslatedObjectLabels(),
	}

	outputs := istio.NewBuilder(ctx, "test-extensions-server")

	serviceEntryIp, err := traffictargetutils.ConstructUniqueIpForKubeService(&v1.ClusterObjectRef{
		Name:        resourceName,
		Namespace:   resourceNamespace,
		ClusterName: resourceCluster,
	})
	if err != nil {
		return nil, err
	}

	// create a service entry to represent our local echo server
	serviceEntry := &istionetworkingv1alpha3.ServiceEntry{
		ObjectMeta: resourceMeta,
		Spec: networkingv1alpha3spec.ServiceEntry{
			Hosts:     []string{EchoServerHostname},
			Addresses: []string{serviceEntryIp.String()},
			Ports: []*networkingv1alpha3spec.Port{{
				Number:     EchoServerPort,
				Protocol:   "HTTP",
				Name:       "http",
				TargetPort: EchoServerPort,
			}},
			Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
			Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
			Endpoints: []*networkingv1alpha3spec.WorkloadEntry{{
				Address: DockerHostAddress,
				Ports:   map[string]uint32{"http": EchoServerPort},
			}},
		},
	}

	// create a virtual service to route to our local service entry
	virtualService := &istionetworkingv1alpha3.VirtualService{
		ObjectMeta: resourceMeta,
		Spec: networkingv1alpha3spec.VirtualService{
			Hosts:    []string{EchoServerHostname},
			Gateways: []string{"istio-system/istio-ingressgateway"},
			Http: []*networkingv1alpha3spec.HTTPRoute{{
				Name: "echo-server-route",
				Match: []*networkingv1alpha3spec.HTTPMatchRequest{{
					Uri:                  &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Prefix{Prefix: "/"}},
				}},
				Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
					Destination:          &networkingv1alpha3spec.Destination{
						Host:                 EchoServerHostname,
					},
				}},
			}},
		},
	}

	outputs.AddServiceEntries(serviceEntry)
	outputs.AddVirtualServices(virtualService)

	return outputs, nil
}
