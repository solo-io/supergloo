package wire

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
)

type CsrAgentContext struct {
	Ctx                             context.Context
	Manager                         mc_manager.AsyncManager
	CsrController                   controller.VirtualMeshCertificateSigningRequestController
	CsrClient                       zephyr_security.VirtualMeshCSRClient
	VirtualMeshCSRDataSourceFactory csr_generator.VirtualMeshCSRDataSourceFactory
	CsrAgentIstioProcessor          csr_generator.VirtualMeshCSRProcessor
}

func CsrAgentContextProvider(
	ctx context.Context,
	mgr mc_manager.AsyncManager,
	csrController controller.VirtualMeshCertificateSigningRequestController,
	virtualMeshCSRDataSourceFactory csr_generator.VirtualMeshCSRDataSourceFactory,
	csrAgentIstioProcessor csr_generator.VirtualMeshCSRProcessor,
	csrClient zephyr_security.VirtualMeshCSRClient,
) CsrAgentContext {
	return CsrAgentContext{
		Ctx:                             ctx,
		Manager:                         mgr,
		CsrController:                   csrController,
		VirtualMeshCSRDataSourceFactory: virtualMeshCSRDataSourceFactory,
		CsrAgentIstioProcessor:          csrAgentIstioProcessor,
		CsrClient:                       csrClient,
	}
}
