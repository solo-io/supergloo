package networking_multicluster

import (
	"context"

	cert_controller "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	"github.com/solo-io/mesh-projects/services/common"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
	cert_signer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-signer"
)

// this is the main entrypoint for all mesh-group multi cluster logic
func NewMeshNetworkingClusterHandler(
	localManager mc_manager.AsyncManager,
	controllerFactories *ControllerFactories,
	clientFactories *ClientFactories,
	meshGroupCertClient cert_signer.MeshGroupCertClient,
	signer certgen.Signer,
	csrDataSourceFactory csr_generator.MeshGroupCSRDataSourceFactory,
) (mc_manager.AsyncManagerHandler, error) {

	handler := &meshNetworkingClusterHandler{
		controllerFactories:  controllerFactories,
		clientFactories:      clientFactories,
		meshGroupCertClient:  meshGroupCertClient,
		signer:               signer,
		csrDataSourceFactory: csrDataSourceFactory,
	}

	// be sure that we are also watching our local cluster
	err := handler.ClusterAdded(localManager.Context(), localManager, common.LocalClusterName)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

type meshNetworkingClusterHandler struct {
	controllerFactories  *ControllerFactories
	clientFactories      *ClientFactories
	meshGroupCertClient  cert_signer.MeshGroupCertClient
	signer               certgen.Signer
	csrDataSourceFactory csr_generator.MeshGroupCSRDataSourceFactory
}

type clusterDependentDeps struct {
	csrController cert_controller.MeshGroupCertificateSigningRequestController
	csrClient     zephyr_security.MeshGroupCSRClient
}

func (m *meshNetworkingClusterHandler) ClusterAdded(ctx context.Context, mgr mc_manager.AsyncManager, clusterName string) error {
	clusterDeps, err := m.initializeClusterScopedDeps(mgr, clusterName)
	if err != nil {
		return err
	}

	certSigner := cert_signer.NewMeshGroupCSRSigner(m.meshGroupCertClient, clusterDeps.csrClient, m.signer)
	mgcsrHandler := m.csrDataSourceFactory(ctx, clusterDeps.csrClient, cert_signer.NewMeshGroupCSRSigningProcessor(certSigner))
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
	csrController, err := m.controllerFactories.MeshGroupCSRControllerFactory(mgr, clusterName)
	if err != nil {
		return nil, err
	}

	csrClient := m.clientFactories.MeshGroupCSRClientFactory(mgr.Manager().GetClient())

	return &clusterDependentDeps{
		csrController: csrController,
		csrClient:     csrClient,
	}, nil
}
