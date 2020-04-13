package wire

import (
	"context"

	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	csr_generator "github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/csr-generator"
)

type CsrAgentContext struct {
	Ctx                             context.Context
	Manager                         mc_manager.AsyncManager
	CsrEventWatcher                 controller.VirtualMeshCertificateSigningRequestEventWatcher
	CsrClient                       zephyr_security.VirtualMeshCertificateSigningRequestClient
	VirtualMeshCSRDataSourceFactory csr_generator.VirtualMeshCSRDataSourceFactory
	CsrAgentIstioProcessor          csr_generator.VirtualMeshCSRProcessor
}

func CsrAgentContextProvider(
	ctx context.Context,
	mgr mc_manager.AsyncManager,
	csrEventWatcher controller.VirtualMeshCertificateSigningRequestEventWatcher,
	virtualMeshCSRDataSourceFactory csr_generator.VirtualMeshCSRDataSourceFactory,
	csrAgentIstioProcessor csr_generator.VirtualMeshCSRProcessor,
	csrClient zephyr_security.VirtualMeshCertificateSigningRequestClient,
) CsrAgentContext {
	return CsrAgentContext{
		Ctx:                             ctx,
		Manager:                         mgr,
		CsrEventWatcher:                 csrEventWatcher,
		VirtualMeshCSRDataSourceFactory: virtualMeshCSRDataSourceFactory,
		CsrAgentIstioProcessor:          csrAgentIstioProcessor,
		CsrClient:                       csrClient,
	}
}
