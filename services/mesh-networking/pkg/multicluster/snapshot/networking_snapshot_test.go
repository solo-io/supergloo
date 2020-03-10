package snapshot_test

import (
	"context"
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

		eventuallyTimeout = time.Second
		pollFrequency     = time.Millisecond

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
			MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
		}

		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, &updatedSnapshot).
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

		generator, err := snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)
		Expect(err).NotTo(HaveOccurred())

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), &updatedSnapshot).
			DoAndReturn(func(ctx context.Context, snap *snapshot.MeshNetworkingSnapshot) {
				didReceiveSnapshot <- struct{}{}
			})

		generator.RegisterListener(listener)

		go func() {
			defer GinkgoRecover()
			generator.StartPushingSnapshots(ctx, pollFrequency)
		}()

		Eventually(didReceiveSnapshot, eventuallyTimeout, pollFrequency).Should(Receive())
	})

	It("should not push snapshots if nothing has changed", func() {
		updatedSnapshot := snapshot.MeshNetworkingSnapshot{
			MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
		}

		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, &updatedSnapshot).
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

		generator, err := snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)
		Expect(err).NotTo(HaveOccurred())

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), &updatedSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *snapshot.MeshNetworkingSnapshot) {
				didReceiveSnapshot <- struct{}{}
			})
		generator.RegisterListener(listener)

		go func() {
			defer GinkgoRecover()
			generator.StartPushingSnapshots(ctx, pollFrequency)
		}()

		Eventually(
			didReceiveSnapshot,
			eventuallyTimeout,
			pollFrequency,
		).Should(Receive(), "should receive the first snapshot")
		Consistently(didReceiveSnapshot).ShouldNot(Receive(), "should not receive any further snapshots")
	})

	It("can aggregate multiple events that roll in close to each other", func() {
		originalSnapshot := &snapshot.MeshNetworkingSnapshot{
			MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
		}

		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, originalSnapshot).
			Return(true)

		updatedSnapshot := &snapshot.MeshNetworkingSnapshot{
			MeshServices: append(originalSnapshot.MeshServices, meshService2),
		}
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService2, updatedSnapshot).
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

		generator, err := snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)
		Expect(err).NotTo(HaveOccurred())

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), updatedSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *snapshot.MeshNetworkingSnapshot) {
				didReceiveSnapshot <- struct{}{}
			})

		generator.RegisterListener(listener)

		go func() {
			defer GinkgoRecover()
			generator.StartPushingSnapshots(ctx, pollFrequency)
		}()

		capturedEventHandler.OnCreate(meshService1)
		capturedEventHandler.OnCreate(meshService2)

		Eventually(
			didReceiveSnapshot,
			eventuallyTimeout,
			pollFrequency,
		).Should(Receive(), "Should eventually receive a snapshot")
	})

	It("can accurately swap out updated resources from the current state of the world", func() {
		originalSnapshot := &snapshot.MeshNetworkingSnapshot{
			MeshServices: []*discovery_v1alpha1.MeshService{meshService1},
		}
		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, originalSnapshot).
			Return(true)

		updatedSnapshot := &snapshot.MeshNetworkingSnapshot{
			MeshServices: append(originalSnapshot.MeshServices, meshService2),
		}
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService2, updatedSnapshot).
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

		generator, err := snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			meshServiceController,
			meshGroupController,
			meshWorkloadController,
		)
		Expect(err).NotTo(HaveOccurred())

		updatedService := &discovery_v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ms-2",
				Namespace: env.DefaultWriteNamespace,
			},
		}

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), updatedSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *snapshot.MeshNetworkingSnapshot) {
				didReceiveSnapshot <- struct{}{}
			}).Times(1)

		generator.RegisterListener(listener)

		newSnapshot := &snapshot.MeshNetworkingSnapshot{
			MeshServices: []*discovery_v1alpha1.MeshService{
				updatedSnapshot.MeshServices[0],
				updatedService,
			},
		}
		validator.EXPECT().
			ValidateMeshServiceUpsert(ctx, updatedService, newSnapshot).
			Return(true)

		listener.EXPECT().
			Sync(gomock.Any(), newSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *snapshot.MeshNetworkingSnapshot) {
				didReceiveSnapshot <- struct{}{}
			}).Times(1)

		go func() {
			defer GinkgoRecover()
			generator.StartPushingSnapshots(ctx, pollFrequency)
		}()

		capturedEventHandler.OnCreate(meshService1)
		capturedEventHandler.OnCreate(meshService2)

		Eventually(
			didReceiveSnapshot,
			eventuallyTimeout*2,
			pollFrequency,
		).Should(Receive(), "Should eventually receive a first snapshot")

		capturedEventHandler.OnUpdate(meshService2, updatedService)

		Eventually(
			didReceiveSnapshot,
			eventuallyTimeout*2,
			pollFrequency,
		).Should(Receive(), "Should eventually receive a second snapshot")
	})
})
