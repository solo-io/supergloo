package registration

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// registration syncer, activates config syncers based on registered meshes
// enables istio config syncer as long as there's a registered istio mesh
type RegistrationSyncer struct {
	configLoop ConfigLoop
}

func NewRegistrationSyncer(configLoop ConfigLoop) *RegistrationSyncer {
	return &RegistrationSyncer{configLoop: configLoop}
}

func (s *RegistrationSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	var enabledFeatures EnabledConfigLoops
	for _, mesh := range snap.Meshes.List() {
		_, ok := mesh.MeshType.(*v1.Mesh_Istio)
		if ok {
			enabledFeatures.Istio = true
			contextutils.LoggerFrom(ctx).Infof("detected istio mesh, enabling istio config syncer")
			break
		}
	}

	for _, meshIngress := range snap.Meshingresses.List() {
		_, ok := meshIngress.MeshIngressType.(*v1.MeshIngress_Gloo)
		if ok {
			enabledFeatures.Gloo = true
			contextutils.LoggerFrom(ctx).Infof("detected gloo mesh-ingress, enabling gloo config syncer")
			break
		}
	}

	for _, mesh := range snap.Meshes.List() {
		if mesh.GetAwsAppMesh() != nil {
			enabledFeatures.AppMesh = true
			contextutils.LoggerFrom(ctx).Infof("detected Aws App Mesh, enabling appmesh config syncer")
			break
		}
	}

	return s.configLoop.Run(ctx, enabledFeatures)
}
