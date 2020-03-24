package wire

import (
	"github.com/google/wire"
	istio_networking "github.com/solo-io/mesh-projects/pkg/clients/istio/networking"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/decider"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/decider/strategies"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/dns"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/resolver"
	istio_federation "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/resolver/meshes/istio"
)

var (
	FederationProviderSet = wire.NewSet(
		// Decider
		strategies.NewFederationStrategyChooser,
		strategies.NewPermissiveFederation,
		decider.NewFederationDecider,
		decider.NewFederationSnapshotListener,

		// Resolver
		istio_networking.NewGatewayClientFactory,
		istio_networking.NewServiceEntryClientFactory,
		istio_networking.NewEnvoyFilterClientFactory,
		dns.NewIpAssigner,
		dns.NewExternalAccessPointGetter,
		istio_federation.NewIstioFederationClient,
		resolver.NewPerMeshFederationClients,
		resolver.NewFederationResolver,
	)
)
