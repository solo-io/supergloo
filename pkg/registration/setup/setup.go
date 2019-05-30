package setup

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/solo-io/go-utils/errors"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/registration"
	"github.com/solo-io/supergloo/pkg/registration/appmesh"
	"github.com/solo-io/supergloo/pkg/registration/gloo"
	istio2 "github.com/solo-io/supergloo/pkg/registration/gloo/istio"
	"github.com/solo-io/supergloo/pkg/registration/gloo/linkerd"
	"github.com/solo-io/supergloo/pkg/registration/istio"
	istiostats "github.com/solo-io/supergloo/pkg/stats/istio"
	linkerdstats "github.com/solo-io/supergloo/pkg/stats/linkerd"
	"k8s.io/helm/pkg/kube"
)

type Options struct {
	SuperglooNamespace             string
	SidecarInjectorImageName       string
	SidecarInjectorImagePullPolicy string
	DisablePrometheusBouncer       bool
}

func RunRegistrationEventLoop(ctx context.Context, cs *clientset.Clientset, customErrHandler func(error), pubsub *registration.PubSub) error {
	ctx = contextutils.WithLogger(ctx, "registration-event-loop")
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

	opts, err := getOptions()
	if err != nil {
		return err
	}

	registrationSyncers := createRegistrationSyncers(cs, pubsub, opts)

	if err := runRegistrationEventLoop(ctx, errHandler, cs, registrationSyncers); err != nil {
		return err
	}

	return nil
}

// Add registration syncers here
func createRegistrationSyncers(clientset *clientset.Clientset, pubSub *registration.PubSub, opts Options) v1.RegistrationSyncer {
	skipPrometheusBounce := opts.DisablePrometheusBouncer
	return v1.RegistrationSyncers{
		istio.NewIstioSecretDeleter(clientset.Kube),
		istiostats.NewIstioPrometheusSyncer(clientset.Prometheus, clientset.Kube, skipPrometheusBounce),
		linkerdstats.NewLinkerdPrometheusSyncer(clientset.Prometheus, clientset.Kube, skipPrometheusBounce),
		gloo.NewGlooRegistrationSyncer(
			clientset,
			linkerd.NewGlooLinkerdMtlsPlugin(clientset),
			istio2.NewGlooIstioMtlsPlugin(clientset),
		),
		appmesh.NewAppMeshRegistrationSyncer(
			reporter.NewReporter("app-mesh-registration-reporter", clientset.Supergloo.Mesh.BaseClient()),
			clientset.Kube,
			clientset.Supergloo.Secret,
			kube.New(nil),
			opts.SuperglooNamespace,
			opts.SidecarInjectorImageName,
			opts.SidecarInjectorImagePullPolicy,
		),
		registration.NewRegistrationSyncer(pubSub),
	}
}

// start the registration event loop
func runRegistrationEventLoop(ctx context.Context, errHandler func(err error), clientset *clientset.Clientset, syncers v1.RegistrationSyncer) error {
	registrationEmitter := v1.NewRegistrationEmitter(clientset.Supergloo.Mesh, clientset.Supergloo.MeshIngress)
	registrationEventLoop := v1.NewRegistrationEventLoop(registrationEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	registrationEventLoopErrs, err := registrationEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-registrationEventLoopErrs:
				errHandler(err)
			case <-ctx.Done():
			}
		}
	}()
	return nil
}

func getOptions() (Options, error) {
	var opts Options

	flag.BoolVar(&opts.DisablePrometheusBouncer, "disable-prometheus-bouncer", false, "disable automatic "+
		"bouncing of prometheus pods on config reload. enable this option if using the prometheus configmap reload")
	flag.Parse()

	if opts.SuperglooNamespace = os.Getenv(constants.PodNamespaceEnvName); opts.SuperglooNamespace == "" {
		return Options{}, errors.Errorf("%s env must be set", constants.PodNamespaceEnvName)
	}

	if opts.SidecarInjectorImageName = os.Getenv(constants.SidecarInjectorImageNameEnvName); opts.SidecarInjectorImageName == "" {
		return Options{}, errors.Errorf("%s env must be set", constants.SidecarInjectorImageNameEnvName)
	}

	if opts.SidecarInjectorImagePullPolicy = os.Getenv(constants.SidecarInjectorImagePullPolicyEnvName); opts.SidecarInjectorImagePullPolicy == "" {
		return Options{}, errors.Errorf("%s env must be set", constants.SidecarInjectorImagePullPolicyEnvName)
	}

	return opts, nil
}
