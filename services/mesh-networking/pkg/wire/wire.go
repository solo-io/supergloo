// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/uninstall/config_lookup"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/apps"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking"
	"github.com/solo-io/service-mesh-hub/pkg/kubeconfig"
	"github.com/solo-io/service-mesh-hub/pkg/security/certgen"
	multicluster_wire "github.com/solo-io/service-mesh-hub/services/common/multicluster/wire"
	csr_generator "github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/csr-generator"
	networking_multicluster "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster"
	controller_factories "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/multicluster/controllers"
	cert_manager "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-manager"
	cert_signer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-signer"
	vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation"
)

func InitializeMeshNetworking(ctx context.Context) (MeshNetworkingContext, error) {
	wire.Build(
		kubernetes_core.NewSecretClient,
		kubernetes_core.NewConfigMapClient,
		kubernetes_core.NewPodClientFactory,
		kubernetes_core.NewNodeClientFactory,
		discovery_core.NewMeshWorkloadClient,
		discovery_core.NewKubernetesClusterClient,
		kubernetes_apps.DeploymentClientFactoryForConfigProvider,
		kubeconfig.SecretToConfigConverterProvider,
		config_lookup.NewKubeConfigLookup,
		zephyr_networking.NewVirtualMeshClient,
		discovery_core.NewMeshClient,
		csr_generator.NewVirtualMeshCSRDataSourceFactory,
		vm_validation.NewVirtualMeshFinder,
		cert_signer.NewVirtualMeshCertClient,
		multicluster_wire.MulticlusterProviderSet,
		multicluster_wire.DynamicClientGetterProvider,
		certgen.NewSigner,
		certgen.NewRootCertGenerator,
		LocalMeshWorkloadControllerProvider,
		LocalMeshServiceControllerProvider,
		ClientFactoryProviderSet,
		ControllerFactoryProviderSet,
		TrafficPolicyProviderSet,
		AccessControlPolicySet,
		FederationProviderSet,
		networking_multicluster.NewMeshNetworkingClusterHandler,
		controller_factories.NewLocalVirtualMeshController,
		vm_validation.NewVirtualMeshValidator,
		cert_manager.VMCSRSnapshotListenerSet,
		MeshNetworkingSnapshotContextProvider,
		MeshNetworkingContextProvider,
	)

	return MeshNetworkingContext{}, nil
}
