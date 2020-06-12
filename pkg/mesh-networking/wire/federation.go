package wire

import (
	"github.com/google/wire"
	istio_networking_providers "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/providers"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/decider"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/dns"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/resolver"
	istio_federation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/resolver/meshes/istio"
	strategies2 "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/strategies"
)

var (
	FederationProviderSet = wire.NewSet(
		// Decider
		strategies2.NewFederationStrategyChooser,
		strategies2.NewPermissiveFederation,
		decider.NewFederationDecider,
		decider.NewFederationSnapshotListener,

		// Resolver
		istio_networking_providers.GatewayClientFactoryProvider,
		istio_networking_providers.ServiceEntryClientFactoryProvider,
		istio_networking_providers.EnvoyFilterClientFactoryProvider,
		dns.NewIpAssigner,
		dns.NewExternalAccessPointGetter,
		istio_federation.NewIstioFederationClient,
		resolver.NewPerMeshFederationClients,
		resolver.NewFederationResolver,
	)
)
