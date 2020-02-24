package snapshot

import (
	"context"
	"sync"
	"time"

	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controllers "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controllers "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
)

type MeshNetworkingSnapshot struct {
	// the current state of the world
	CurrentState Resources

	// specific resources that have changed since the last snapshot
	Delta Delta
}

type Resources struct {
	MeshServices  []*discovery_v1alpha1.MeshService
	MeshGroups    []*networking_v1alpha1.MeshGroup
	MeshWorkloads []*discovery_v1alpha1.MeshWorkload
}

type UpdatedMeshService struct {
	Old *discovery_v1alpha1.MeshService
	New *discovery_v1alpha1.MeshService
}

type UpdatedMeshGroup struct {
	Old *networking_v1alpha1.MeshGroup
	New *networking_v1alpha1.MeshGroup
}

type UpdatedMeshWorkload struct {
	Old *discovery_v1alpha1.MeshWorkload
	New *discovery_v1alpha1.MeshWorkload
}

type UpdatedResources struct {
	MeshServices  []UpdatedMeshService
	MeshGroups    []UpdatedMeshGroup
	MeshWorkloads []UpdatedMeshWorkload
}

type Delta struct {
	Created Resources
	Updated UpdatedResources
	Deleted Resources
}

type SnapshotListener func(context.Context, MeshNetworkingSnapshot)

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

// an implementation of `NetworkingSnapshotGenerator` that is guaranteed to only ever push
// snapshots that are considered valid by the `SnapshotValidator` to its listeners
func NewNetworkingSnapshotGenerator(
	ctx context.Context,
	snapshotValidator SnapshotValidator,
	meshServiceController discovery_controllers.MeshServiceController,
	meshGroupController networking_controllers.MeshGroupController,
	meshWorkloadController discovery_controllers.MeshWorkloadController,
) NetworkingSnapshotGenerator {
	generator := &networkingSnapshotGenerator{
		snapshotValidator: snapshotValidator,
		snapshot:          MeshNetworkingSnapshot{},
	}

	meshServiceController.AddEventHandler(ctx, &discovery_controllers.MeshServiceEventHandlerFuncs{
		OnCreate: func(obj *discovery_v1alpha1.MeshService) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			updatedMeshServices := append([]*discovery_v1alpha1.MeshService{}, generator.snapshot.CurrentState.MeshServices...)
			updatedMeshServices = append(updatedMeshServices, obj)

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshServices = updatedMeshServices

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Created.MeshServices = append(generator.snapshot.Delta.Created.MeshServices, obj)
			}

			return nil
		},
		OnUpdate: func(old, new *discovery_v1alpha1.MeshService) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			var updatedMeshServices []*discovery_v1alpha1.MeshService
			for _, existingMeshService := range generator.snapshot.CurrentState.MeshServices {
				if existingMeshService.GetName() == old.GetName() && existingMeshService.GetNamespace() == old.GetNamespace() {
					updatedMeshServices = append(updatedMeshServices, new)
				} else {
					updatedMeshServices = append(updatedMeshServices, existingMeshService)
				}
			}

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshServices = updatedMeshServices

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Updated.MeshServices = append(generator.snapshot.Delta.Updated.MeshServices, UpdatedMeshService{
					Old: old,
					New: new,
				})
			}

			return nil
		},
		OnDelete: func(obj *discovery_v1alpha1.MeshService) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			var updatedMeshServices []*discovery_v1alpha1.MeshService
			for _, meshService := range generator.snapshot.CurrentState.MeshServices {
				if meshService.GetName() == obj.GetName() && meshService.GetNamespace() == obj.GetNamespace() {
					continue
				}

				updatedMeshServices = append(updatedMeshServices, meshService)
			}

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshServices = updatedMeshServices

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Deleted.MeshServices = append(generator.snapshot.Delta.Deleted.MeshServices, obj)
			}

			return nil
		},
		OnGeneric: func(obj *discovery_v1alpha1.MeshService) error {
			return nil
		},
	})

	meshGroupController.AddEventHandler(ctx, &networking_controllers.MeshGroupEventHandlerFuncs{
		OnCreate: func(obj *networking_v1alpha1.MeshGroup) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			updatedMeshGroups := append([]*networking_v1alpha1.MeshGroup{}, generator.snapshot.CurrentState.MeshGroups...)
			updatedMeshGroups = append(updatedMeshGroups, obj)

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshGroups = updatedMeshGroups

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Created.MeshGroups = append(generator.snapshot.Delta.Created.MeshGroups, obj)
			}

			return nil
		},
		OnUpdate: func(old, new *networking_v1alpha1.MeshGroup) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			var updatedMeshGroups []*networking_v1alpha1.MeshGroup
			for _, existingMeshGroup := range generator.snapshot.CurrentState.MeshGroups {
				if existingMeshGroup.GetName() == old.GetName() && existingMeshGroup.GetNamespace() == old.GetNamespace() {
					updatedMeshGroups = append(updatedMeshGroups, new)
				} else {
					updatedMeshGroups = append(updatedMeshGroups, existingMeshGroup)
				}
			}

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshGroups = updatedMeshGroups

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Updated.MeshGroups = append(generator.snapshot.Delta.Updated.MeshGroups, UpdatedMeshGroup{
					Old: old,
					New: new,
				})
			}

			return nil
		},
		OnDelete: func(obj *networking_v1alpha1.MeshGroup) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			var updatedMeshGroups []*networking_v1alpha1.MeshGroup
			for _, meshGroup := range generator.snapshot.CurrentState.MeshGroups {
				if meshGroup.GetName() == obj.GetName() && meshGroup.GetNamespace() == obj.GetNamespace() {
					continue
				}

				updatedMeshGroups = append(updatedMeshGroups, meshGroup)
			}

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshGroups = updatedMeshGroups

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Deleted.MeshGroups = append(generator.snapshot.Delta.Deleted.MeshGroups, obj)
			}

			return nil
		},
		OnGeneric: func(obj *networking_v1alpha1.MeshGroup) error {
			return nil
		},
	})

	meshWorkloadController.AddEventHandler(ctx, &discovery_controllers.MeshWorkloadEventHandlerFuncs{
		OnCreate: func(obj *discovery_v1alpha1.MeshWorkload) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			updatedMeshWorkloads := append([]*discovery_v1alpha1.MeshWorkload{}, generator.snapshot.CurrentState.MeshWorkloads...)
			updatedMeshWorkloads = append(updatedMeshWorkloads, obj)

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshWorkloads = updatedMeshWorkloads

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Created.MeshWorkloads = append(generator.snapshot.Delta.Created.MeshWorkloads, obj)
			}

			return nil
		},
		OnUpdate: func(old, new *discovery_v1alpha1.MeshWorkload) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			var updatedMeshWorkloads []*discovery_v1alpha1.MeshWorkload
			for _, existingMeshWorkload := range generator.snapshot.CurrentState.MeshWorkloads {
				if existingMeshWorkload.GetName() == old.GetName() && existingMeshWorkload.GetNamespace() == old.GetNamespace() {
					updatedMeshWorkloads = append(updatedMeshWorkloads, new)
				} else {
					updatedMeshWorkloads = append(updatedMeshWorkloads, existingMeshWorkload)
				}
			}

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshWorkloads = updatedMeshWorkloads

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Updated.MeshWorkloads = append(generator.snapshot.Delta.Updated.MeshWorkloads, UpdatedMeshWorkload{
					Old: old,
					New: new,
				})
			}

			return nil
		},
		OnDelete: func(obj *discovery_v1alpha1.MeshWorkload) error {
			generator.snapshotMutex.Lock()
			defer generator.snapshotMutex.Unlock()

			var updatedMeshWorkloads []*discovery_v1alpha1.MeshWorkload
			for _, meshWorkload := range generator.snapshot.CurrentState.MeshWorkloads {
				if meshWorkload.GetName() == obj.GetName() && meshWorkload.GetNamespace() == obj.GetNamespace() {
					continue
				}

				updatedMeshWorkloads = append(updatedMeshWorkloads, meshWorkload)
			}

			updatedSnapshot := generator.snapshot
			updatedSnapshot.CurrentState.MeshWorkloads = updatedMeshWorkloads

			if generator.snapshotValidator.Validate(updatedSnapshot) {
				generator.isSnapshotPushNeeded = true
				generator.snapshot = updatedSnapshot
				generator.snapshot.Delta.Deleted.MeshWorkloads = append(generator.snapshot.Delta.Deleted.MeshWorkloads, obj)
			}

			return nil
		},
		OnGeneric: func(obj *discovery_v1alpha1.MeshWorkload) error {
			return nil
		},
	})

	return generator
}

type networkingSnapshotGenerator struct {
	snapshotValidator SnapshotValidator

	listeners     []SnapshotListener
	listenerMutex sync.Mutex

	// important that snapshot is NOT a reference- we depend on being able to copy it
	// and change fields without mutating the real thing
	// accesses to `isSnapshotPushNeeded` should be gated on the `snapshotMutex`
	snapshot             MeshNetworkingSnapshot
	isSnapshotPushNeeded bool
	snapshotMutex        sync.Mutex
}

func (f *networkingSnapshotGenerator) RegisterListener(listener SnapshotListener) {
	f.listenerMutex.Lock()
	defer f.listenerMutex.Unlock()

	f.listeners = append(f.listeners, listener)
}

func (f *networkingSnapshotGenerator) StartPushingSnapshots(ctx context.Context, snapshotFrequency time.Duration) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(snapshotFrequency):
				f.snapshotMutex.Lock()
				f.listenerMutex.Lock()

				if f.isSnapshotPushNeeded {
					for _, listener := range f.listeners {
						listener(ctx, f.snapshot)
					}

					f.isSnapshotPushNeeded = false
				}

				// important to unlock the mutexes in the same order as they were locked here
				// it's a runtime error to attempt to unlock an already unlocked mutex
				// if the order is changed here, a race condition could cause a repeated unlock
				f.snapshotMutex.Unlock()
				f.listenerMutex.Unlock()
			}
		}
	}()
}
