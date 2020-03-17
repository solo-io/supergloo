package meshes

import (
	"context"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/dns"
)

//go:generate mockgen -source ./interfaces.go -destination ./mock/mock_interfaces.go

// per-mesh implementation of setting up federation resources
type MeshFederationClient interface {
	// set up appropriate resources in the cluster of the mesh service (the target cluster)
	// returns the externally-resolvable address (NOTE: can be either an IP or a hostname) that the service is now reachable at
	FederateServiceSide(
		ctx context.Context,
		meshGroup *networking_v1alpha1.MeshGroup,
		meshService *discovery_v1alpha1.MeshService,
	) (eap dns.ExternalAccessPoint, err error)

	// set up appropriate resources in the cluster of the mesh workload (the client cluster) where the traffic will originate
	FederateClientSide(
		ctx context.Context,
		eap dns.ExternalAccessPoint,
		meshService *discovery_v1alpha1.MeshService,
		meshWorkload *discovery_v1alpha1.MeshWorkload,
	) error
}
