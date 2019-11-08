package setup

import (
	"context"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/mesh-config/pkg/rbac"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"go.uber.org/zap"
)

func Run(ctx context.Context, writeNamespace string, errHandler func(error)) error {
	contextutils.LoggerFrom(ctx).Infow("running", zap.Any("write namespace", writeNamespace))
	// istio rbac
	rbacConfigClient, err := initializeRbacConfigClient(ctx)
	if err != nil {
		return err
	}
	rbacConfigReconciler := v1alpha1.NewRbacConfigReconciler(rbacConfigClient)

	meshClient, err := initializeMeshClient(ctx)
	if err != nil {
		return err
	}
	meshReconciler := v1.NewMeshReconciler(meshClient)

	emitter := v1.NewRbacEmitter(meshClient, rbacConfigClient)
	syncer := rbac.NewRbacSyncer(writeNamespace, meshReconciler, rbacConfigReconciler)
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

func initializeRbacConfigClient(ctx context.Context) (v1alpha1.RbacConfigClient, error) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic("KUBECONFIG is not defined")
	}
	kubeCache := kube.NewKubeCache(ctx)
	rbacConfigClientFactory := factory.KubeResourceClientFactory{
		Crd:         v1alpha1.RbacConfigCrd,
		Cfg:         cfg,
		SharedCache: kubeCache,
		// this is registered by istio, however we need to register in case istio has not yet been deployed
		// so that the watches can start
		SkipCrdCreation: false,
	}

	client, err := v1alpha1.NewRbacConfigClient(&rbacConfigClientFactory)
	if err != nil {
		return nil, err
	}
	if err := client.Register(); err != nil {
		return nil, err
	}
	return client, nil
}

func initializeMeshClient(ctx context.Context) (v1.MeshClient, error) {
	cfg, err := kubeutils.GetConfig("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic("KUBECONFIG is not defined")
	}
	kubeCache := kube.NewKubeCache(ctx)
	meshClientFactory := factory.KubeResourceClientFactory{
		Crd:             v1.MeshCrd,
		Cfg:             cfg,
		SharedCache:     kubeCache,
		SkipCrdCreation: false,
	}

	client, err := v1.NewMeshClient(&meshClientFactory)
	if err != nil {
		return nil, err
	}
	if err := client.Register(); err != nil {
		return nil, err
	}
	return client, nil
}
