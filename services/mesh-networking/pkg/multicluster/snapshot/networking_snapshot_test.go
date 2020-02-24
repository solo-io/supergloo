package snapshot_test

import (
	"context"
	"sync"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	mock_snapshot "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot/mocks"
	mock_zephyr_discovery "github.com/solo-io/mesh-projects/test/mocks/zephyr/discovery"
	mock_zephyr_networking "github.com/solo-io/mesh-projects/test/mocks/zephyr/networking"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Networking Snapshot", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		eventuallyTimeout   = time.Second
		consistentlyTimeout = time.Millisecond * 100 // don't want to make our tests take a ton of time to run
		pollFrequency       = time.Millisecond

		meshService1 = &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ms-1",
				Namespace: env.DefaultWriteNamespace,
			},
		}
		meshService2 = &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ms-2",
				Namespace: env.DefaultWriteNamespace,
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can receive events", func() {
		updatedSnapshot := snapshot.MeshNetworkingSnapshot{
			CurrentState: snapshot.Resources{
				MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
			},
		}

		validator := mock_snapshot.NewMockSnapshotValidator(ctrl)
		validator.EXPECT().
			Validate(updatedSnapshot).
			Return(true)

		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *controller.MeshServiceEventHandlerFuncs) error {
				return eventHandler.OnCreate(meshService1)
			})

		meshGroupController := mock_zephyr_networking.NewMockMeshGroupController(ctrl)
		meshGroupController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadController := mock_zephyr_discovery.NewMockMeshWorkloadController(ctrl)
		meshWorkloadController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator := snapshot.NewNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)

		didReceiveSnapshot := false
		var boolMutex sync.Mutex
		generator.RegisterListener(func(ctx context.Context, networkingSnapshot snapshot.MeshNetworkingSnapshot) {
			boolMutex.Lock()
			defer boolMutex.Unlock()
			didReceiveSnapshot = true

			Expect(networkingSnapshot.CurrentState.MeshServices).To(HaveLen(1))
			Expect(networkingSnapshot.CurrentState.MeshServices[0]).To(Equal(meshService1))
			Expect(networkingSnapshot.Delta.Created.MeshServices).To(HaveLen(1))
			Expect(networkingSnapshot.Delta.Created.MeshServices[0]).To(Equal(meshService1))
		})

		generator.StartPushingSnapshots(ctx, pollFrequency)

		Eventually(func() bool {
			boolMutex.Lock()
			defer boolMutex.Unlock()
			return didReceiveSnapshot
		}, eventuallyTimeout, pollFrequency).Should(BeTrue(), "Should eventually receive a snapshot")
	})

	It("should not push snapshots if nothing has changed", func() {
		updatedSnapshot := snapshot.MeshNetworkingSnapshot{
			CurrentState: snapshot.Resources{
				MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
			},
		}

		validator := mock_snapshot.NewMockSnapshotValidator(ctrl)
		validator.EXPECT().
			Validate(updatedSnapshot).
			Return(true)

		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *controller.MeshServiceEventHandlerFuncs) error {
				return eventHandler.OnCreate(meshService1)
			})

		meshGroupController := mock_zephyr_networking.NewMockMeshGroupController(ctrl)
		meshGroupController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadController := mock_zephyr_discovery.NewMockMeshWorkloadController(ctrl)
		meshWorkloadController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator := snapshot.NewNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)

		numReceivedSnapshots := 0
		var counterMutex sync.Mutex
		generator.RegisterListener(func(ctx context.Context, networkingSnapshot snapshot.MeshNetworkingSnapshot) {
			counterMutex.Lock()
			defer counterMutex.Unlock()
			numReceivedSnapshots += 1

			Expect(networkingSnapshot.CurrentState.MeshServices).To(HaveLen(1))
			Expect(networkingSnapshot.CurrentState.MeshServices[0]).To(Equal(meshService1))
			Expect(networkingSnapshot.Delta.Created.MeshServices).To(HaveLen(1))
			Expect(networkingSnapshot.Delta.Created.MeshServices[0]).To(Equal(meshService1))
		})

		generator.StartPushingSnapshots(ctx, pollFrequency)

		Eventually(func() int {
			counterMutex.Lock()
			defer counterMutex.Unlock()
			return numReceivedSnapshots
		}, eventuallyTimeout, pollFrequency).Should(BeNumerically(">=", 1), "Should eventually receive a snapshot")
		Consistently(func() int {
			counterMutex.Lock()
			defer counterMutex.Unlock()
			return numReceivedSnapshots
		}, consistentlyTimeout, pollFrequency).Should(Equal(1), "No further snapshots should be received")
	})

	It("can aggregate multiple events that roll in close to each other", func() {
		updatedSnapshot := snapshot.MeshNetworkingSnapshot{
			CurrentState: snapshot.Resources{
				MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
			},
		}

		validator := mock_snapshot.NewMockSnapshotValidator(ctrl)
		validator.EXPECT().
			Validate(updatedSnapshot).
			Return(true)

		updatedSnapshot.CurrentState.MeshServices = append(updatedSnapshot.CurrentState.MeshServices, meshService2)
		updatedSnapshot.Delta.Created.MeshServices = append(updatedSnapshot.Delta.Created.MeshServices, meshService1)
		validator.EXPECT().
			Validate(updatedSnapshot).
			Return(true)

		var capturedEventHandler *controller.MeshServiceEventHandlerFuncs
		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = eventHandler
				return nil
			})

		meshGroupController := mock_zephyr_networking.NewMockMeshGroupController(ctrl)
		meshGroupController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadController := mock_zephyr_discovery.NewMockMeshWorkloadController(ctrl)
		meshWorkloadController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator := snapshot.NewNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)

		didReceiveSnapshot := false
		var boolMutex sync.Mutex
		generator.RegisterListener(func(ctx context.Context, networkingSnapshot snapshot.MeshNetworkingSnapshot) {
			Expect(didReceiveSnapshot).To(Equal(false), "Should only receive one snapshot in this test")
			boolMutex.Lock()
			defer boolMutex.Unlock()
			didReceiveSnapshot = true

			Expect(networkingSnapshot.CurrentState.MeshServices).To(HaveLen(2))
			Expect(networkingSnapshot.CurrentState.MeshServices[0]).To(Equal(meshService1))
			Expect(networkingSnapshot.CurrentState.MeshServices[1]).To(Equal(meshService2))
			Expect(networkingSnapshot.Delta.Created.MeshServices).To(HaveLen(2))
			Expect(networkingSnapshot.Delta.Created.MeshServices[0]).To(Equal(meshService1))
			Expect(networkingSnapshot.Delta.Created.MeshServices[1]).To(Equal(meshService2))
		})

		generator.StartPushingSnapshots(ctx, time.Millisecond*500)

		capturedEventHandler.OnCreate(meshService1)
		capturedEventHandler.OnCreate(meshService2)

		Eventually(func() bool {
			boolMutex.Lock()
			defer boolMutex.Unlock()
			return didReceiveSnapshot
		}, eventuallyTimeout, pollFrequency).Should(BeTrue(), "Should eventually receive a snapshot")
	})

	It("can accurately swap out updated resources from the current state of the world", func() {
		updatedSnapshot := snapshot.MeshNetworkingSnapshot{
			CurrentState: snapshot.Resources{
				MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
			},
		}

		validator := mock_snapshot.NewMockSnapshotValidator(ctrl)
		validator.EXPECT().
			Validate(updatedSnapshot).
			Return(true)

		updatedSnapshot.CurrentState.MeshServices = append(updatedSnapshot.CurrentState.MeshServices, meshService2)
		updatedSnapshot.Delta.Created.MeshServices = append(updatedSnapshot.Delta.Created.MeshServices, meshService1)
		validator.EXPECT().
			Validate(updatedSnapshot).
			Return(true)

		var capturedEventHandler *controller.MeshServiceEventHandlerFuncs
		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = eventHandler
				return nil
			})

		meshGroupController := mock_zephyr_networking.NewMockMeshGroupController(ctrl)
		meshGroupController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadController := mock_zephyr_discovery.NewMockMeshWorkloadController(ctrl)
		meshWorkloadController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator := snapshot.NewNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)

		updatedService := &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "updated-name",
				Namespace: env.DefaultWriteNamespace,
			},
		}

		numSnapshot := 0
		var counterMutex sync.Mutex
		generator.RegisterListener(func(ctx context.Context, networkingSnapshot snapshot.MeshNetworkingSnapshot) {
			counterMutex.Lock()
			defer counterMutex.Unlock()
			numSnapshot += 1

			if numSnapshot == 1 {
				Expect(networkingSnapshot.CurrentState.MeshServices).To(HaveLen(2))
				Expect(networkingSnapshot.CurrentState.MeshServices[0]).To(Equal(meshService1))
				Expect(networkingSnapshot.CurrentState.MeshServices[1]).To(Equal(meshService2))
				Expect(networkingSnapshot.Delta.Created.MeshServices).To(HaveLen(2))
				Expect(networkingSnapshot.Delta.Created.MeshServices[0]).To(Equal(meshService1))
				Expect(networkingSnapshot.Delta.Created.MeshServices[1]).To(Equal(meshService2))
			} else if numSnapshot == 2 {
				Expect(networkingSnapshot.CurrentState.MeshServices).To(HaveLen(2))
				Expect(networkingSnapshot.CurrentState.MeshServices[0]).To(Equal(meshService1))
				Expect(networkingSnapshot.CurrentState.MeshServices[1]).To(Equal(updatedService))
				Expect(networkingSnapshot.Delta.Created.MeshServices).To(HaveLen(0))
				Expect(networkingSnapshot.Delta.Updated.MeshServices).To(HaveLen(1))
				Expect(networkingSnapshot.Delta.Updated.MeshServices[0].New).To(Equal(updatedService))
				Expect(networkingSnapshot.Delta.Updated.MeshServices[0].Old).To(Equal(meshService2))
			} else {
				Fail("Should not receive more than two snapshots")
			}
		})

		generator.StartPushingSnapshots(ctx, time.Second)

		capturedEventHandler.OnCreate(meshService1)
		capturedEventHandler.OnCreate(meshService2)

		Eventually(func() int {
			counterMutex.Lock()
			defer counterMutex.Unlock()
			return numSnapshot
		}, eventuallyTimeout*2, pollFrequency).Should(Equal(1), "Should eventually receive a first snapshot")

		updatedSnapshot.CurrentState.MeshServices[1] = updatedService
		updatedSnapshot.Delta = snapshot.Delta{}
		validator.EXPECT().
			Validate(updatedSnapshot).
			Return(true)

		capturedEventHandler.OnUpdate(meshService2, updatedService)

		Eventually(func() int {
			counterMutex.Lock()
			defer counterMutex.Unlock()
			return numSnapshot
		}, eventuallyTimeout*2, pollFrequency).Should(Equal(2), "Should eventually receive a second snapshot")
	})
})
