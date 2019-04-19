package registration

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"go.uber.org/zap"
)

// registration syncer, activates config syncers based on registered meshes
// enables istio config syncer as long as there's a registered istio mesh
type RegistrationSyncer struct {
	configLoops ConfigLoopStarters
}

func NewRegistrationSyncer(configLoop ...ConfigLoopStarter) *RegistrationSyncer {
	return &RegistrationSyncer{configLoops: configLoop}
}

func (s *RegistrationSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	var enabledFeatures EnabledConfigLoops
	for _, mesh := range snap.Meshes.List() {
		switch mesh.MeshType.(type) {
		case *v1.Mesh_Istio:
			enabledFeatures.Istio = true
			contextutils.LoggerFrom(ctx).Infof("detected istio mesh %v", mesh.GetMetadata().Ref())
		case *v1.Mesh_LinkerdMesh:
			enabledFeatures.Linkerd = true
			contextutils.LoggerFrom(ctx).Infof("detected linkerd mesh %v", mesh.GetMetadata().Ref())
		case *v1.Mesh_AwsAppMesh:
			enabledFeatures.AppMesh = true
			contextutils.LoggerFrom(ctx).Infof("detected AWS AppMesh mesh %v", mesh.GetMetadata().Ref())
		}
	}

	for _, meshIngress := range snap.Meshingresses.List() {
		_, ok := meshIngress.MeshIngressType.(*v1.MeshIngress_Gloo)
		if ok {
			enabledFeatures.Gloo = true
			contextutils.LoggerFrom(ctx).Infof("detected gloo mesh-ingress %v", meshIngress.GetMetadata().Ref())
			break
		}
	}

	for _, loopFunc := range s.configLoops {
		if err := RunConfigLoop(ctx, enabledFeatures, loopFunc); err != nil {
			return err
		}
	}

	return nil
}
func RunConfigLoop(ctx context.Context, enabledFeatures EnabledConfigLoops, starter ConfigLoopStarter) error {
	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Minute * 1,
	}

	loop, err := starter(ctx, enabledFeatures)
	if err != nil {
		return err
	}

	if loop == nil {
		return nil
	}

	return RunEventLoop(ctx, loop, watchOpts)

}

func RunEventLoop(ctx context.Context, loop eventloop.EventLoop, opts clients.WatchOpts) error {
	logger := contextutils.LoggerFrom(ctx)
	combinedErrs, err := loop.Run(nil, opts)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case err := <-combinedErrs:
				if err != nil {
					logger.With(zap.Error(err)).Info("config event loop failure")
				}
			case <-ctx.Done():
			}
		}
	}()
	return nil
}
