package cert_manager_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster/mocks"
	cert_manager "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/security/cert-manager"
	mock_cert_manager "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/security/cert-manager/mocks"
	mock_vm_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/validation/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_security "github.com/solo-io/service-mesh-hub/test/mocks/clients/security.smh.solo.io/v1alpha1"
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
		dynamicClientGetter *mock_multicluster.MockDynamicClientGetter
		csrClient           *mock_smh_security.MockVirtualMeshCertificateSigningRequestClient
		certConfigProducer  *mock_cert_manager.MockCertConfigProducer

		mockCsrClientFactory smh_security.VirtualMeshCertificateSigningRequestClientFactory = func(
			client client.Client) smh_security.VirtualMeshCertificateSigningRequestClient {
			return csrClient
		}
		testErr     = eris.New("hello")
		clusterName = "cluster-name"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		meshClient = mock_core.NewMockMeshClient(ctrl)
		meshRefFinder = mock_vm_validation.NewMockVirtualMeshFinder(ctrl)
		dynamicClientGetter = mock_multicluster.NewMockDynamicClientGetter(ctrl)
		certConfigProducer = mock_cert_manager.NewMockCertConfigProducer(ctrl)
		csrClient = mock_smh_security.NewMockVirtualMeshCertificateSigningRequestClient(ctrl)
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
			vm := &smh_networking.VirtualMesh{
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{},
				},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(smh_networking_types.VirtualMeshStatus{
				CertificateStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return an error if mesh is not type istio", func() {
			vm := &smh_networking.VirtualMesh{
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{},
				},
			}

			mesh := &smh_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
				},
				Status: smh_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*smh_discovery.Mesh{mesh}, nil)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(smh_networking_types.VirtualMeshStatus{
				CertificateStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: cert_manager.UnsupportedMeshTypeError(mesh).Error(),
				},
			}))
		})

		It("will return an error if cert config fails", func() {
			vm := &smh_networking.VirtualMesh{
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{},
				},
			}

			mesh := &smh_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{},
					},
				},
				Status: smh_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*smh_discovery.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(smh_networking_types.VirtualMeshStatus{
				CertificateStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: cert_manager.UnableToGatherCertConfigInfo(testErr, mesh, vm).Error(),
				},
			}))
		})

		It("will return an error multicluster clientset cannot be found", func() {
			vm := &smh_networking.VirtualMesh{
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{},
				},
			}

			mesh := &smh_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{},
					},
				},
				Status: smh_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*smh_discovery.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(nil, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, mc_manager.ClientNotFoundError(clusterName))

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(smh_networking_types.VirtualMeshStatus{
				CertificateStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: mc_manager.ClientNotFoundError(clusterName).Error(),
				},
			}))
		})

		It("will return an error multicluster clientset cannot be found", func() {
			vm := &smh_networking.VirtualMesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{},
				},
			}

			mesh := &smh_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{},
					},
				},
				Status: smh_discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*smh_discovery.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(nil, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, nil)

			csrClient.EXPECT().
				GetVirtualMeshCertificateSigningRequest(ctx, client.ObjectKey{Name: "istio1-5-name-cert-request", Namespace: container_runtime.GetWriteNamespace()}).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(smh_networking_types.VirtualMeshStatus{
				CertificateStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return if csr creation fails", func() {
			vm := &smh_networking.VirtualMesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{},
				},
			}

			mesh := &smh_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
						Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{},
					},
				},
				Status: smh_discovery_types.MeshStatus{},
			}

			certConfig := &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
				Hosts:    []string{"hello", "world"},
				Org:      "test",
				MeshType: smh_core_types.MeshType_ISTIO1_5,
			}

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(certConfig, nil)

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*smh_discovery.Mesh{mesh}, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, nil)
			statusErr := errors.NewNotFound(schema.GroupResource{}, "")
			csrClient.EXPECT().
				GetVirtualMeshCertificateSigningRequest(ctx, client.ObjectKey{Name: "istio1-5-name-cert-request", Namespace: container_runtime.GetWriteNamespace()}).
				Return(nil, statusErr)

			csrClient.EXPECT().
				CreateVirtualMeshCertificateSigningRequest(ctx, &smh_security.VirtualMeshCertificateSigningRequest{
					ObjectMeta: k8s_meta.ObjectMeta{
						Name:      "istio1-5-name-cert-request",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
						VirtualMeshRef: &smh_core_types.ResourceRef{
							Name:      vm.GetName(),
							Namespace: vm.GetNamespace(),
						},
						CertConfig: certConfig,
					},
					Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
						ComputedStatus: &smh_core_types.Status{
							State:   smh_core_types.Status_UNKNOWN,
							Message: "awaiting automated csr generation",
						},
					},
				}).
				Return(testErr)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status).To(Equal(smh_networking_types.VirtualMeshStatus{
				CertificateStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return nil if csr creation passes", func() {
			vm := &smh_networking.VirtualMesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: smh_networking_types.VirtualMeshSpec{
					Meshes: []*smh_core_types.ResourceRef{},
				},
			}

			mesh := &smh_discovery.Mesh{
				ObjectMeta: k8s_meta.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &smh_discovery_types.MeshSpec_Istio1_6_{
						Istio1_6: &smh_discovery_types.MeshSpec_Istio1_6{},
					},
				},
				Status: smh_discovery_types.MeshStatus{},
			}

			certConfig := &smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig{
				Hosts:    []string{"hello", "world"},
				Org:      "test",
				MeshType: smh_core_types.MeshType_ISTIO1_6,
			}

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(vm, mesh).
				Return(certConfig, nil)

			meshRefFinder.EXPECT().
				GetMeshesForVirtualMesh(ctx, vm).
				Return([]*smh_discovery.Mesh{mesh}, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName, gomock.Any()).
				Return(nil, nil)
			statusErr := errors.NewNotFound(schema.GroupResource{}, "")
			csrClient.EXPECT().
				GetVirtualMeshCertificateSigningRequest(ctx, client.ObjectKey{Name: "istio1-6-name-cert-request", Namespace: container_runtime.GetWriteNamespace()}).
				Return(nil, statusErr)

			csrClient.EXPECT().
				CreateVirtualMeshCertificateSigningRequest(ctx, &smh_security.VirtualMeshCertificateSigningRequest{
					ObjectMeta: k8s_meta.ObjectMeta{
						Name:      "istio1-6-name-cert-request",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
						VirtualMeshRef: &smh_core_types.ResourceRef{
							Name:      vm.GetName(),
							Namespace: vm.GetNamespace(),
						},
						CertConfig: certConfig,
					},
					Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
						ComputedStatus: &smh_core_types.Status{
							Message: "awaiting automated csr generation",
						},
					},
				}).
				Return(nil)

			status := csrProcessor.InitializeCertificateForVirtualMesh(ctx, vm)
			Expect(status.CertificateStatus).To(Equal(&smh_core_types.Status{
				State: smh_core_types.Status_ACCEPTED,
			}))
		})

	})

})
