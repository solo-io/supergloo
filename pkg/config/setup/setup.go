package setup

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/pkg/config/appmesh"
	"github.com/solo-io/supergloo/pkg/registration"
	appmeshtranslator "github.com/solo-io/supergloo/pkg/translator/appmesh"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
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

type SuperglooConfigEventLoopRunner struct {
	Clientset  *clientset.Clientset
	ErrHandler func(error)
}

func NewSuperglooCongigLoopStarter(clientset *clientset.Clientset, errHandler func(error)) *SuperglooConfigEventLoopRunner {
	return &SuperglooConfigEventLoopRunner{Clientset: clientset, ErrHandler: errHandler}
}

func (s *SuperglooConfigEventLoopRunner) Run(ctx context.Context, enabled registration.EnabledConfigLoops) error {
	ctx = contextutils.WithLogger(ctx, "config-event-loop")
	logger := contextutils.LoggerFrom(ctx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("config error: %v", err)
		if s.ErrHandler != nil {
			s.ErrHandler(err)
		}
	}

	configSyncers, err := createConfigSyncers(ctx, s.Clientset, enabled)
	if err != nil {
		return err
	}

	if err := runConfigEventLoop(ctx, s.Clientset, errHandler, configSyncers); err != nil {
		return err
	}

	return nil
}

// Add config syncers here
func createConfigSyncers(ctx context.Context, cs *clientset.Clientset, enabled registration.EnabledConfigLoops) (v1.ConfigSyncer, error) {
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

	return syncers, nil
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

// start the istio config event loop
func runConfigEventLoop(ctx context.Context, clientset *clientset.Clientset, errHandler func(err error), syncers v1.ConfigSyncer) error {
	configEmitter := v1.NewConfigEmitter(
		clientset.Input.Mesh,
		clientset.Input.MeshIngress,
		clientset.Input.MeshGroup,
		clientset.Input.RoutingRule,
		clientset.Input.SecurityRule,
		clientset.Input.TlsSecret,
		clientset.Input.Upstream,
		clientset.Discovery.Pod,
	)
	configEventLoop := v1.NewConfigEventLoop(configEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	configEventLoopErrs, err := configEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-configEventLoopErrs:
				errHandler(err)
			case <-ctx.Done():
			}
		}
	}()
	return nil
}
