package meshes

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/federation/dns"
)

//go:generate mockgen -source ./interfaces.go -destination ./mock/mock_interfaces.go

// per-mesh implementation of setting up federation resources
type MeshFederationClient interface {
	// set up appropriate resources in the cluster of the mesh service (the target cluster)
	// returns the externally-resolvable address (NOTE: can be either an IP or a hostname) that the service is now reachable at
	FederateServiceSide(
		ctx context.Context,
		installationNamespace string,
		virtualMesh *zephyr_networking.VirtualMesh,
		meshService *zephyr_discovery.MeshService,
	) (eap dns.ExternalAccessPoint, err error)

	// set up appropriate resources in the cluster of the mesh workload (the client cluster) where the traffic will originate
	FederateClientSide(
		ctx context.Context,
		installationNamespace string,
		eap dns.ExternalAccessPoint,
		meshService *zephyr_discovery.MeshService,
		meshWorkload *zephyr_discovery.MeshWorkload,
	) error
}
