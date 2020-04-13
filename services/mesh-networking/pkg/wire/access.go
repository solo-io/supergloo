package wire

import (
	"github.com/google/wire"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients/istio/security"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer"
	istio_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer/istio-enforcer"
	access_control_policy "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator/istio-translator"
)

var (
	AccessControlPolicySet = wire.NewSet(
		LocalAccessControlPolicyEventWatcherProvider,
		security.AuthorizationPolicyClientFactoryProvider,
		access_control_policy.NewAcpTranslatorLoop,
		istio_translator.NewIstioTranslator,
		AccessControlPolicyMeshTranslatorsProvider,
		zephyr_networking.NewAccessControlPolicyClient,
		// Global AccessControlPolicy enforcer
		istio_enforcer.NewIstioEnforcer,
		access_control_enforcer.NewEnforcerLoop,
		GlobalAccessControlPolicyMeshEnforcersProvider,
	)
)

func LocalAccessControlPolicyEventWatcherProvider(mgr mc_manager.AsyncManager) controller.AccessControlPolicyEventWatcher {
	return controller.NewAccessControlPolicyEventWatcher("management-plane-access-control-event-watcher", mgr.Manager())
}

func AccessControlPolicyMeshTranslatorsProvider(
	istioTranslator istio_translator.IstioTranslator,
) []access_control_policy.AcpMeshTranslator {
	return []access_control_policy.AcpMeshTranslator{
		istioTranslator,
	}
}

func GlobalAccessControlPolicyMeshEnforcersProvider(
	istioEnforcer istio_enforcer.IstioEnforcer,
) []access_control_enforcer.AccessPolicyMeshEnforcer {
	return []access_control_enforcer.AccessPolicyMeshEnforcer{
		istioEnforcer,
	}
}
