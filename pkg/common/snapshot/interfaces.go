package snapshot

import (
	"context"
	"time"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type MeshNetworkingSnapshotValidator interface {
	// if the validator returns true, the snapshot should be considered valid
	// if the validator returns false, then:
	//   * the snapshot should be considered invalid and not used
	//   * error status(es) may have been written to the offending resource(s)
	ValidateVirtualMeshUpsert(ctx context.Context, obj *smh_networking.VirtualMesh, snapshot *MeshNetworkingSnapshot) bool

	ValidateVirtualMeshDelete(ctx context.Context, obj *smh_networking.VirtualMesh, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshServiceUpsert(ctx context.Context, obj *smh_discovery.MeshService, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshServiceDelete(ctx context.Context, obj *smh_discovery.MeshService, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshWorkloadUpsert(ctx context.Context, obj *smh_discovery.MeshWorkload, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshWorkloadDelete(ctx context.Context, obj *smh_discovery.MeshWorkload, snapshot *MeshNetworkingSnapshot) bool
}

type MeshNetworkingSnapshotGenerator interface {
	RegisterListener(MeshNetworkingSnapshotListener)

	// push the current snapshot to the listeners when all of the following conditions hold:
	//   * another `snapshotFrequency` period has elapsed
	//   * <-ctx.Done() has not been signaled
	//   * the snapshot has changed since the last time it was pushed out to the listeners
	// this method blocks, so start it in a separate go routine
	StartPushingSnapshots(ctx context.Context, snapshotFrequency time.Duration)
}

type MeshNetworkingSnapshotListener interface {
	Sync(context.Context, *MeshNetworkingSnapshot)
}

type MeshNetworkingSnapshotListenerFunc struct {
	OnSync func(context.Context, *MeshNetworkingSnapshot)
}

func (m *MeshNetworkingSnapshotListenerFunc) Sync(ctx context.Context, snap *MeshNetworkingSnapshot) {
	if m.OnSync != nil {
		m.OnSync(ctx, snap)
	}
}
