package decider

import (
	"context"

	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
)

/***************************************
*
* A Federation Decider updates Mesh Services with the decision we've made about where they're being federated to.
* It does this by deciding what FederationStrategy it's going to use, based on the Mesh Group's federation config
* The actual service update is done by the FederationStrategy; the group status is updated here.
*
****************************************/
type FederationDecider interface {
	DecideFederation(ctx context.Context, networkingSnapshot snapshot.MeshNetworkingSnapshot)
}
