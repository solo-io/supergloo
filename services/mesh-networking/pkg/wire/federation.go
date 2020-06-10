package wire

import (
	"github.com/google/wire"
	istio_networking "github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/dns"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/resolver"
	istio_federation "github.com/solo-io/service-mesh-hub/pkg/common/federation/resolver/meshes/istio"
	strategies2 "github.com/solo-io/service-mesh-hub/pkg/common/federation/strategies"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/decider"
)

var (
	FederationProviderSet = wire.NewSet(
		// Decider
		strategies2.NewFederationStrategyChooser,
		strategies2.NewPermissiveFederation,
		decider.NewFederationDecider,
		decider.NewFederationSnapshotListener,

		// Resolver
		istio_networking.GatewayClientFactoryProvider,
		istio_networking.ServiceEntryClientFactoryProvider,
		istio_networking.EnvoyFilterClientFactoryProvider,
		dns.NewIpAssigner,
		dns.NewExternalAccessPointGetter,
		istio_federation.NewIstioFederationClient,
		resolver.NewPerMeshFederationClients,
		resolver.NewFederationResolver,
	)
)
