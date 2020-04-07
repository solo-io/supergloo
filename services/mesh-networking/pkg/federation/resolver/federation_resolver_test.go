package resolver_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	mock_discovery_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/dns"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/resolver"
	mock_meshes "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/resolver/meshes/mock"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		virtualMeshClient := mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *controller.MeshServiceEventHandlerFuncs

		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			meshServiceController,
		).Start(ctx)

		oldMeshService := &discovery_v1alpha1.MeshService{
			Spec: discovery_types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name: "doesn't matter",
				},
			},
			Status: discovery_types.MeshServiceStatus{
				FederationStatus: &core_types.Status{
					State: core_types.Status_ACCEPTED,
				},
			},
		}
		newMeshService := *oldMeshService
		newMeshService.Status = discovery_types.MeshServiceStatus{
			FederationStatus: &core_types.Status{
				State: core_types.Status_INVALID,
			},
		}
		err := capturedEventHandler.Update(oldMeshService, &newMeshService)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does nothing when a service has no federation metadata yet", func() {
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshWorkloadClient := mock_discovery_core.NewMockMeshWorkloadClient(ctrl)
		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		virtualMeshClient := mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *controller.MeshServiceEventHandlerFuncs

		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			meshServiceController,
		).Start(ctx)

		service1 := &discovery_v1alpha1.MeshService{
			Spec: discovery_types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name: "doesn't matter",
				},
			},
			Status: discovery_types.MeshServiceStatus{},
		}

		err := capturedEventHandler.Create(service1)
		Expect(err).NotTo(HaveOccurred())
	})

	It("will add bad status, and log failures if any federation failes", func() {
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshWorkloadClient := mock_discovery_core.NewMockMeshWorkloadClient(ctrl)
		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		virtualMeshClient := mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *controller.MeshServiceEventHandlerFuncs

		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			meshServiceController,
		).Start(ctx)

		workload1 := &discovery_v1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workload-1",
				Namespace: "ns",
			},
			Spec: discovery_types.MeshWorkloadSpec{},
		}

		workload2 := &discovery_v1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workload-2",
				Namespace: "ns",
			},
			Spec: discovery_types.MeshWorkloadSpec{},
		}

		service1 := &discovery_v1alpha1.MeshService{
			Spec: discovery_types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name: "doesn't matter",
				},
				Federation: &discovery_types.MeshServiceSpec_Federation{
					FederatedToWorkloads: []*core_types.ResourceRef{
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
			Status: discovery_types.MeshServiceStatus{},
		}

		meshWorkloadClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(service1.Spec.GetFederation().GetFederatedToWorkloads()[0])).
			Return(nil, testErr)

		meshWorkloadClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(service1.Spec.GetFederation().GetFederatedToWorkloads()[1])).
			Return(nil, testErr)

		meshServiceClient.EXPECT().
			UpdateStatus(
				ctx,
				&discovery_v1alpha1.MeshService{
					Spec: service1.Spec,
					Status: discovery_types.MeshServiceStatus{
						FederationStatus: &core_types.Status{
							State: core_types.Status_PROCESSING_ERROR,
							Message: resolver.FailedToFederateServices(
								service1,
								service1.Spec.GetFederation().GetFederatedToWorkloads(),
							),
						},
					},
				},
			).
			Return(nil)

		err := capturedEventHandler.Create(service1)
		Expect(err).NotTo(HaveOccurred())

		testLogger.EXPECT().
			NumEntries(2)
		testLogger.EXPECT().
			FirstEntry().
			Level(zapcore.WarnLevel).
			HaveMessage(resolver.FailedToFederateServiceMessage).
			Have("mesh_workload", fmt.Sprintf("%s.%s", workload1.Name, workload1.Namespace)).
			Have("mesh_service", fmt.Sprintf("%s.%s", service1.Name, service1.Namespace))
		testLogger.EXPECT().
			LastEntry().
			HaveMessage(resolver.FailedToFederateServiceMessage).
			Level(zapcore.WarnLevel).
			Have("mesh_workload", fmt.Sprintf("%s.%s", workload2.Name, workload2.Namespace)).
			Have("mesh_service", fmt.Sprintf("%s.%s", service1.Name, service1.Namespace))
	})

	It("can federate Istio to Istio", func() {
		meshClient := mock_discovery_core.NewMockMeshClient(ctrl)
		meshWorkloadClient := mock_discovery_core.NewMockMeshWorkloadClient(ctrl)
		meshServiceClient := mock_discovery_core.NewMockMeshServiceClient(ctrl)
		virtualMeshClient := mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		meshFederationClient := mock_meshes.NewMockMeshFederationClient(ctrl)

		federationClients := resolver.PerMeshFederationClients{
			Istio: meshFederationClient,
		}

		var capturedEventHandler *controller.MeshServiceEventHandlerFuncs

		meshServiceController := mock_zephyr_discovery.NewMockMeshServiceController(ctrl)
		meshServiceController.EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, funcs *controller.MeshServiceEventHandlerFuncs) error {
				capturedEventHandler = funcs
				return nil
			})

		resolver.NewFederationResolver(
			meshClient,
			meshWorkloadClient,
			meshServiceClient,
			virtualMeshClient,
			federationClients,
			meshServiceController,
		).Start(ctx)

		federatedServiceRef := &core_types.ResourceRef{
			Name:      "federated-service",
			Namespace: env.DefaultWriteNamespace,
		}
		kubeServiceRef := &core_types.ResourceRef{
			Name:      "test-svc",
			Namespace: "application-ns",
		}
		meshWorkloadRef := &core_types.ResourceRef{
			Name:      "client-workload",
			Namespace: "client-ns",
		}
		serverClusterName := "server-cluster"
		clientClusterRef := &core_types.ResourceRef{
			Name: "client-cluster",
		}
		clientMesh := &discovery_v1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "client-mesh",
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: discovery_types.MeshSpec{
				Cluster: clientClusterRef,
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						Installation: &discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: "istio-system",
						},
					},
				},
			},
		}
		serverMesh := &discovery_v1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "server-mesh",
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{
					Istio: &discovery_types.MeshSpec_IstioMesh{
						Installation: &discovery_types.MeshSpec_MeshInstallation{
							InstallationNamespace: "istio-system",
						},
					},
				},
			},
		}

		federatedService := &discovery_v1alpha1.MeshService{
			ObjectMeta: clients.ResourceRefToObjectMeta(federatedServiceRef),
			Spec: discovery_types.MeshServiceSpec{
				Federation: &discovery_types.MeshServiceSpec_Federation{
					MulticlusterDnsName:  dns.BuildMulticlusterDnsName(kubeServiceRef, serverClusterName),
					FederatedToWorkloads: []*core_types.ResourceRef{meshWorkloadRef},
				},
				KubeService: &discovery_types.MeshServiceSpec_KubeService{
					Ref: kubeServiceRef,
				},
				Mesh: clients.ObjectMetaToResourceRef(serverMesh.ObjectMeta),
			},
		}
		federatedToWorkload := &discovery_v1alpha1.MeshWorkload{
			Spec: discovery_types.MeshWorkloadSpec{
				Mesh: clients.ObjectMetaToResourceRef(clientMesh.ObjectMeta),
			},
		}
		virtualMeshContainingService := &networking_v1alpha1.VirtualMesh{
			Spec: types2.VirtualMeshSpec{
				Meshes: []*core_types.ResourceRef{clients.ObjectMetaToResourceRef(serverMesh.ObjectMeta)},
			},
		}
		externalAddress := "255.255.255.255" // intentional garbage
		port := uint32(32000)                // equally intentional garbage

		meshWorkloadClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(meshWorkloadRef)).
			Return(federatedToWorkload, nil)
		meshClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(clients.ObjectMetaToResourceRef(clientMesh.ObjectMeta))).
			Return(clientMesh, nil)
		meshClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(clients.ObjectMetaToResourceRef(serverMesh.ObjectMeta))).
			Return(serverMesh, nil)
		virtualMeshClient.EXPECT().
			List(ctx).
			Return(&networking_v1alpha1.VirtualMeshList{
				Items: []networking_v1alpha1.VirtualMesh{*virtualMeshContainingService},
			}, nil)
		eap := dns.ExternalAccessPoint{
			Address: externalAddress,
			Port:    port,
		}
		meshFederationClient.EXPECT().
			FederateServiceSide(ctx, virtualMeshContainingService, federatedService).
			Return(eap, nil)
		meshFederationClient.EXPECT().
			FederateClientSide(ctx, eap, federatedService, federatedToWorkload).
			Return(nil)
		serviceCopy := *federatedService
		serviceCopy.Status.FederationStatus = &core_types.Status{
			State: core_types.Status_ACCEPTED,
		}
		meshServiceClient.EXPECT().
			UpdateStatus(ctx, &serviceCopy).
			Return(nil)

		err := capturedEventHandler.OnCreate(federatedService)
		Expect(err).NotTo(HaveOccurred())
	})
})
