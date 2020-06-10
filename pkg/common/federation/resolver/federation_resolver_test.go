package resolver_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/dns"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/resolver"
	mock_meshes "github.com/solo-io/service-mesh-hub/pkg/common/federation/resolver/meshes/mock"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	mock_discovery_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/smh/discovery"
	"go.uber.org/zap/zapcore"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Federation Decider", func() {
	var (
		ctrl       *gomock.Controller
		ctx        context.Context
		testLogger *test_logging.TestLogger

		testErr = eris.New("hello")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		testLogger = test_logging.NewTestLogger()
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does nothing when a service's status is the only thing that changes", func() {
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshWorkloadClient := mock_discovery_core.NewMockMeshWorkloadClient(ctrl)
		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		virtualMeshClient := mock_smh_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs

		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			MeshServiceEventWatcher,
		).Start(ctx)

		oldMeshService := &smh_discovery.MeshService{
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name: "doesn't matter",
				},
			},
			Status: smh_discovery_types.MeshServiceStatus{
				FederationStatus: &smh_core_types.Status{
					State: smh_core_types.Status_ACCEPTED,
				},
			},
		}
		newMeshService := *oldMeshService
		newMeshService.Status = smh_discovery_types.MeshServiceStatus{
			FederationStatus: &smh_core_types.Status{
				State: smh_core_types.Status_INVALID,
			},
		}
		err := capturedEventHandler.UpdateMeshService(oldMeshService, &newMeshService)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does nothing when a service has no federation metadata yet", func() {
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshWorkloadClient := mock_discovery_core.NewMockMeshWorkloadClient(ctrl)
		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		virtualMeshClient := mock_smh_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs

		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			MeshServiceEventWatcher,
		).Start(ctx)

		service1 := &smh_discovery.MeshService{
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name: "doesn't matter",
				},
			},
			Status: smh_discovery_types.MeshServiceStatus{},
		}

		err := capturedEventHandler.CreateMeshService(service1)
		Expect(err).NotTo(HaveOccurred())
	})

	It("will add bad status, and log failures if any federation failes", func() {
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshWorkloadClient := mock_discovery_core.NewMockMeshWorkloadClient(ctrl)
		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		virtualMeshClient := mock_smh_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs

		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			MeshServiceEventWatcher,
		).Start(ctx)

		workload1 := &smh_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "workload-1",
				Namespace: "ns",
			},
			Spec: smh_discovery_types.MeshWorkloadSpec{},
		}

		workload2 := &smh_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "workload-2",
				Namespace: "ns",
			},
			Spec: smh_discovery_types.MeshWorkloadSpec{},
		}

		service1 := &smh_discovery.MeshService{
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name: "doesn't matter",
				},
				Federation: &smh_discovery_types.MeshServiceSpec_Federation{
					FederatedToWorkloads: []*smh_core_types.ResourceRef{
						{
							Name:      workload1.Name,
							Namespace: workload1.Namespace,
						},
						{
							Name:      workload2.Name,
							Namespace: workload2.Namespace,
						},
					},
				},
			},
			Status: smh_discovery_types.MeshServiceStatus{},
		}
		eventCtx := container_runtime.EventContext(ctx, container_runtime.CreateEvent, service1)
		meshWorkloadClient.EXPECT().
			GetMeshWorkload(eventCtx, selection.ResourceRefToObjectKey(service1.Spec.GetFederation().GetFederatedToWorkloads()[0])).
			Return(nil, testErr)

		meshWorkloadClient.EXPECT().
			GetMeshWorkload(eventCtx, selection.ResourceRefToObjectKey(service1.Spec.GetFederation().GetFederatedToWorkloads()[1])).
			Return(nil, testErr)

		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(
				eventCtx,
				&smh_discovery.MeshService{
					Spec: service1.Spec,
					Status: smh_discovery_types.MeshServiceStatus{
						FederationStatus: &smh_core_types.Status{
							State: smh_core_types.Status_PROCESSING_ERROR,
							Message: resolver.FailedToFederateServices(
								service1,
								service1.Spec.GetFederation().GetFederatedToWorkloads(),
							),
						},
					},
				},
			).
			Return(nil)

		err := capturedEventHandler.CreateMeshService(service1)
		Expect(err).NotTo(HaveOccurred())

		testLogger.EXPECT().
			NumEntries(4)
		testLogger.EXPECT().
			Entry(testLogger.NumLogEntries()-3).
			Level(zapcore.WarnLevel).
			HaveMessage(resolver.FailedToFederateServiceMessage).
			Have("mesh_workload", fmt.Sprintf("%s.%s", workload1.Name, workload1.Namespace)).
			Have("mesh_service", fmt.Sprintf("%s.%s", service1.Name, service1.Namespace))
		testLogger.EXPECT().
			Entry(testLogger.NumLogEntries()-2).
			HaveMessage(resolver.FailedToFederateServiceMessage).
			Level(zapcore.WarnLevel).
			Have("mesh_workload", fmt.Sprintf("%s.%s", workload2.Name, workload2.Namespace)).
			Have("mesh_service", fmt.Sprintf("%s.%s", service1.Name, service1.Namespace))
	})

	It("can federate Istio to Istio", func() {
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshWorkloadClient := mock_discovery_core.NewMockMeshWorkloadClient(ctrl)
		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		virtualMeshClient := mock_smh_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs

		MeshServiceEventWatcher := mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		MeshServiceEventWatcher.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			MeshServiceEventWatcher,
		).Start(ctx)

		federatedServiceRef := &smh_core_types.ResourceRef{
			Name:      "federated-service",
			Namespace: container_runtime.GetWriteNamespace(),
		}
		kubeServiceRef := &smh_core_types.ResourceRef{
			Name:      "test-svc",
			Namespace: "application-ns",
		}
		meshWorkloadRef := &smh_core_types.ResourceRef{
			Name:      "client-workload",
			Namespace: "client-ns",
		}
		serverClusterName := "server-cluster"
		clientClusterRef := &smh_core_types.ResourceRef{
			Name: "client-cluster",
		}
		clientMesh := &smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "client-mesh",
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.MeshSpec{
				Cluster: clientClusterRef,
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			},
		}
		serverMesh := &smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "server-mesh",
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
						Metadata: &smh_discovery_types.MeshSpec_IstioMesh{
							Installation: &smh_discovery_types.MeshSpec_MeshInstallation{
								InstallationNamespace: "istio-system",
							},
						},
					},
				},
			},
		}

		federatedService := &smh_discovery.MeshService{
			ObjectMeta: selection.ResourceRefToObjectMeta(federatedServiceRef),
			Spec: smh_discovery_types.MeshServiceSpec{
				Federation: &smh_discovery_types.MeshServiceSpec_Federation{
					MulticlusterDnsName:  dns.BuildMulticlusterDnsName(kubeServiceRef, serverClusterName),
					FederatedToWorkloads: []*smh_core_types.ResourceRef{meshWorkloadRef},
				},
				KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
					Ref: kubeServiceRef,
				},
				Mesh: selection.ObjectMetaToResourceRef(serverMesh.ObjectMeta),
			},
		}
		federatedToWorkload := &smh_discovery.MeshWorkload{
			Spec: smh_discovery_types.MeshWorkloadSpec{
				Mesh: selection.ObjectMetaToResourceRef(clientMesh.ObjectMeta),
			},
		}
		virtualMeshContainingService := &smh_networking.VirtualMesh{
			Spec: smh_networking_types.VirtualMeshSpec{
				Meshes: []*smh_core_types.ResourceRef{selection.ObjectMetaToResourceRef(serverMesh.ObjectMeta)},
			},
		}
		externalAddress := "255.255.255.255" // intentional garbage
		port := uint32(32000)                // equally intentional garbage

		eventCtx := container_runtime.EventContext(ctx, container_runtime.CreateEvent, federatedService)
		meshWorkloadClient.EXPECT().
			GetMeshWorkload(eventCtx, selection.ResourceRefToObjectKey(meshWorkloadRef)).
			Return(federatedToWorkload, nil)
		meshClient.EXPECT().
			GetMesh(eventCtx, selection.ResourceRefToObjectKey(selection.ObjectMetaToResourceRef(clientMesh.ObjectMeta))).
			Return(clientMesh, nil)
		meshClient.EXPECT().
			GetMesh(eventCtx, selection.ResourceRefToObjectKey(selection.ObjectMetaToResourceRef(serverMesh.ObjectMeta))).
			Return(serverMesh, nil)
		virtualMeshClient.EXPECT().
			ListVirtualMesh(eventCtx).
			Return(&smh_networking.VirtualMeshList{
				Items: []smh_networking.VirtualMesh{*virtualMeshContainingService},
			}, nil)
		eap := dns.ExternalAccessPoint{
			Address: externalAddress,
			Port:    port,
		}
		meshFederationClient.EXPECT().
			FederateServiceSide(contextutils.WithLogger(eventCtx, "istio"), "istio-system", virtualMeshContainingService, federatedService).
			Return(eap, nil)
		meshFederationClient.EXPECT().
			FederateClientSide(contextutils.WithLogger(eventCtx, "istio"), "istio-system", eap, federatedService, federatedToWorkload).
			Return(nil)
		serviceCopy := *federatedService
		serviceCopy.Status.FederationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(eventCtx, &serviceCopy).
			Return(nil)

		err := capturedEventHandler.OnCreate(federatedService)
		Expect(err).NotTo(HaveOccurred())
	})
})
