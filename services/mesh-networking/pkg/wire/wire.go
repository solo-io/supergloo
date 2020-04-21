// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
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
		kubernetes_core.SecretClientProvider,
		kubernetes_core.ConfigMapClientProvider,
		kubernetes_core.PodClientFactoryProvider,
		kubernetes_core.NodeClientFactoryProvider,
		kubernetes_apps.DeploymentClientFactoryProvider,
		zephyr_discovery.MeshClientProvider,
		zephyr_discovery.MeshServiceClientProvider,
		zephyr_discovery.MeshWorkloadClientProvider,
		zephyr_networking.VirtualMeshClientProvider,
		zephyr_networking.TrafficPolicyClientProvider,
		zephyr_networking.AccessControlPolicyClientProvider,
		csr_generator.NewVirtualMeshCSRDataSourceFactory,
		vm_validation.NewVirtualMeshFinder,
		cert_signer.NewVirtualMeshCertClient,
		multicluster_wire.MulticlusterProviderSet,
		multicluster_wire.DynamicClientGetterProvider,
		certgen.NewSigner,
		certgen.NewRootCertGenerator,
		LocalMeshWorkloadEventWatcherProvider,
		LocalMeshServiceEventWatcherProvider,
		ClientFactoryProviderSet,
		ControllerFactoryProviderSet,
		TrafficPolicyProviderSet,
		AccessControlPolicySet,
		FederationProviderSet,
		networking_multicluster.NewMeshNetworkingClusterHandler,
		controller_factories.NewLocalVirtualMeshEventWatcher,
		vm_validation.NewVirtualMeshValidator,
		cert_manager.VMCSRSnapshotListenerSet,
		MeshNetworkingSnapshotContextProvider,
		MeshNetworkingContextProvider,
		AwsSet,
	)

	return MeshNetworkingContext{}, nil
}
