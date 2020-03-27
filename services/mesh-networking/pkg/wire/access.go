package wire

import (
	"github.com/google/wire"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/clients/istio/security"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	access_control_policy "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-policy-translator"
	istio_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-policy-translator/istio-translator"
)

var (
	AccessControlPolicySet = wire.NewSet(
		LocalAccessControlPolicyProvider,
		security.AuthorizationPolicyClientFactoryProvider,
		access_control_policy.NewAcpTranslatorLoop,
		istio_translator.NewIstioTranslator,
		AccessControlPolicyMeshTranslatorsProvider,
		zephyr_networking.NewAccessControlPolicyClient,
	)
)

func LocalAccessControlPolicyProvider(mgr mc_manager.AsyncManager) (controller.AccessControlPolicyController, error) {
	return controller.NewAccessControlPolicyController("management-plane-access-control-controller", mgr.Manager())
}

func AccessControlPolicyMeshTranslatorsProvider(
	istioTranslator istio_translator.IstioTranslator,
) []access_control_policy.AcpMeshTranslator {
	return []access_control_policy.AcpMeshTranslator{istioTranslator}
}
