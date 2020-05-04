package cert_manager_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s/mocks"
	cert_manager "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-manager"
	mock_cert_manager "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-manager/mocks"
	mock_vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_security "github.com/solo-io/service-mesh-hub/test/mocks/clients/security.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("csr manager", func() {

	var (
		ctx                 context.Context
		ctrl                *gomock.Controller
		csrProcessor        cert_manager.VirtualMeshCertificateManager
		meshClient          *mock_core.MockMeshClient
		meshRefFinder       *mock_vm_validation.MockVirtualMeshFinder
		dynamicClientGetter *mock_mc_manager.MockDynamicClientGetter
		csrClient           *mock_zephyr_security.MockVirtualMeshCertificateSigningRequestClient
		certConfigProducer  *mock_cert_manager.MockCertConfigProducer

		mockCsrClientFactory zephyr_security.VirtualMeshCertificateSigningRequestClientFactory = func(
			client client.Client) zephyr_security.VirtualMeshCertificateSigningRequestClient {
			return csrClient
		}
		testErr     = eris.New("hello")
		clusterName = "cluster-name"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		meshClient = mock_core.NewMockMeshClient(ctrl)
		meshRefFinder = mock_vm_validation.NewMockVirtualMeshFinder(ctrl)
		dynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		certConfigProducer = mock_cert_manager.NewMockCertConfigProducer(ctrl)
		csrClient = mock_zephyr_security.NewMockVirtualMeshCertificateSigningRequestClient(ctrl)
		csrProcessor = cert_manager.NewVirtualMeshCsrProcessor(
			dynamicClientGetter,
			meshClient,
			meshRefFinder,
			mockCsrClientFactory,
			certConfigProducer,
		)
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("create", func() {

		It("will return an error if mesh finder fails", func() {
			vm := &zephyr_networking.VirtualMesh{
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{},
				},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(zephyr_networking_types.VirtualMeshStatus{
				CertificateStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return an error if mesh is not type istio", func() {
			vm := &zephyr_networking.VirtualMesh{
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{},
				},
			}

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
				},
				Status: zephyr_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*zephyr_discovery.Mesh{mesh}, nil)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(zephyr_networking_types.VirtualMeshStatus{
				CertificateStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
					Message: cert_manager.UnsupportedMeshTypeError(mesh).Error(),
				},
			}))
		})

		It("will return an error if cert config fails", func() {
			vm := &zephyr_networking.VirtualMesh{
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{},
				},
			}

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
					},
				},
				Status: zephyr_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*zephyr_discovery.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(zephyr_networking_types.VirtualMeshStatus{
				CertificateStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
					Message: cert_manager.UnableToGatherCertConfigInfo(testErr, mesh, vm).Error(),
				},
			}))
		})

		It("will return an error multicluster clientset cannot be found", func() {
			vm := &zephyr_networking.VirtualMesh{
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{},
				},
			}

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
					},
				},
				Status: zephyr_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*zephyr_discovery.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(nil, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, mc_manager.ClientNotFoundError(clusterName))

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(zephyr_networking_types.VirtualMeshStatus{
				CertificateStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
					Message: mc_manager.ClientNotFoundError(clusterName).Error(),
				},
			}))
		})

		It("will return an error multicluster clientset cannot be found", func() {
			vm := &zephyr_networking.VirtualMesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{},
				},
			}

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
					},
				},
				Status: zephyr_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*zephyr_discovery.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(nil, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, nil)

			csrClient.EXPECT().
				GetVirtualMeshCertificateSigningRequest(ctx, client.ObjectKey{Name: "istio-name-cert-request", Namespace: env.GetWriteNamespace()}).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(zephyr_networking_types.VirtualMeshStatus{
				CertificateStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return if csr creation fails", func() {
			vm := &zephyr_networking.VirtualMesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{},
				},
			}

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
					},
				},
				Status: zephyr_discovery_types.MeshStatus{},
			}

			certConfig := &zephyr_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
				Hosts:    []string{"hello", "world"},
				Org:      "test",
				MeshType: zephyr_core_types.MeshType_ISTIO,
			}

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(certConfig, nil)

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*zephyr_discovery.Mesh{mesh}, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, nil)
			statusErr := errors.NewNotFound(schema.GroupResource{}, "")
			csrClient.EXPECT().
				GetVirtualMeshCertificateSigningRequest(ctx, client.ObjectKey{Name: "istio-name-cert-request", Namespace: env.GetWriteNamespace()}).
				Return(nil, statusErr)

			csrClient.EXPECT().
				CreateVirtualMeshCertificateSigningRequest(ctx, &zephyr_security.VirtualMeshCertificateSigningRequest{
					ObjectMeta: k8s_meta.ObjectMeta{
						Name:      "istio-name-cert-request",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_security_types.VirtualMeshCertificateSigningRequestSpec{
						VirtualMeshRef: &zephyr_core_types.ResourceRef{
							Name:      vm.GetName(),
							Namespace: vm.GetNamespace(),
						},
						CertConfig: certConfig,
					},
					Status: zephyr_security_types.VirtualMeshCertificateSigningRequestStatus{
						ComputedStatus: &zephyr_core_types.Status{
							State:   zephyr_core_types.Status_UNKNOWN,
							Message: "awaiting automated csr generation",
						},
					},
				}).
				Return(testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(zephyr_networking_types.VirtualMeshStatus{
				CertificateStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return nil if csr creation passes", func() {
			vm := &zephyr_networking.VirtualMesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: zephyr_networking_types.VirtualMeshSpec{
					Meshes: []*zephyr_core_types.ResourceRef{},
				},
			}

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Istio{
						Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
					},
				},
				Status: zephyr_discovery_types.MeshStatus{},
			}

			certConfig := &zephyr_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
				Hosts:    []string{"hello", "world"},
				Org:      "test",
				MeshType: zephyr_core_types.MeshType_ISTIO,
			}

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(certConfig, nil)

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*zephyr_discovery.Mesh{mesh}, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, nil)
			statusErr := errors.NewNotFound(schema.GroupResource{}, "")
			csrClient.EXPECT().
				GetVirtualMeshCertificateSigningRequest(ctx, client.ObjectKey{Name: "istio-name-cert-request", Namespace: env.GetWriteNamespace()}).
				Return(nil, statusErr)

			csrClient.EXPECT().
				CreateVirtualMeshCertificateSigningRequest(ctx, &zephyr_security.VirtualMeshCertificateSigningRequest{
					ObjectMeta: k8s_meta.ObjectMeta{
						Name:      "istio-name-cert-request",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_security_types.VirtualMeshCertificateSigningRequestSpec{
						VirtualMeshRef: &zephyr_core_types.ResourceRef{
							Name:      vm.GetName(),
							Namespace: vm.GetNamespace(),
						},
						CertConfig: certConfig,
					},
					Status: zephyr_security_types.VirtualMeshCertificateSigningRequestStatus{
						ComputedStatus: &zephyr_core_types.Status{
							Message: "awaiting automated csr generation",
						},
					},
				}).
				Return(nil)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status.CertificateStatus).To(Equal(&zephyr_core_types.Status{
				State: zephyr_core_types.Status_ACCEPTED,
			}))
		})

	})

})
