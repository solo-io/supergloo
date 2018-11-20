package setup

import (
	"context"
	"time"

	"github.com/solo-io/supergloo/pkg/install"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/utils/errutils"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/consul"
	"github.com/solo-io/supergloo/pkg/translator/istio"
	"github.com/solo-io/supergloo/pkg/translator/linkerd2"
	"k8s.io/client-go/kubernetes"
)

func Main() error {
	// TODO: ilackarms: suport options
	kubeCache := kube.NewKubeCache()
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	destinationRuleClient, err := v1alpha3.NewDestinationRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha3.DestinationRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return err
	}
	if err := destinationRuleClient.Register(); err != nil {
		return err
	}

	virtualServiceClient, err := v1alpha3.NewVirtualServiceClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha3.VirtualServiceCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return err
	}
	if err := virtualServiceClient.Register(); err != nil {
		return err
	}

	prometheusClient, err := prometheusv1.NewConfigClient(&factory.KubeConfigMapClientFactory{
		Clientset: kubeClient,
	})
	if err != nil {
		return err
	}
	if err := prometheusClient.Register(); err != nil {
		return err
	}

	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return err
	}
	if err := meshClient.Register(); err != nil {
		return err
	}

	installClient, err := v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return err
	}
	if err := installClient.Register(); err != nil {
		return err
	}

	routingRuleClient, err := v1.NewRoutingRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1.RoutingRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return err
	}
	if err := routingRuleClient.Register(); err != nil {
		return err
	}

	upstreamClient, err := gloov1.NewUpstreamClient(&factory.KubeResourceClientFactory{
		Crd:         gloov1.UpstreamCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return err
	}
	if err := upstreamClient.Register(); err != nil {
		return err
	}

	secretClient, err := gloov1.NewSecretClient(&factory.KubeSecretClientFactory{
		Clientset: kubeClient,
	})
	if err != nil {
		return err
	}
	if err := secretClient.Register(); err != nil {
		return err
	}

	installEmitter := v1.NewInstallEmitter(installClient)

	translatorEmitter := v1.NewTranslatorEmitter(meshClient, routingRuleClient, upstreamClient, secretClient)

	rpt := reporter.NewReporter("supergloo", meshClient.BaseClient())
	writeErrs := make(chan error)

	istioRoutingSyncer := &istio.MeshRoutingSyncer{
		DestinationRuleReconciler: v1alpha3.NewDestinationRuleReconciler(destinationRuleClient),
		VirtualServiceReconciler:  v1alpha3.NewVirtualServiceReconciler(virtualServiceClient),
		Reporter:                  rpt,
		WriteSelector:             map[string]string{"reconciler.solo.io": "supergloo.istio.routing"},
		WriteNamespace:            "supergloo-system",
	}

	linkerd2PrometheusSyncer := linkerd2.NewPrometheusSyncer(kubeClient, prometheusClient)
	istioPrometheusSyncer := istio.NewPrometheusSyncer(kubeClient, prometheusClient)

	consulEncryptionSyncer := &consul.ConsulSyncer{}
	consulPolicySyncer := &consul.PolicySyncer{}
	istioEncryptionSyncer := &istio.EncryptionSyncer{
		Kube:         kubeClient,
		SecretClient: secretClient,
	}

	translatorSyncers := v1.TranslatorSyncers{
		istioRoutingSyncer,
		istioPrometheusSyncer,
		linkerd2PrometheusSyncer,
		consulEncryptionSyncer,
		consulPolicySyncer,
		istioEncryptionSyncer,
	}

	installSyncer := &install.InstallSyncer{
		Kube:       kubeClient,
		MeshClient: meshClient,
		// TODO: set a security client when we resolve minishift issues
	}
	installSyncers := v1.InstallSyncers{
		installSyncer,
	}

	translatorEventLoop := v1.NewTranslatorEventLoop(translatorEmitter, translatorSyncers)
	installEventLoop := v1.NewInstallEventLoop(installEmitter, installSyncers)

	ctx := contextutils.WithLogger(context.Background(), "supergloo")
	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Second * 1,
	}

	translatorEventLoopErrs, err := translatorEventLoop.Run([]string{"supergloo-system", "gloo-system"}, watchOpts)
	if err != nil {
		return err
	}
	go errutils.AggregateErrs(watchOpts.Ctx, writeErrs, translatorEventLoopErrs, "translator_event_loop")

	installEventLoopErrs, err := installEventLoop.Run([]string{"supergloo-system", "gloo-system"}, watchOpts)
	if err != nil {
		return err
	}
	go errutils.AggregateErrs(watchOpts.Ctx, writeErrs, installEventLoopErrs, "install_event_loop")

	logger := contextutils.LoggerFrom(watchOpts.Ctx)

	for {
		select {
		case err := <-writeErrs:
			logger.Errorf("error: %v", err)
		case <-watchOpts.Ctx.Done():
			close(writeErrs)
			return nil
		}
	}
}
