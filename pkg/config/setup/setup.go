package setup

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	policyv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/config/appmesh"
	"github.com/solo-io/supergloo/pkg/config/istio"
	"github.com/solo-io/supergloo/pkg/config/linkerd"
	"github.com/solo-io/supergloo/pkg/registration"
	appmeshtranslator "github.com/solo-io/supergloo/pkg/translator/appmesh"
	istiotranslator "github.com/solo-io/supergloo/pkg/translator/istio"
	istioplugins "github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	linkerdtranslator "github.com/solo-io/supergloo/pkg/translator/linkerd"
	linkerdplugins "github.com/solo-io/supergloo/pkg/translator/linkerd/plugins"
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

		if enabled.Linkerd {
			linkerdSyncer, err := createLinkerdConfigSyncer(ctx, cs)
			if err != nil {
				return nil, err
			}
			syncers = append(syncers, linkerdSyncer)
		}

		if enabled.AppMesh {
			appMeshSyncer, err := createAppmeshConfigSyncer(cs)
			if err != nil {
				return nil, err
			}
			syncers = append(syncers, appMeshSyncer)
		}

		ctx = contextutils.WithLogger(ctx, "config-event-loop")

		configEmitter := v1.NewConfigEmitter(
			cs.Supergloo.Mesh,
			cs.Supergloo.MeshIngress,
			cs.Supergloo.MeshGroup,
			cs.Supergloo.RoutingRule,
			cs.Supergloo.SecurityRule,
			cs.Supergloo.TlsSecret,
			cs.Supergloo.Upstream,
			cs.Discovery.Pod,
		)
		configEventLoop := v1.NewConfigEventLoop(configEmitter, syncers)

		return configEventLoop, nil
	}
}

func createAppmeshConfigSyncer(cs *clientset.Clientset) (v1.ConfigSyncer, error) {
	translator := appmeshtranslator.NewAppMeshTranslator()
	reconciler := appmesh.NewReconciler(appmesh.NewAppMeshClientBuilder(cs.Supergloo.Secret))
	newReporter := makeReporter("appmesh-config-reporter", cs.Supergloo)

	return appmesh.NewAppMeshConfigSyncer(translator, reconciler, newReporter)
}

func createIstioConfigSyncer(ctx context.Context, cs *clientset.Clientset) (v1.ConfigSyncer, error) {
	istioClients, err := clientset.IstioFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing istio clients")
	}

	translator := istiotranslator.NewTranslator(istioplugins.Plugins(cs.Kube))

	reconcilers := istio.NewIstioReconcilers(map[string]string{"created_by": "istio-config-syncer"},
		rbacv1alpha1.NewRbacConfigReconciler(istioClients.RbacConfig),
		rbacv1alpha1.NewServiceRoleReconciler(istioClients.ServiceRole),
		rbacv1alpha1.NewServiceRoleBindingReconciler(istioClients.ServiceRoleBinding),
		policyv1alpha1.NewMeshPolicyReconciler(istioClients.MeshPolicy),
		v1alpha3.NewDestinationRuleReconciler(istioClients.DestinationRule),
		v1alpha3.NewVirtualServiceReconciler(istioClients.VirtualService),
		v1.NewTlsSecretReconciler(cs.Supergloo.TlsSecret),
	)

	newReporter := makeReporter("istio-config-reporter", cs.Supergloo)

	return istio.NewIstioConfigSyncer(translator, reconcilers, newReporter), nil
}

func createLinkerdConfigSyncer(ctx context.Context, cs *clientset.Clientset) (v1.ConfigSyncer, error) {
	clients, err := clientset.LinkerdFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing linkerd clients")
	}

	translator := linkerdtranslator.NewTranslator(linkerdplugins.Plugins(cs.Kube))

	reconcilers := linkerd.NewLinkerdReconcilers(map[string]string{"created_by": "linkerd-config-syncer"},
		v1.NewServiceProfileReconciler(clients.ServiceProfile),
	)

	newReporter := makeReporter("linkerd-config-reporter", cs.Supergloo)

	return linkerd.NewLinkerdConfigSyncer(translator, reconcilers, newReporter), nil
}

func makeReporter(name string, cs *clientset.SuperglooClients) reporter.Reporter {
	return reporter.NewReporter(name,
		cs.Mesh.BaseClient(),
		cs.Upstream.BaseClient(),
		cs.RoutingRule.BaseClient(),
		cs.SecurityRule.BaseClient())
}
