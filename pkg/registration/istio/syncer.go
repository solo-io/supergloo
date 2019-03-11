package istio

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/istio/setup"
)

// registration syncer for istio
// enables istio config syncer as long as there's a registered istio mesh
type IstioRegistrationSyncer struct {
	Clientset  *clientset.Clientset
	ErrHandler func(error)
}

func NewIstioRegistrationSyncer(clientset *clientset.Clientset, errHandler func(error)) *IstioRegistrationSyncer {
	return &IstioRegistrationSyncer{Clientset: clientset, ErrHandler: errHandler}
}

func (s *IstioRegistrationSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	var enableIstioFeatures bool
	for _, mesh := range snap.Meshes.List() {
		_, ok := mesh.MeshType.(*v1.Mesh_Istio)
		if ok {
			enableIstioFeatures = true
			break
		}
	}
	if !enableIstioFeatures {
		return nil
	}
	contextutils.LoggerFrom(ctx).Infof("detected istio installation, enabling istio config syncer")

	return setup.RunIstioConfigEventLoop(ctx, s.Clientset, s.ErrHandler)
}
