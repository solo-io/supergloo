package setup

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/pkg/config/gloo"

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

type EnabledConfigLoops struct {
	Istio bool
	Gloo  bool
}

func RunConfigEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error), enabled EnabledConfigLoops) error {
	ctx = contextutils.WithLogger(ctx, "config-event-loop")
	logger := contextutils.LoggerFrom(ctx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("config error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	configSyncers, err := createConfigSyncers(ctx, cs, enabled)
	if err != nil {
		return err
	}

	if err := runConfigEventLoop(ctx, cs, errHandler, configSyncers); err != nil {
		return err
	}

	return nil
}

// Add config syncers here
func createConfigSyncers(ctx context.Context, cs *clientset.Clientset, enabled EnabledConfigLoops) (v1.ConfigSyncers, error) {
	var syncers v1.ConfigSyncers

	if enabled.Istio {
		istioSyncer, err := createIstioConfigSyncer(ctx, cs)
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, istioSyncer)
	}

	if enabled.Gloo {
		glooSyncer, err := createGlooConfigSyncer(ctx, cs)
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, glooSyncer)

	}

	return syncers, nil
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

func createGlooConfigSyncer(ctx context.Context, cs *clientset.Clientset) (v1.ConfigSyncer, error) {
	newReporter := reporter.NewReporter("gloo-config-reporter",
		cs.Input.Mesh.BaseClient(),
		cs.Input.Upstream.BaseClient(),
		cs.Input.Upstream.BaseClient())

	return gloo.NewGlooConfigSyncer(newReporter, cs), nil
}

// start the istio config event loop
func runConfigEventLoop(ctx context.Context, clientset *clientset.Clientset, errHandler func(err error), syncers v1.ConfigSyncers) error {
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
