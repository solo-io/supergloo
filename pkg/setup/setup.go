package setup

import (
	"context"
	"os"
	"time"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	custombase "github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	customkube "github.com/solo-io/supergloo/pkg/api/external/kubernetes/core/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"
	"k8s.io/client-go/kubernetes"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/stats"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	policyv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	istioinstall "github.com/solo-io/supergloo/pkg/install/istio"
)

// customCtx and customErrHandler are expected to be passed by tests
func Main(customCtx context.Context, customErrHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	rootCtx := createRootContext(customCtx)
	logger := contextutils.LoggerFrom(rootCtx)

	errHandler := func(err error) {
		if err == nil {
			return
		}
		logger.Errorf("install event loop error: %v", err)
		if customErrHandler != nil {
			customErrHandler(err)
		}
	}

	clients, err := createClients(rootCtx)
	if err != nil {
		return errors.Wrap(err, "initializing clients")
	}

	return runInstallEventLoop(rootCtx, errHandler, clients, createInstallSyncers(clients.installClient, clients.meshClient))
}

func createRootContext(customCtx context.Context) context.Context {
	rootCtx := customCtx
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	rootCtx = contextutils.WithLogger(rootCtx, "supergloo")
	return rootCtx
}

type clientset struct {
	// input clients
	installClient      v1.InstallClient
	meshClient         v1.MeshClient
	upstreamClient     gloov1.UpstreamClient
	podClient          customkube.PodClient
	routingRuleClient  v1.RoutingRuleClient
	securityRuleClient v1.SecurityRuleClient

	// output clients
	IstioClients
}

type IstioClients struct {
	rbacConfigClient         rbacv1alpha1.RbacConfigClient
	serviceRoleClient        rbacv1alpha1.ServiceRoleClient
	serviceRoleBindingClient rbacv1alpha1.ServiceRoleBindingClient
	meshPolicyClient         policyv1alpha1.MeshPolicyClient
	destinationRuleClient    v1alpha3.DestinationRuleClient
	virtualServiceClient     v1alpha3.VirtualServiceClient
}

func createClients(ctx context.Context) (*clientset, error) {
	kubeCache := kube.NewKubeCache(ctx)
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}

	/*
		supergloo clients
	*/

	installClient, err := v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := installClient.Register(); err != nil {
		return nil, err
	}

	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := meshClient.Register(); err != nil {
		return nil, err
	}

	upstreamClient, err := gloov1.NewUpstreamClient(&factory.KubeResourceClientFactory{
		Crd:         gloov1.UpstreamCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := upstreamClient.Register(); err != nil {
		return nil, err
	}

	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	coreCache, err := cache.NewKubeCoreCache(ctx, kube)
	if err != nil {
		return nil, err
	}
	podClient := customkube.NewPodClientWithBase(custombase.NewResourceClient(kube, coreCache))
	if err := podClient.Register(); err != nil {
		return nil, err
	}

	routingRuleClient, err := v1.NewRoutingRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1.RoutingRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := routingRuleClient.Register(); err != nil {
		return nil, err
	}

	securityRuleClient, err := v1.NewSecurityRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1.SecurityRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := securityRuleClient.Register(); err != nil {
		return nil, err
	}

	/*
		istio clients
	*/
	rbacConfigClient, err := rbacv1alpha1.NewRbacConfigClient(&factory.KubeResourceClientFactory{
		Crd:         rbacv1alpha1.RbacConfigCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := rbacConfigClient.Register(); err != nil {
		return nil, err
	}

	serviceRoleClient, err := rbacv1alpha1.NewServiceRoleClient(&factory.KubeResourceClientFactory{
		Crd:         rbacv1alpha1.ServiceRoleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := serviceRoleClient.Register(); err != nil {
		return nil, err
	}

	serviceRoleBindingClient, err := rbacv1alpha1.NewServiceRoleBindingClient(&factory.KubeResourceClientFactory{
		Crd:         rbacv1alpha1.ServiceRoleBindingCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := serviceRoleBindingClient.Register(); err != nil {
		return nil, err
	}

	meshPolicyClient, err := policyv1alpha1.NewMeshPolicyClient(&factory.KubeResourceClientFactory{
		Crd:         policyv1alpha1.MeshPolicyCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := meshPolicyClient.Register(); err != nil {
		return nil, err
	}

	destinationRuleClient, err := v1alpha3.NewDestinationRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha3.DestinationRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := destinationRuleClient.Register(); err != nil {
		return nil, err
	}

	virtualServiceClient, err := v1alpha3.NewVirtualServiceClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha3.VirtualServiceCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, err
	}
	if err := virtualServiceClient.Register(); err != nil {
		return nil, err
	}

	return &clientset{
		installClient:      installClient,
		meshClient:         meshClient,
		upstreamClient:     upstreamClient,
		podClient:          podClient,
		routingRuleClient:  routingRuleClient,
		securityRuleClient: securityRuleClient,
		IstioClients: IstioClients{
			rbacConfigClient:         rbacConfigClient,
			serviceRoleClient:        serviceRoleClient,
			serviceRoleBindingClient: serviceRoleBindingClient,
			meshPolicyClient:         meshPolicyClient,
			destinationRuleClient:    destinationRuleClient,
			virtualServiceClient:     virtualServiceClient,
		},
	}, nil
}

func runInstallEventLoop(ctx context.Context, errHandler func(err error), c *clientset, syncers v1.InstallSyncers) error {
	installEmitter := v1.NewInstallEmitter(c.installClient)
	installEventLoop := v1.NewInstallEventLoop(installEmitter, syncers)

	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	installEventLoopErrs, err := installEventLoop.Run(nil, watchOpts)
	if err != nil {
		return err
	}
	for {
		select {
		case err := <-installEventLoopErrs:
			errHandler(err)
		case <-ctx.Done():
			return nil
		}
	}
}

// Add install syncers here
func createInstallSyncers(installClient v1.InstallClient, meshClient v1.MeshClient) v1.InstallSyncers {
	return v1.InstallSyncers{
		istioinstall.NewInstallSyncer(nil,
			meshClient,
			reporter.NewReporter("istio-install-reporter", installClient.BaseClient())),
	}
}

// Add config syncers here
func createConfigSyncers(kc kubernetes.Interface,
	meshClient v1.MeshClient,
	upstreamClient gloov1.UpstreamClient,
	podClient customkube.PodClient,
	routingRuleClient v1.RoutingRuleClient,
	securityRuleClient v1.SecurityRuleClient,
	istioClients IstioClients) v1.ConfigSyncer {

	translator := istio.NewTranslator(plugins.Plugins(kc))

	reconcilers := istio.NewIstioReconcilers(map[string]string{"owner_ref": "istio-config-syncer"},
		rbacv1alpha1.NewRbacConfigReconciler(istioClients.rbacConfigClient),
		rbacv1alpha1.NewServiceRoleReconciler(istioClients.serviceRoleClient),
		rbacv1alpha1.NewServiceRoleBindingReconciler(istioClients.serviceRoleBindingClient),
		policyv1alpha1.NewMeshPolicyReconciler(istioClients.meshPolicyClient),
		v1alpha3.NewDestinationRuleReconciler(istioClients.destinationRuleClient),
		v1alpha3.NewVirtualServiceReconciler(istioClients.virtualServiceClient),
	)

	reporter := reporter.NewReporter("istio-config-reporter",
		meshClient.BaseClient(),
		upstreamClient.BaseClient(),
		podClient.BaseClient(),
		routingRuleClient.BaseClient(),
		securityRuleClient.BaseClient())

	return v1.ConfigSyncers{
		istio.NewIstioConfigSyncer(translator, reconcilers, reporter),
	}
}

/*
meshes := snapshot.Meshes.List()
	meshGroups := snapshot.Meshgroups.List()
	upstreams := snapshot.Upstreams.List()
	pods := snapshot.Pods.List()
	routingRules := snapshot.Routingrules.List()
	securityRules := snapshot.Securityrules.List()

*/
