package snapshot

import (
	"context"
	"time"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type SnapshotValidator interface {
	// if the validator returns true, the snapshot should be considered valid
	// if the validator returns false, then:
	//   * the snapshot should be considered invalid and not used
	//   * error status(es) may have been written to the offending resource(s)
	Validate(MeshNetworkingSnapshot) bool
}

type NetworkingSnapshotGenerator interface {
	RegisterListener(SnapshotListener)

	// push the current snapshot to the listeners when all of the following conditions hold:
	//   * another `snapshotFrequency` period has elapsed
	//   * <-ctx.Done() has not been signaled
	//   * the snapshot has changed since the last time it was pushed out to the listeners
	// this method does not block the current goroutine
	StartPushingSnapshots(ctx context.Context, snapshotFrequency time.Duration)
}
