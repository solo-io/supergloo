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
	GroupMgcsrSnapshotListenerSet = wire.NewSet(
		NewIstioCertConfigProducer,
		NewMeshGroupCsrProcessor,
		NewGroupMgcsrSnapshotListener,
	)

	NoMeshGroupsChangedMessage = "no meshgroups were created or updated during this sync"
)

type GroupMgcsrSnapshotListener snapshot.MeshNetworkingSnapshotListener

func NewGroupMgcsrSnapshotListener(
	csrProcessor MeshGroupCertificateManager,
	meshGroupClient zephyr_networking.MeshGroupClient,
) GroupMgcsrSnapshotListener {
	return &snapshot.MeshNetworkingSnapshotListenerFunc{
		OnSync: func(ctx context.Context, snap *snapshot.MeshNetworkingSnapshot) {
			logger := contextutils.LoggerFrom(ctx)
			// If no mesh groups have been updated return immediately
			if len(snap.MeshGroups) == 0 {
				logger.Debug(NoMeshGroupsChangedMessage)
				return
			}

			for _, meshGroup := range snap.MeshGroups {
				status := csrProcessor.InitializeCertificateForMeshGroup(ctx, meshGroup)
				if status.CertificateStatus.Status != types.ComputedStatus_ACCEPTED {
					logger.Debugw("csr processor failed", zap.Error(eris.New(status.CertificateStatus.Message)))
				}
				meshGroup.Status = status
				err := meshGroupClient.UpdateStatus(ctx, meshGroup)
				if err != nil {
					logger.Errorf("Error updating certificate status on mesh group %+v", meshGroup.ObjectMeta)
				}
			}
		},
	}
}
