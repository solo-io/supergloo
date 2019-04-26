package registration

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// registration syncer, activates config syncers based on registered meshes
// enables istio config syncer as long as there's a registered istio mesh
type RegistrationSyncer struct {
	pubSub *PubSub
}

func NewRegistrationSyncer(pubSub *PubSub) *RegistrationSyncer {
	return &RegistrationSyncer{pubSub: pubSub}
}

func (s *RegistrationSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	var enabledFeatures EnabledConfigLoops
	for _, mesh := range snap.Meshes.List() {
		switch mesh.MeshType.(type) {
		case *v1.Mesh_Istio:
			enabledFeatures.Istio = true
			contextutils.LoggerFrom(ctx).Infof("detected istio mesh %v", mesh.GetMetadata().Ref())
		case *v1.Mesh_Linkerd:
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

	// Send updates to all subscribers
	s.pubSub.publish(ctx, enabledFeatures)

	return nil
}
