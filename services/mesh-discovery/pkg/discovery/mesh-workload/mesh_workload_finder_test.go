package mesh_workload_test

import (
	"context"
	"errors"

	pb_types "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/mesh-projects/services/common/cluster/core/v1/controller"
	mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-workload"
	mock_mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-workload/mocks"
	test_logging "github.com/solo-io/mesh-projects/test/logging"
	corev1 "k8s.io/api/core/v1"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MeshWorkloadFinder", func() {
	var (
		ctx                         context.Context
		mockLocalMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockLocalMeshClient         *mock_core.MockMeshClient
		mockMeshWorkloadScanner     *mock_mesh_workload.MockMeshWorkloadScanner
		clusterName                 = "clusterName"
		meshWorkloadFinder          controller.PodEventHandler
		pod                         = &corev1.Pod{}
		discoveredMeshWorkload      *discoveryv1alpha1.MeshWorkload
		testLogger                  = test_logging.NewTestLogger()
		notFoundErr                 = &k8s_errs.StatusError{
			ErrStatus: metav1.Status{
				Reason: metav1.StatusReasonNotFound,
			},
		}
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		mockLocalMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockLocalMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshWorkloadScanner = mock_mesh_workload.NewMockMeshWorkloadScanner(ctrl)

		meshWorkloadFinder = mesh_workload.NewMeshWorkloadFinder(
			ctx,
			clusterName,
			mockLocalMeshWorkloadClient,
			mockLocalMeshClient,
			[]mesh_workload.MeshWorkloadScanner{mockMeshWorkloadScanner})
		discoveredMeshWorkload = &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeControllerRef: &core_types.ResourceRef{
					Namespace: "controller-namespace",
				},
			},
		}
	})

	It("should create MeshWorkload if Istio injected workload found", func() {
		meshName := "meshName"
		meshNamespace := "meshNamespace"
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: meshName, Namespace: meshNamespace},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		meshSpec := &core_types.ResourceRef{
			Kind:      &pb_types.StringValue{Value: mesh.TypeMeta.Kind},
			Name:      mesh.Name,
			Namespace: meshNamespace,
			Cluster:   &pb_types.StringValue{Value: clusterName},
		}
		expectedMeshWorkload := *discoveredMeshWorkload
		expectedMeshWorkload.Spec.Mesh = meshSpec
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, nil)
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().Get(ctx, objKey).Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil)
		mockLocalMeshWorkloadClient.EXPECT().Create(ctx, discoveredMeshWorkload).Return(nil)
		err := meshWorkloadFinder.Create(pod)
		Expect(pod.ClusterName).To(Equal(clusterName))
		Expect(err).ToNot(HaveOccurred())
		Expect(discoveredMeshWorkload).To(Equal(&expectedMeshWorkload))
	})

	It("should return error if fatal error while scanning workloads", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, expectedErr)
		err := meshWorkloadFinder.Create(pod)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should log warning if non-fatal error while scanning workloads", func() {
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, expectedErr)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, nil)
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().Get(ctx, objKey).Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil)
		mockLocalMeshWorkloadClient.EXPECT().Create(ctx, discoveredMeshWorkload).Return(nil)
		_ = meshWorkloadFinder.Create(pod)
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingNonFatal))
	})

	It("should return error if fatal error while populating Mesh resource ref", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, nil)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(nil, expectedErr)
		err := meshWorkloadFinder.Create(pod)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingError)
	})

	It("should create new MeshWorkload if Istio injected workload updated", func() {
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeControllerRef: &core_types.ResourceRef{
					Namespace: "controller-namespace",
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(newDiscoveredMeshWorkload, nil)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		err := meshWorkloadFinder.Update(pod, newPod)
		Expect(pod.ClusterName).To(Equal(clusterName))
		Expect(newPod.ClusterName).To(Equal(clusterName))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should do nothing if MeshWorkload unchanged", func() {
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeControllerRef: &core_types.ResourceRef{
					Namespace: "controller-namespace",
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(newDiscoveredMeshWorkload, nil)
		objKey, _ := client.ObjectKeyFromObject(newDiscoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().Get(ctx, objKey).Return(nil, notFoundErr)
		mockLocalMeshWorkloadClient.EXPECT().Create(ctx, newDiscoveredMeshWorkload).Return(nil)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		err := meshWorkloadFinder.Update(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should update MeshWorkload if changed", func() {
		newNamespace := "new-controller-namespace"
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeControllerRef: &core_types.ResourceRef{
					Namespace: newNamespace,
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(newDiscoveredMeshWorkload, nil)
		objKey, _ := client.ObjectKeyFromObject(newDiscoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().Get(ctx, objKey).Return(nil, notFoundErr)
		mockLocalMeshWorkloadClient.EXPECT().Create(ctx, newDiscoveredMeshWorkload).Return(nil)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		mockLocalMeshWorkloadClient.EXPECT().Update(ctx, newDiscoveredMeshWorkload).Return(nil)
		err := meshWorkloadFinder.Update(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error if error processing old pod", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, expectedErr)
		err := meshWorkloadFinder.Update(pod, &corev1.Pod{})
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingError)
	})

	It("should return error if error processing new pod", func() {
		newPod := &corev1.Pod{}
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(nil, expectedErr)
		err := meshWorkloadFinder.Update(pod, newPod)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingError)
	})

	It("should return nil if old and new Pods are not mesh injected", func() {
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, nil).Times(2)
		err := meshWorkloadFinder.Update(pod, &corev1.Pod{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create new MeshWorkload if pod is now Istio injected", func() {
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(discoveredMeshWorkload, nil)
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().Get(ctx, objKey).Return(nil, notFoundErr)
		mockLocalMeshWorkloadClient.EXPECT().Create(ctx, discoveredMeshWorkload).Return(nil)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		err := meshWorkloadFinder.Update(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should log warnings for non-fatal errors on update", func() {
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeControllerRef: &core_types.ResourceRef{
					Namespace: "controller-namespace",
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(discoveredMeshWorkload, errors.New(""))
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(newDiscoveredMeshWorkload, errors.New(""))
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		err := meshWorkloadFinder.Update(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingNonFatal)
	})
})
