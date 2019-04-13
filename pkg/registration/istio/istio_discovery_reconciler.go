package istio

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

// The purpose of this syncer is to take the result of the mesh discovery and reconcile istio meshes with
// that data
type IstioDiscoveryReconciler struct {
	kube           kubernetes.Interface
	meshClient     v1.MeshClient
	meshReconciler v1.MeshReconciler
}

func NewIstioDiscoveryReconciler(kube kubernetes.Interface, meshClient v1.MeshClient) *IstioDiscoveryReconciler {
	meshReconciler := v1.NewMeshReconciler(meshClient)
	return &IstioDiscoveryReconciler{kube: kube, meshClient: meshClient, meshReconciler: meshReconciler}
}

func (d *IstioDiscoveryReconciler) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-secret-deleter-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)

	meshes := snap.Meshes.List()

	// delete citadel certs for any istio mesh which specifies a root cert
	for _, mesh := range meshes {
		istioMesh := mesh.GetIstio()
		if istioMesh == nil {
			continue
		}
		if mesh.DiscoveryMetadata.MtlsConfig == nil || !mesh.DiscoveryMetadata.MtlsConfig.MtlsEnabled ||
			mesh.DiscoveryMetadata.MtlsConfig.RootCertificate == nil {
			continue
		}
		logger.With(zap.String("mesh", mesh.Metadata.Ref().Key())).Debugf("updating mtls config")
		mesh.MtlsConfig = mesh.DiscoveryMetadata.MtlsConfig
	}

	return d.meshReconciler.Reconcile("", meshes, nil, clients.ListOpts{})
}
