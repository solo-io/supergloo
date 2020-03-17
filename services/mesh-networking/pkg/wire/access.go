package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	access_control_poilcy "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-poilcy"
)

var (
	AccessControlPolicySet = wire.NewSet(
		LocalAccessControlPolicyProvider,
		security.NewAuthorizationPolicyClient,
		access_control_poilcy.NewAccessControlPolicyTranslator,
	)
)

func LocalAccessControlPolicyProvider(mgr mc_manager.AsyncManager) (controller.AccessControlPolicyController, error) {
	return controller.NewAccessControlPolicyController("management-plane-access-control-controller", mgr.Manager())
}
