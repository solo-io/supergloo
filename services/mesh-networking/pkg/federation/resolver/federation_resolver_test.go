package resolver_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/clients"
	mock_discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/dns"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/resolver"
	mock_meshes "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/resolver/meshes/mock"
	mock_zephyr_discovery "github.com/solo-io/mesh-projects/test/mocks/zephyr/discovery"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Federation Decider", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
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
		).Start(ctx, meshServiceController)

		oldMeshService := &discovery_v1alpha1.MeshService{
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name: "doesn't matter",
				},
			},
			Status: types.MeshServiceStatus{
				FederationStatus: &core_types.ComputedStatus{
					Status: core_types.ComputedStatus_ACCEPTED,
				},
			},
		}
		newMeshService := *oldMeshService
		newMeshService.Status = types.MeshServiceStatus{
			FederationStatus: &core_types.ComputedStatus{
				Status: core_types.ComputedStatus_INVALID,
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
		).Start(ctx, meshServiceController)

		service1 := &discovery_v1alpha1.MeshService{
			Spec: types.MeshServiceSpec{
				Mesh: &core_types.ResourceRef{
					Name: "doesn't matter",
				},
			},
			Status: types.MeshServiceStatus{},
		}

		err := capturedEventHandler.Create(service1)
		Expect(err).NotTo(HaveOccurred())
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
		).Start(ctx, meshServiceController)

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
			ObjectMeta: v1.ObjectMeta{
				Name:      "client-mesh",
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.MeshSpec{
				Cluster: clientClusterRef,
				MeshType: &types.MeshSpec_Istio{
					Istio: &types.IstioMesh{
						Installation: &types.MeshInstallation{
							InstallationNamespace: "istio-system",
						},
					},
				},
			},
		}
		serverMesh := &discovery_v1alpha1.Mesh{
			ObjectMeta: v1.ObjectMeta{
				Name:      "server-mesh",
				Namespace: env.DefaultWriteNamespace,
			},
			Spec: types.MeshSpec{
				MeshType: &types.MeshSpec_Istio{
					Istio: &types.IstioMesh{
						Installation: &types.MeshInstallation{
							InstallationNamespace: "istio-system",
						},
					},
				},
			},
		}

		federatedService := &discovery_v1alpha1.MeshService{
			ObjectMeta: clients.ResourceRefToObjectMeta(federatedServiceRef),
			Spec: types.MeshServiceSpec{
				Federation: &types.Federation{
					MulticlusterDnsName:  dns.BuildMulticlusterDnsName(kubeServiceRef, serverClusterName),
					FederatedToWorkloads: []*core_types.ResourceRef{meshWorkloadRef},
				},
				KubeService: &types.KubeService{
					Ref: kubeServiceRef,
				},
				Mesh: clients.ObjectMetaToResourceRef(serverMesh.ObjectMeta),
			},
		}
		federatedToWorkload := &discovery_v1alpha1.MeshWorkload{
			Spec: types.MeshWorkloadSpec{
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
		serviceCopy.Status.FederationStatus = &core_types.ComputedStatus{
			Status: core_types.ComputedStatus_ACCEPTED,
		}
		meshServiceClient.EXPECT().
			UpdateStatus(ctx, &serviceCopy).
			Return(nil)

		err := capturedEventHandler.OnCreate(federatedService)
		Expect(err).NotTo(HaveOccurred())
	})
})
