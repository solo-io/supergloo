package wire

import (
	"github.com/google/wire"
	v1beta1 "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/providers"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/cloud"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/matcher"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/translation"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-enforcer"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-enforcer/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-enforcer/istio"
	access_control_policy "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-policy-translator/istio-translator"
)

var (
	AccessControlPolicySet = wire.NewSet(
		LocalAccessControlPolicyEventWatcherProvider,
		v1beta1.AuthorizationPolicyClientFactoryProvider,
		access_control_policy.NewAcpTranslatorLoop,
		istio_translator.NewIstioTranslator,
		AccessControlPolicyMeshTranslatorsProvider,
		smh_networking.NewAccessControlPolicyClient,
		// Global AccessControlPolicy enforcer
		istio.NewIstioEnforcer,
		appmesh.NewAppmeshEnforcer,
		translation.NewAppmeshTranslator,
		translation.NewAppmeshAccessControlDao,
		matcher.NewAppmeshMatcher,
		clients.AppmeshClientFactoryProvider,
		cloud.NewAwsCloudStore,
		access_control_enforcer.NewEnforcerLoop,
		GlobalAccessControlPolicyMeshEnforcersProvider,
		translation.NewAppmeshTranslationReconciler,
	)
)

func LocalAccessControlPolicyEventWatcherProvider(mgr mc_manager.AsyncManager) smh_networking_controller.AccessControlPolicyEventWatcher {
	return smh_networking_controller.NewAccessControlPolicyEventWatcher("management-plane-access-control-event-watcher", mgr.Manager())
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
) []access_control_enforcer.AccessPolicyMeshEnforcer {
	return []access_control_enforcer.AccessPolicyMeshEnforcer{
		istioEnforcer,
		appmeshEnforcer,
	}
}
