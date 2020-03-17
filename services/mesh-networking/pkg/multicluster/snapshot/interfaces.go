package snapshot

import (
	"context"
	"time"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type MeshNetworkingSnapshotValidator interface {
	// if the validator returns true, the snapshot should be considered valid
	// if the validator returns false, then:
	//   * the snapshot should be considered invalid and not used
	//   * error status(es) may have been written to the offending resource(s)
	ValidateVirtualMeshUpsert(ctx context.Context, obj *networking_v1alpha1.VirtualMesh, snapshot *MeshNetworkingSnapshot) bool

	ValidateVirtualMeshDelete(ctx context.Context, obj *networking_v1alpha1.VirtualMesh, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshServiceUpsert(ctx context.Context, obj *discovery_v1alpha1.MeshService, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshServiceDelete(ctx context.Context, obj *discovery_v1alpha1.MeshService, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshWorkloadUpsert(ctx context.Context, obj *discovery_v1alpha1.MeshWorkload, snapshot *MeshNetworkingSnapshot) bool

	ValidateMeshWorkloadDelete(ctx context.Context, obj *discovery_v1alpha1.MeshWorkload, snapshot *MeshNetworkingSnapshot) bool
}

type MeshNetworkingSnapshotGenerator interface {
	RegisterListener(MeshNetworkingSnapshotListener)

	// push the current snapshot to the listeners when all of the following conditions hold:
	//   * another `snapshotFrequency` period has elapsed
	//   * <-ctx.Done() has not been signaled
	//   * the snapshot has changed since the last time it was pushed out to the listeners
	// this method blocks, so start it in a seperate go routine
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
