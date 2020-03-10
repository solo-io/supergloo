package wire

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
)

type CsrAgentContext struct {
	Ctx                           context.Context
	Manager                       mc_manager.AsyncManager
	CsrController                 controller.MeshGroupCertificateSigningRequestController
	CsrClient                     zephyr_security.MeshGroupCSRClient
	MeshGroupCSRDataSourceFactory csr_generator.MeshGroupCSRDataSourceFactory
	CsrAgentIstioProcessor        csr_generator.MeshGroupCSRProcessor
}

func CsrAgentContextProvider(
	ctx context.Context,
	mgr mc_manager.AsyncManager,
	csrController controller.MeshGroupCertificateSigningRequestController,
	meshGroupCSRDataSourceFactory csr_generator.MeshGroupCSRDataSourceFactory,
	csrAgentIstioProcessor csr_generator.MeshGroupCSRProcessor,
	csrClient zephyr_security.MeshGroupCSRClient,
) CsrAgentContext {
	return CsrAgentContext{
		Ctx:                           ctx,
		Manager:                       mgr,
		CsrController:                 csrController,
		MeshGroupCSRDataSourceFactory: meshGroupCSRDataSourceFactory,
		CsrAgentIstioProcessor:        csrAgentIstioProcessor,
		CsrClient:                     csrClient,
	}
}
