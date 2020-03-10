package cert_manager_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	mock_zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security/mocks"
	"github.com/solo-io/mesh-projects/pkg/env"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	cert_manager "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-manager"
	mock_cert_manager "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-manager/mocks"
	mock_group_validation "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/validation/mocks"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("csr manager", func() {

	var (
		ctx                 context.Context
		ctrl                *gomock.Controller
		csrProcessor        cert_manager.MeshGroupCertificateManager
		meshClient          *mock_core.MockMeshClient
		meshRefFinder       *mock_group_validation.MockGroupMeshFinder
		dynamicClientGetter *mock_mc_manager.MockDynamicClientGetter
		csrClient           *mock_zephyr_security.MockMeshGroupCSRClient
		certConfigProducer  *mock_cert_manager.MockCertConfigProducer

		mockCsrClientFactory zephyr_security.MeshGroupCSRClientFactory = func(
			client client.Client) zephyr_security.MeshGroupCSRClient {
			return csrClient
		}
		testErr     = eris.New("hello")
		clusterName = "cluster-name"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		meshClient = mock_core.NewMockMeshClient(ctrl)
		meshRefFinder = mock_group_validation.NewMockGroupMeshFinder(ctrl)
		dynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		certConfigProducer = mock_cert_manager.NewMockCertConfigProducer(ctrl)
		csrClient = mock_zephyr_security.NewMockMeshGroupCSRClient(ctrl)
		csrProcessor = cert_manager.NewMeshGroupCsrProcessor(
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
			mg := &networking_v1alpha1.MeshGroup{
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{},
				},
			}

			meshRefFinder.EXPECT().
				GetMeshesForGroup(ctx, mg).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForMeshGroup(ctx, mg)
			Expect(status).To(Equal(networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return an error if mesh is not type istio", func() {
			mg := &networking_v1alpha1.MeshGroup{
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{},
				},
			}

			mesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
				},
				Status: discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForGroup(ctx, mg).
				Return([]*discovery_v1alpha1.Mesh{mesh}, nil)

			status := csrProcessor.InitializeCertificateForMeshGroup(ctx, mg)
			Expect(status).To(Equal(networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: cert_manager.UnsupportedMeshTypeError(mesh).Error(),
				},
			}))
		})

		It("will return an error if cert config fails", func() {
			mg := &networking_v1alpha1.MeshGroup{
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{},
				},
			}

			mesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.IstioMesh{},
					},
				},
				Status: discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForGroup(ctx, mg).
				Return([]*discovery_v1alpha1.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(mg, mesh).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForMeshGroup(ctx, mg)
			Expect(status).To(Equal(networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: cert_manager.UnableToGatherCertConfigInfo(testErr, mesh, mg).Error(),
				},
			}))
		})

		It("will return an error multicluster clientset cannot be found", func() {
			mg := &networking_v1alpha1.MeshGroup{
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{},
				},
			}

			mesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.IstioMesh{},
					},
				},
				Status: discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForGroup(ctx, mg).
				Return([]*discovery_v1alpha1.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(mg, mesh).
				Return(nil, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(clusterName, gomock.Any()).
				Return(nil, false)

			status := csrProcessor.InitializeCertificateForMeshGroup(ctx, mg)
			Expect(status).To(Equal(networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: cert_manager.DynamicClientDoesNotExistForClusterError(clusterName).Error(),
				},
			}))
		})

		It("will return an error multicluster clientset cannot be found", func() {
			mg := &networking_v1alpha1.MeshGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{},
				},
			}

			mesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.IstioMesh{},
					},
				},
				Status: discovery_types.MeshStatus{},
			}

			meshRefFinder.EXPECT().
				GetMeshesForGroup(ctx, mg).
				Return([]*discovery_v1alpha1.Mesh{mesh}, nil)

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(mg, mesh).
				Return(nil, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(clusterName, gomock.Any()).
				Return(nil, true)

			csrClient.EXPECT().
				Get(ctx, "istio-name-cert-request", env.DefaultWriteNamespace).
				Return(nil, testErr)

			status := csrProcessor.InitializeCertificateForMeshGroup(ctx, mg)
			Expect(status).To(Equal(networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return if csr creation fails", func() {
			mg := &networking_v1alpha1.MeshGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{},
				},
			}

			mesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.IstioMesh{},
					},
				},
				Status: discovery_types.MeshStatus{},
			}

			certConfig := &security_types.CertConfig{
				Hosts:    []string{"hello", "world"},
				Org:      "test",
				MeshType: core_types.MeshType_ISTIO,
			}

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(mg, mesh).
				Return(certConfig, nil)

			meshRefFinder.EXPECT().
				GetMeshesForGroup(ctx, mg).
				Return([]*discovery_v1alpha1.Mesh{mesh}, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(clusterName, gomock.Any()).
				Return(nil, true)
			statusErr := errors.NewNotFound(schema.GroupResource{}, "")
			csrClient.EXPECT().
				Get(ctx, "istio-name-cert-request", env.DefaultWriteNamespace).
				Return(nil, statusErr)

			csrClient.EXPECT().
				Create(ctx, &v1alpha1.MeshGroupCertificateSigningRequest{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "istio-name-cert-request",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: security_types.MeshGroupCertificateSigningRequestSpec{
						MeshGroupRef: &core_types.ResourceRef{
							Name:      mg.GetName(),
							Namespace: mg.GetNamespace(),
						},
						CertConfig: certConfig,
					},
					Status: security_types.MeshGroupCertificateSigningRequestStatus{
						ComputedStatus: &core_types.ComputedStatus{
							Status:  0,
							Message: "awaiting automated csr generation",
						},
					},
				}).
				Return(testErr)

			status := csrProcessor.InitializeCertificateForMeshGroup(ctx, mg)
			Expect(status).To(Equal(networking_types.MeshGroupStatus{
				CertificateStatus: &core_types.ComputedStatus{
					Status:  core_types.ComputedStatus_INVALID,
					Message: testErr.Error(),
				},
			}))
		})

		It("will return nil if csr creation passes", func() {
			mg := &networking_v1alpha1.MeshGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: networking_types.MeshGroupSpec{
					Meshes: []*core_types.ResourceRef{},
				},
			}

			mesh := &discovery_v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace",
					Name:      "name",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &discovery_types.MeshSpec_Istio{
						Istio: &discovery_types.IstioMesh{},
					},
				},
				Status: discovery_types.MeshStatus{},
			}

			certConfig := &security_types.CertConfig{
				Hosts:    []string{"hello", "world"},
				Org:      "test",
				MeshType: core_types.MeshType_ISTIO,
			}

			certConfigProducer.EXPECT().
				ConfigureCertificateInfo(mg, mesh).
				Return(certConfig, nil)

			meshRefFinder.EXPECT().
				GetMeshesForGroup(ctx, mg).
				Return([]*discovery_v1alpha1.Mesh{mesh}, nil)

			dynamicClientGetter.EXPECT().
				GetClientForCluster(clusterName, gomock.Any()).
				Return(nil, true)
			statusErr := errors.NewNotFound(schema.GroupResource{}, "")
			csrClient.EXPECT().
				Get(ctx, "istio-name-cert-request", env.DefaultWriteNamespace).
				Return(nil, statusErr)

			csrClient.EXPECT().
				Create(ctx, &v1alpha1.MeshGroupCertificateSigningRequest{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "istio-name-cert-request",
						Namespace: env.DefaultWriteNamespace,
					},
					Spec: security_types.MeshGroupCertificateSigningRequestSpec{
						MeshGroupRef: &core_types.ResourceRef{
							Name:      mg.GetName(),
							Namespace: mg.GetNamespace(),
						},
						CertConfig: certConfig,
					},
					Status: security_types.MeshGroupCertificateSigningRequestStatus{
						ComputedStatus: &core_types.ComputedStatus{
							Message: "awaiting automated csr generation",
						},
					},
				}).
				Return(nil)

			status := csrProcessor.InitializeCertificateForMeshGroup(ctx, mg)
			Expect(status.CertificateStatus).To(Equal(&core_types.ComputedStatus{
				Status: core_types.ComputedStatus_ACCEPTED,
			}))
		})

	})

})
