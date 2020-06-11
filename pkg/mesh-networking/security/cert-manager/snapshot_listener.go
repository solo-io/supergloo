package cert_manager

import (
	"context"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	networking_snapshot "github.com/solo-io/service-mesh-hub/pkg/common/snapshot"
	"go.uber.org/zap"
)

var (
	VMCSRSnapshotListenerSet = wire.NewSet(
		NewIstioCertConfigProducer,
		NewVirtualMeshCsrProcessor,
		NewVMCSRSnapshotListener,
	)

	NoVirtualMeshesChangedMessage = "no virtual meshes were created or updated during this sync"
)

type VMCSRSnapshotListener networking_snapshot.MeshNetworkingSnapshotListener

func NewVMCSRSnapshotListener(
	csrProcessor VirtualMeshCertificateManager,
	virtualMeshClient smh_networking.VirtualMeshClient,
) VMCSRSnapshotListener {
	return &networking_snapshot.MeshNetworkingSnapshotListenerFunc{
		OnSync: func(ctx context.Context, snap *networking_snapshot.MeshNetworkingSnapshot) {
			logger := contextutils.LoggerFrom(ctx)
			// If no virtual meshes have been updated return immediately
			if len(snap.VirtualMeshes) == 0 {
				logger.Debug(NoVirtualMeshesChangedMessage)
				return
			}

			for _, virtualMesh := range snap.VirtualMeshes {
				status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, virtualMesh)
				if status.CertificateStatus.State != smh_core_types.Status_ACCEPTED {
					logger.Debugw("csr processor failed", zap.Error(eris.New(status.CertificateStatus.Message)))
				}
				virtualMesh.Status = status
				err := virtualMeshClient.UpdateVirtualMeshStatus(ctx, virtualMesh)
				if err != nil {
					logger.Errorf("Error updating certificate status on virtual mesh %s.%s",
						virtualMesh.Name, virtualMesh.Namespace)
				}
			}
		},
	}
}
