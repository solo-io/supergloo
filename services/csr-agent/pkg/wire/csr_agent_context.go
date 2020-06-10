package wire

import (
	"context"

	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	smh_security_controller "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	csr_generator "github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/csr-generator"
)

type CsrAgentContext struct {
	Ctx                             context.Context
	Manager                         mc_manager.AsyncManager
	CsrEventWatcher                 smh_security_controller.VirtualMeshCertificateSigningRequestEventWatcher
	CsrClient                       smh_security.VirtualMeshCertificateSigningRequestClient
	VirtualMeshCSRDataSourceFactory csr_generator.VirtualMeshCSRDataSourceFactory
	CsrAgentIstioProcessor          csr_generator.VirtualMeshCSRProcessor
}

func CsrAgentContextProvider(
	ctx context.Context,
	mgr mc_manager.AsyncManager,
	csrEventWatcher smh_security_controller.VirtualMeshCertificateSigningRequestEventWatcher,
	virtualMeshCSRDataSourceFactory csr_generator.VirtualMeshCSRDataSourceFactory,
	csrAgentIstioProcessor csr_generator.VirtualMeshCSRProcessor,
	csrClient smh_security.VirtualMeshCertificateSigningRequestClient,
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
