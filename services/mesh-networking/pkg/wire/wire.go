//+build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	multicluster_wire "github.com/solo-io/mesh-projects/services/common/multicluster/wire"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/controllers"
	networking_multicluster "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster"
	controller_factories "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/controllers"
	cert_manager "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-manager"
	cert_signer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-signer"
	group_validation "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/validation"
)

func InitializeMeshNetworking(ctx context.Context) (MeshNetworkingContext, error) {
	wire.Build(
		kubernetes_core.NewSecretsClient,
		zephyr_networking.NewMeshGroupClient,
		discovery_core.NewMeshClient,
		csr_generator.NewMeshGroupCSRDataSourceFactory,
		group_validation.NewGroupMeshFinder,
		cert_signer.NewMeshGroupCertClient,
		multicluster_wire.MulticlusterProviderSet,
		multicluster_wire.DynamicClientGetterProvider,
		certgen.NewSigner,
		ClientFactoryProviderSet,
		ControllerFactoryProviderSet,
		TrafficPolicyProviderSet,
		networking_multicluster.NewMeshNetworkingClusterHandler,
		controllers.NewMeshServiceControllerFactory,
		controllers.NewMeshWorkloadControllerFactory,
		controller_factories.NewMeshGroupControllerFactory,
		group_validation.NewMeshGroupValidator,
		cert_manager.GroupMgcsrSnapshotListenerSet,
		MeshNetworkingSnapshotContextProvider,
		MeshNetworkingContextProvider,
	)

	return MeshNetworkingContext{}, nil
}
