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
	HelloServerPort     = uint32(1234)
	HelloServerHostname = "hello.global"
	DockerHostAddress   = "host.docker.internal"
)

// testExtensionsServer is an e2e implementation of a grpc extensions service for Networking
// that adds a route to an HelloWorld server running on the local machine (reachable via `host.docker.internal` from inside KinD)
func getCreateMeshPatchesFunc() func(ctx context.Context, mesh *v1alpha2.MeshSpec) (istio.Builder, error) {
	return func(ctx context.Context, mesh *v1alpha2.MeshSpec) (istio.Builder, error) {
		istioMesh := mesh.GetIstio()
		if istioMesh == nil {
			return nil, nil
		}

		resourceCluster := istioMesh.GetInstallation().GetCluster()
		resourceNamespace := istioMesh.GetInstallation().GetNamespace()
		resourceName := "hello-server"

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

		// create a service entry to represent our local hello server
		serviceEntry := &istionetworkingv1alpha3.ServiceEntry{
			ObjectMeta: resourceMeta,
			Spec: networkingv1alpha3spec.ServiceEntry{
				Hosts:     []string{HelloServerHostname},
				Addresses: []string{serviceEntryIp.String()},
				Ports: []*networkingv1alpha3spec.Port{{
					Number:   HelloServerPort,
					Protocol: "TCP",
					Name:     "http",
				}},
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
				Endpoints: []*networkingv1alpha3spec.WorkloadEntry{{
					Address: DockerHostAddress,
					Ports:   map[string]uint32{"http": HelloServerPort},
				}},
			},
		}

		// create a virtual service to route to our local service entry
		virtualService := &istionetworkingv1alpha3.VirtualService{
			ObjectMeta: resourceMeta,
			Spec: networkingv1alpha3spec.VirtualService{
				Hosts:    []string{HelloServerHostname},
				Gateways: []string{"mesh"},
				Http: []*networkingv1alpha3spec.HTTPRoute{{
					Name: "hello-server-route",
					Match: []*networkingv1alpha3spec.HTTPMatchRequest{{
						Uri: &networkingv1alpha3spec.StringMatch{MatchType: &networkingv1alpha3spec.StringMatch_Prefix{Prefix: "/"}},
					}},
					Route: []*networkingv1alpha3spec.HTTPRouteDestination{{
						Destination: &networkingv1alpha3spec.Destination{
							Host: HelloServerHostname,
						},
					}},
				}},
			},
		}

		outputs.AddServiceEntries(serviceEntry)
		outputs.AddVirtualServices(virtualService)

		return outputs, nil
	}
}
