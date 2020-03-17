package cert_manager

import (
	"context"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	"go.uber.org/zap"
)

var (
	VMCSRSnapshotListenerSet = wire.NewSet(
		NewIstioCertConfigProducer,
		NewVirtualMeshCsrProcessor,
		NewVMCSRSnapshotListener,
	)

	NoVirtualMeshsChangedMessage = "no virtual meshes were created or updated during this sync"
)

type VMCSRSnapshotListener snapshot.MeshNetworkingSnapshotListener

func NewVMCSRSnapshotListener(
	csrProcessor VirtualMeshCertificateManager,
	virtualMeshClient zephyr_networking.VirtualMeshClient,
) VMCSRSnapshotListener {
	return &snapshot.MeshNetworkingSnapshotListenerFunc{
		OnSync: func(ctx context.Context, snap *snapshot.MeshNetworkingSnapshot) {
			logger := contextutils.LoggerFrom(ctx)
			// If no virtual meshs have been updated return immediately
			if len(snap.VirtualMeshes) == 0 {
				logger.Debug(NoVirtualMeshsChangedMessage)
				return
			}

			for _, virtualMesh := range snap.VirtualMeshes {
				status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, virtualMesh)
				if status.CertificateStatus.Status != types.ComputedStatus_ACCEPTED {
					logger.Debugw("csr processor failed", zap.Error(eris.New(status.CertificateStatus.Message)))
				}
				virtualMesh.Status = status
				err := virtualMeshClient.UpdateStatus(ctx, virtualMesh)
				if err != nil {
					logger.Errorf("Error updating certificate status on virtual mesh %+v", virtualMesh.ObjectMeta)
				}
			}
		},
	}
}
