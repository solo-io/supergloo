package setup

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/configutils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/envutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/internal/config"
	mp_kube "github.com/solo-io/mesh-projects/services/internal/kube"
	"github.com/solo-io/mesh-projects/services/internal/mcutils"
	"github.com/solo-io/mesh-projects/services/mesh-config/pkg/syncer"
	cr_translator "github.com/solo-io/mesh-projects/services/mesh-config/pkg/translator"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, writeNamespace string, errHandler func(error)) error {
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("waiting for crds to be registered before starting up",
		zap.Any("service entry", v1alpha1.RbacConfigCrd))

	// Do not start running until rbac crd has been registered
	ticker := time.Tick(2 * time.Second)
	cfg := mp_kube.MustGetKubeConfig(ctx)
	eg := &errgroup.Group{}
	eg.Go(func() error {
		for {
			select {
			case <-ticker:
				if config.CrdsExist(cfg, v1alpha1.RbacConfigCrd, v1alpha1.ClusterRbacConfigCrd) {
					return nil
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	logger.Infow("running", zap.Any("write namespace", writeNamespace))
	podNamespace := envutils.MustGetPodNamespace(ctx)
	kubernetesInterface := mp_kube.MustGetClient(ctx, cfg)
	configMapClient := configutils.NewConfigMapClient(kubernetesInterface)
	operatorConfig, err := config.GetOperatorConfig(ctx, configMapClient, podNamespace)
	if err != nil {
		return err
	}
	initialSettings := config.GetInitialSettings(podNamespace, operatorConfig)
	watchAggregator := wrapper.NewWatchAggregator()
	clientForClusterHandler := mcutils.NewClientForClusterHandler(watchAggregator)
	cacheManager, err := config.NewCacheManager(ctx)
	if err != nil {
		return err
	}

	clientSet := config.MustGetMeshConfigClientSet(ctx, cacheManager, cfg, clientForClusterHandler, initialSettings)
	meshReconciler := v1.NewMeshReconciler(clientSet.Mesh())
	rbacConfigReconciler := v1alpha1.NewClusterRbacConfigReconciler(clientSet.ClusterRbacConfig())
	translator := cr_translator.NewTranslator(clientSet)
	emitter := v1.NewRbacEmitter(clientSet.Mesh())
	syncer := syncer.NewRbacSyncer(writeNamespace, meshReconciler, rbacConfigReconciler, translator)
	eventLoop := v1.NewRbacEventLoop(emitter, syncer)
	errs, err := eventLoop.Run(nil, clients.WatchOpts{})
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-errs:
				errHandler(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
