package networking_snapshot_test

import (
	"context"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	networking_snapshot "github.com/solo-io/service-mesh-hub/pkg/common/networking-snapshot"
	mock_snapshot "github.com/solo-io/service-mesh-hub/pkg/common/networking-snapshot/mocks"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/smh/discovery"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/smh/networking"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Networking Snapshot", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		eventuallyTimeout = time.Second
		pollFrequency     = time.Millisecond

		meshService1 = &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "ms-1",
				Namespace: container_runtime.GetWriteNamespace(),
			},
		}
		meshService2 = &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "ms-2",
				Namespace: container_runtime.GetWriteNamespace(),
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
		updatedSnapshot := networking_snapshot.MeshNetworkingSnapshot{
			MeshServices: []*smh_discovery.MeshService{meshService1},
		}

		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, &updatedSnapshot).
			Return(true)

		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				return eventHandler.OnCreate(meshService1)
			})

		virtualMeshEventWatcher := mock_smh_networking.NewMockVirtualMeshEventWatcher(ctrl)
		virtualMeshEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadEventWatcher := mock_smh_discovery.NewMockMeshWorkloadEventWatcher(ctrl)
		meshWorkloadEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator, err := networking_snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			MeshServiceEventWatcher,
			virtualMeshEventWatcher,
			meshWorkloadEventWatcher,
		)
		Expect(err).NotTo(HaveOccurred())

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), &updatedSnapshot).
			DoAndReturn(func(ctx context.Context, snap *networking_snapshot.MeshNetworkingSnapshot) {
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
		updatedSnapshot := networking_snapshot.MeshNetworkingSnapshot{
			MeshServices: []*smh_discovery.MeshService{meshService1},
		}

		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, &updatedSnapshot).
			Return(true)

		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				return eventHandler.OnCreate(meshService1)
			})

		virtualMeshEventWatcher := mock_smh_networking.NewMockVirtualMeshEventWatcher(ctrl)
		virtualMeshEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadEventWatcher := mock_smh_discovery.NewMockMeshWorkloadEventWatcher(ctrl)
		meshWorkloadEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator, err := networking_snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			MeshServiceEventWatcher,
			virtualMeshEventWatcher,
			meshWorkloadEventWatcher,
		)
		Expect(err).NotTo(HaveOccurred())

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), &updatedSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *networking_snapshot.MeshNetworkingSnapshot) {
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
		originalSnapshot := &networking_snapshot.MeshNetworkingSnapshot{
			MeshServices: []*smh_discovery.MeshService{meshService1},
		}

		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, originalSnapshot).
			Return(true)

		updatedSnapshot := &networking_snapshot.MeshNetworkingSnapshot{
			MeshServices: append(originalSnapshot.MeshServices, meshService2),
		}
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService2, updatedSnapshot).
			Return(true)

		var capturedEventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs
		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = eventHandler
				return nil
			})

		virtualMeshEventWatcher := mock_smh_networking.NewMockVirtualMeshEventWatcher(ctrl)
		virtualMeshEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadEventWatcher := mock_smh_discovery.NewMockMeshWorkloadEventWatcher(ctrl)
		meshWorkloadEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator, err := networking_snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			MeshServiceEventWatcher,
			virtualMeshEventWatcher,
			meshWorkloadEventWatcher,
		)
		Expect(err).NotTo(HaveOccurred())

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), updatedSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *networking_snapshot.MeshNetworkingSnapshot) {
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
		originalSnapshot := &networking_snapshot.MeshNetworkingSnapshot{
			MeshServices: []*smh_discovery.MeshService{meshService1},
		}
		validator := mock_snapshot.NewMockMeshNetworkingSnapshotValidator(ctrl)
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService1, originalSnapshot).
			Return(true)

		updatedSnapshot := &networking_snapshot.MeshNetworkingSnapshot{
			MeshServices: append(originalSnapshot.MeshServices, meshService2),
		}
		validator.EXPECT().
			ValidateMeshServiceUpsert(gomock.Any(), meshService2, updatedSnapshot).
			Return(true)

		var capturedEventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs
		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = eventHandler
				return nil
			})

		virtualMeshEventWatcher := mock_smh_networking.NewMockVirtualMeshEventWatcher(ctrl)
		virtualMeshEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		meshWorkloadEventWatcher := mock_smh_discovery.NewMockMeshWorkloadEventWatcher(ctrl)
		meshWorkloadEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			Return(nil)

		generator, err := networking_snapshot.NewMeshNetworkingSnapshotGenerator(
			ctx,
			validator,
			MeshServiceEventWatcher,
			virtualMeshEventWatcher,
			meshWorkloadEventWatcher,
		)
		Expect(err).NotTo(HaveOccurred())

		updatedService := &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "ms-2",
				Namespace: container_runtime.GetWriteNamespace(),
			},
		}

		didReceiveSnapshot := make(chan struct{})
		listener := mock_snapshot.NewMockMeshNetworkingSnapshotListener(ctrl)
		listener.EXPECT().
			Sync(gomock.Any(), updatedSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *networking_snapshot.MeshNetworkingSnapshot) {
				didReceiveSnapshot <- struct{}{}
			}).Times(1)

		generator.RegisterListener(listener)

		newSnapshot := &networking_snapshot.MeshNetworkingSnapshot{
			MeshServices: []*smh_discovery.MeshService{
				updatedSnapshot.MeshServices[0],
				updatedService,
			},
		}
		validator.EXPECT().
			ValidateMeshServiceUpsert(ctx, updatedService, newSnapshot).
			Return(true)

		listener.EXPECT().
			Sync(gomock.Any(), newSnapshot).
			DoAndReturn(func(ctx context.Context, networkingSnapshot *networking_snapshot.MeshNetworkingSnapshot) {
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
