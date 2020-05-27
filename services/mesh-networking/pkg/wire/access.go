package wire

import (
	"github.com/google/wire"
	access_control_enforcer2 "github.com/solo-io/service-mesh-hub/pkg/access-control/enforcer"
	"github.com/solo-io/service-mesh-hub/pkg/access-control/enforcer/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/access-control/enforcer/istio"
	"github.com/solo-io/service-mesh-hub/pkg/api/istio/security/v1beta1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	appmesh2 "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer"
	access_control_policy "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator/istio-translator"
)

var (
	AccessControlPolicySet = wire.NewSet(
		LocalAccessControlPolicyEventWatcherProvider,
		v1beta1.AuthorizationPolicyClientFactoryProvider,
		access_control_policy.NewAcpTranslatorLoop,
		istio_translator.NewIstioTranslator,
		AccessControlPolicyMeshTranslatorsProvider,
		zephyr_networking.NewAccessControlPolicyClient,
		// Global AccessControlPolicy enforcer
		istio.NewIstioEnforcer,
		appmesh.NewAppmeshEnforcer,
		appmesh2.NewAppmeshTranslator,
		appmesh.NewAppmeshAccessControlDao,
		appmesh2.AppmeshClientFactoryProvider,
		appmesh2.AppmeshRawClientFactoryProvider,
		appmesh2.NewAppmeshMatcher,
		access_control_enforcer.NewEnforcerLoop,
		GlobalAccessControlPolicyMeshEnforcersProvider,
	)
)

func LocalAccessControlPolicyEventWatcherProvider(mgr mc_manager.AsyncManager) zephyr_networking_controller.AccessControlPolicyEventWatcher {
	return zephyr_networking_controller.NewAccessControlPolicyEventWatcher("management-plane-access-control-event-watcher", mgr.Manager())
}

func AccessControlPolicyMeshTranslatorsProvider(
	istioTranslator istio_translator.IstioTranslator,
) []access_control_policy.AcpMeshTranslator {
	return []access_control_policy.AcpMeshTranslator{
		istioTranslator,
	}
}

func GlobalAccessControlPolicyMeshEnforcersProvider(
	istioEnforcer istio.IstioEnforcer,
	appmeshEnforcer appmesh.AppmeshEnforcer,
) []access_control_enforcer2.AccessPolicyMeshEnforcer {
	return []access_control_enforcer2.AccessPolicyMeshEnforcer{
		istioEnforcer,
		appmeshEnforcer,
	}
}
