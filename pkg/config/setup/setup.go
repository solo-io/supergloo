package setup

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/supergloo/pkg/config/smi"
	smitranslator "github.com/solo-io/supergloo/pkg/translator/smi"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/config/appmesh"
	"github.com/solo-io/supergloo/pkg/config/istio"
	"github.com/solo-io/supergloo/pkg/config/linkerd"
	appmeshtranslator "github.com/solo-io/supergloo/pkg/translator/appmesh"
	istiotranslator "github.com/solo-io/supergloo/pkg/translator/istio"
	istioplugins "github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	linkerdtranslator "github.com/solo-io/supergloo/pkg/translator/linkerd"
	linkerdplugins "github.com/solo-io/supergloo/pkg/translator/linkerd/plugins"
)

func RunConfigEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error)) error {
	ctx = contextutils.WithLogger(ctx, "config-event-loop")

	logger := contextutils.LoggerFrom(ctx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("registration error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	appmeshConfigSyncer := createAppmeshConfigSyncer(cs)
	istioConfigSyncer, err := createIstioConfigSyncer(ctx, cs)
	if err != nil {
		return err
	}
	linkerdConfigSyncer, err := createLinkerdConfigSyncer(ctx, cs)
	if err != nil {
		return err
	}
	smiConfigSyncer, err := createSmiConfigSyncer(ctx, cs)
	if err != nil {
		return err
	}

	return runConfigEventLoop(ctx, errHandler, cs, v1.ConfigSyncers{
		appmeshConfigSyncer,
		istioConfigSyncer,
		linkerdConfigSyncer,
		smiConfigSyncer,
	})
}

func runConfigEventLoop(ctx context.Context, errHandler func(error), cs *clientset.Clientset, syncers v1.ConfigSyncer) error {

	configEmitter := v1.NewConfigEmitter(
		cs.Supergloo.Mesh,
		cs.Supergloo.MeshIngress,
		cs.Supergloo.MeshGroup,
		cs.Supergloo.RoutingRule,
		cs.Supergloo.SecurityRule,
		cs.Supergloo.TlsSecret,
		cs.Supergloo.Upstream,
		cs.Discovery.Pod,
		cs.Discovery.Service,
	)
	configEventLoop := v1.NewConfigEventLoop(configEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}
	eventLoopErrs, err := configEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-eventLoopErrs:
				errHandler(err)
			case <-ctx.Done():
			}
		}
	}()

	return nil
}

func createAppmeshConfigSyncer(cs *clientset.Clientset) v1.ConfigSyncer {
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
		istioClients.RbacConfig,
		istioClients.ServiceRole,
		istioClients.ServiceRoleBinding,
		istioClients.MeshPolicy,
		istioClients.DestinationRule,
		istioClients.VirtualService,
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
		clients.ServiceProfile,
	)

	newReporter := makeReporter("linkerd-config-reporter", cs.Supergloo)

	return linkerd.NewLinkerdConfigSyncer(translator, reconcilers, newReporter), nil
}

func createSmiConfigSyncer(ctx context.Context, cs *clientset.Clientset) (v1.ConfigSyncer, error) {
	clients, err := clientset.SMIFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "initializing smi clients")
	}

	translator := smitranslator.NewTranslator()

	reconcilers := smi.NewSMIReconcilers(map[string]string{"created_by": "smi-config-syncer"},
		clients.TrafficTarget,
		clients.HTTPRouteGroup,
		clients.TrafficSplit,
	)

	newReporter := makeReporter("smi-config-reporter", cs.Supergloo)

	return smi.NewSmiConfigSyncer(translator, reconcilers, newReporter), nil
}

func makeReporter(name string, cs *clientset.SuperglooClients) reporter.Reporter {
	return reporter.NewReporter(name,
		cs.Mesh.BaseClient(),
		cs.Upstream.BaseClient(),
		cs.RoutingRule.BaseClient(),
		cs.SecurityRule.BaseClient())
}
