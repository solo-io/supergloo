package registration

import (
	"context"

	"github.com/solo-io/supergloo/pkg/config/setup"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// registration syncer, activates config syncers based on registered meshes
// enables istio config syncer as long as there's a registered istio mesh
type RegistrationSyncer struct {
	Clientset  *clientset.Clientset
	ErrHandler func(error)
}

func NewRegistrationSyncer(clientset *clientset.Clientset, errHandler func(error)) *RegistrationSyncer {
	return &RegistrationSyncer{Clientset: clientset, ErrHandler: errHandler}
}

func (s *RegistrationSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	var enabledFeatures setup.EnabledConfigLoops
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

	return setup.RunConfigEventLoop(ctx, s.Clientset, s.ErrHandler, enabledFeatures)
}
