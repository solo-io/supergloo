package setup

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/supergloo/pkg/config/appmesh"
	"github.com/solo-io/supergloo/pkg/registration"
	appmeshtranslator "github.com/solo-io/supergloo/pkg/translator/appmesh"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	policyv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/config/istio"
	istiotranslator "github.com/solo-io/supergloo/pkg/translator/istio"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
)

func NewSuperglooConfigLoopStarter(clientset *clientset.Clientset) registration.ConfigLoopStarters {
	return registration.ConfigLoopStarters{createConfigStarters(clientset)}
}

// Add config syncers here
func createConfigStarters(cs *clientset.Clientset) registration.ConfigLoopStarter {

	return func(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
		var syncers v1.ConfigSyncers

		if enabled.Istio {
			istioSyncer, err := createIstioConfigSyncer(ctx, cs)
			if err != nil {
				return nil, err
			}
			syncers = append(syncers, istioSyncer)
		}

		if enabled.AppMesh {
			appMeshSyncer, err := createAppmeshConfigSyncer(ctx, cs)
			if err != nil {
				return nil, err
			}
			syncers = append(syncers, appMeshSyncer)
		}

		ctx = contextutils.WithLogger(ctx, "config-event-loop")

		configEmitter := v1.NewConfigEmitter(
			cs.Input.Mesh,
			cs.Input.MeshIngress,
			cs.Input.MeshGroup,
			cs.Input.RoutingRule,
			cs.Input.SecurityRule,
			cs.Input.TlsSecret,
			cs.Input.Upstream,
			cs.Discovery.Pod,
		)
		configEventLoop := v1.NewConfigEventLoop(configEmitter, syncers)

		return configEventLoop, nil
	}

}

func createAppmeshConfigSyncer(ctx context.Context, cs *clientset.Clientset) (v1.ConfigSyncer, error) {
	translator := appmeshtranslator.NewAppMeshTranslator()

	newReporter := reporter.NewReporter("appmesh-config-reporter",
		cs.Input.Mesh.BaseClient(),
		cs.Input.Upstream.BaseClient(),
		cs.Input.RoutingRule.BaseClient(),
		cs.Input.SecurityRule.BaseClient())

	return appmesh.NewAppMeshConfigSyncer(translator, newReporter)
}

func createIstioConfigSyncer(ctx context.Context, cs *clientset.Clientset) (v1.ConfigSyncer, error) {
	istioClients, err := clientset.IstioFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing istio clients")
	}

	translator := istiotranslator.NewTranslator(plugins.Plugins(cs.Kube))

	reconcilers := istio.NewIstioReconcilers(map[string]string{"owner_ref": "istio-config-syncer"},
		rbacv1alpha1.NewRbacConfigReconciler(istioClients.RbacConfig),
		rbacv1alpha1.NewServiceRoleReconciler(istioClients.ServiceRole),
		rbacv1alpha1.NewServiceRoleBindingReconciler(istioClients.ServiceRoleBinding),
		policyv1alpha1.NewMeshPolicyReconciler(istioClients.MeshPolicy),
		v1alpha3.NewDestinationRuleReconciler(istioClients.DestinationRule),
		v1alpha3.NewVirtualServiceReconciler(istioClients.VirtualService),
		v1.NewTlsSecretReconciler(cs.Input.TlsSecret),
	)

	newReporter := reporter.NewReporter("istio-config-reporter",
		cs.Input.Mesh.BaseClient(),
		cs.Input.Upstream.BaseClient(),
		cs.Input.RoutingRule.BaseClient(),
		cs.Input.SecurityRule.BaseClient())

	return istio.NewIstioConfigSyncer(translator, reconcilers, newReporter), nil
}
