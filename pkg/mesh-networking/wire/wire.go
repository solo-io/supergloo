// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	kubernetes_apps "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	multicluster_wire "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/wire"
	csr_generator "github.com/solo-io/service-mesh-hub/pkg/common/csr-generator"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen"
	"github.com/solo-io/service-mesh-hub/pkg/common/filesystem/files"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	networking_multicluster "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/compute-target"
	controller_factories "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/compute-target/controllers"
	cert_manager "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/security/cert-manager"
	cert_signer "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/security/cert-signer"
	vm_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MeshServiceReaderProvider(client client.Client) smh_discovery.MeshServiceReader {
	return smh_discovery.MeshServiceClientProvider(client)
}
func MeshWorkloadReaderProvider(client client.Client) smh_discovery.MeshWorkloadReader {
	return smh_discovery.MeshWorkloadClientProvider(client)
}

func InitializeMeshNetworking(ctx context.Context) (MeshNetworkingContext, error) {
	wire.Build(
		kubernetes_core.SecretClientProvider,
		kubernetes_core.ConfigMapClientProvider,
		kubernetes_core.PodClientFactoryProvider,
		kubernetes_core.NodeClientFactoryProvider,
		kubernetes_apps.DeploymentClientFactoryProvider,
		smh_discovery.MeshClientProvider,
		MeshServiceReaderProvider,
		smh_discovery.MeshServiceClientProvider,
		smh_discovery.MeshWorkloadClientProvider,
		MeshWorkloadReaderProvider,
		smh_networking.VirtualMeshClientProvider,
		smh_networking.TrafficPolicyClientProvider,
		smh_networking.AccessControlPolicyClientProvider,
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
		ComputeTargetCredentialsHandlersProvider,
		kubeconfig.NewConverter,
		files.NewDefaultFileReader,
	)

	return MeshNetworkingContext{}, nil
}
