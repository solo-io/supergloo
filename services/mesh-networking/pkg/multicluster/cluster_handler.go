package networking_multicluster

import (
	"context"

	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	cert_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/security/certgen"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	csr_generator "github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/csr-generator"
	cert_signer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-signer"
)

// this is the main entrypoint for all virtual-mesh multi cluster logic
func NewMeshNetworkingClusterHandler(
	localManager mc_manager.AsyncManager,
	controllerFactories *ControllerFactories,
	clientFactories *ClientFactories,
	virtualMeshCertClient cert_signer.VirtualMeshCertClient,
	signer certgen.Signer,
	csrDataSourceFactory csr_generator.VirtualMeshCSRDataSourceFactory,
) (mc_manager.AsyncManagerHandler, error) {

	handler := &meshNetworkingClusterHandler{
		controllerFactories:   controllerFactories,
		clientFactories:       clientFactories,
		virtualMeshCertClient: virtualMeshCertClient,
		signer:                signer,
		csrDataSourceFactory:  csrDataSourceFactory,
	}

	return handler, nil
}

type meshNetworkingClusterHandler struct {
	controllerFactories   *ControllerFactories
	clientFactories       *ClientFactories
	virtualMeshCertClient cert_signer.VirtualMeshCertClient
	signer                certgen.Signer
	csrDataSourceFactory  csr_generator.VirtualMeshCSRDataSourceFactory
}

type clusterDependentDeps struct {
	csrController cert_controller.VirtualMeshCertificateSigningRequestController
	csrClient     zephyr_security.VirtualMeshCertificateSigningRequestClient
}

func (m *meshNetworkingClusterHandler) ClusterAdded(ctx context.Context, mgr mc_manager.AsyncManager, clusterName string) error {
	clusterDeps, err := m.initializeClusterScopedDeps(mgr, clusterName)
	if err != nil {
		return err
	}

	certSigner := cert_signer.NewVirtualMeshCSRSigner(m.virtualMeshCertClient, clusterDeps.csrClient, m.signer)
	mgcsrHandler := m.csrDataSourceFactory(ctx, clusterDeps.csrClient, cert_signer.NewVirtualMeshCSRSigningProcessor(certSigner))
	if err = clusterDeps.csrController.AddEventHandler(ctx, mgcsrHandler); err != nil {
		return err
	}

	return nil
}

func (m *meshNetworkingClusterHandler) ClusterRemoved(clusterName string) error {
	return nil
}

func (m *meshNetworkingClusterHandler) initializeClusterScopedDeps(
	mgr mc_manager.AsyncManager,
	clusterName string,
) (*clusterDependentDeps, error) {
	csrController, err := m.controllerFactories.VirtualMeshCSRControllerFactory(mgr, clusterName)
	if err != nil {
		return nil, err
	}

	csrClient := m.clientFactories.VirtualMeshCSRClientFactory(mgr.Manager().GetClient())

	return &clusterDependentDeps{
		csrController: csrController,
		csrClient:     csrClient,
	}, nil
}
