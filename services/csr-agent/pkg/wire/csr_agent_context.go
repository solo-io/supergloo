package wire

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
)

type CsrAgentContext struct {
	Ctx           context.Context
	Manager       mc_manager.AsyncManager
	CsrHandler    controller.MeshGroupCertificateSigningRequestEventHandler
	CsrController controller.MeshGroupCertificateSigningRequestController
}

func CsrAgentContextProvider(
	ctx context.Context,
	mgr mc_manager.AsyncManager,
	csrHandler controller.MeshGroupCertificateSigningRequestEventHandler,
	csrController controller.MeshGroupCertificateSigningRequestController,
) CsrAgentContext {
	return CsrAgentContext{
		Ctx:           ctx,
		Manager:       mgr,
		CsrHandler:    csrHandler,
		CsrController: csrController,
	}
}
