package translation

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/output"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/internal/mocks"
	mock_mesh "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/mocks"
	mock_meshservice "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshservice/mocks"
	mock_meshworkload "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/meshworkload/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {

	var (
		ctl *gomock.Controller
		ctx context.Context

		mockDependencyFactory      *MockDependencyFactory
		mockMeshTranslator         *mock_mesh.MockTranslator
		mockMeshworkloadTranslator *mock_meshworkload.MockTranslator
		mockMeshserviceTranslator  *mock_meshservice.MockTranslator
	)

	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDependencyFactory = NewMockDependencyFactory(ctl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctl)
		mockMeshworkloadTranslator = mock_meshworkload.NewMockTranslator(ctl)
		mockMeshserviceTranslator = mock_meshservice.NewMockTranslator(ctl)
	})

	AfterEach(func() {
		ctl.Finish()
	})
	It("translates", func() {
		t := NewTranslator().(*translator)
		t.dependencies = mockDependencyFactory

		configMaps := corev1sets.NewConfigMapSet(&corev1.ConfigMap{})
		services := corev1sets.NewServiceSet(&corev1.Service{})
		pods := corev1sets.NewPodSet(&corev1.Pod{})
		nodes := corev1sets.NewNodeSet(&corev1.Node{})
		deployments := appsv1sets.NewDeploymentSet(&appsv1.Deployment{})
		replicaSets := appsv1sets.NewReplicaSetSet(&appsv1.ReplicaSet{})
		daemonSets := appsv1sets.NewDaemonSetSet(&appsv1.DaemonSet{})
		statefulSets := appsv1sets.NewStatefulSetSet(&appsv1.StatefulSet{})

		in := input.NewSnapshot(
			"mesh-discovery",
			configMaps,
			services,
			pods,
			nodes,
			deployments,
			replicaSets,
			daemonSets,
			statefulSets,
		)

		mockDependencyFactory.EXPECT().MakeMeshTranslator(ctx, in).Return(mockMeshTranslator)
		mockDependencyFactory.EXPECT().MakeMeshWorkloadTranslator(ctx, in).Return(mockMeshworkloadTranslator)
		mockDependencyFactory.EXPECT().MakeMeshServiceTranslator(ctx).Return(mockMeshserviceTranslator)

		labeledMeta := metav1.ObjectMeta{Labels: labelutils.ClusterLabels("cluster")}

		meshes := v1alpha2sets.NewMeshSet(&v1alpha2.Mesh{ObjectMeta: labeledMeta})
		meshWorkloads := v1alpha2sets.NewMeshWorkloadSet(&v1alpha2.MeshWorkload{ObjectMeta: labeledMeta})
		meshServices := v1alpha2sets.NewMeshServiceSet(&v1alpha2.MeshService{ObjectMeta: labeledMeta})

		mockMeshTranslator.EXPECT().TranslateMeshes(deployments).Return(meshes)
		mockMeshworkloadTranslator.EXPECT().TranslateMeshWorkloads(deployments, daemonSets, statefulSets, meshes).Return(meshWorkloads)
		mockMeshserviceTranslator.EXPECT().TranslateMeshServices(services, meshWorkloads).Return(meshServices)

		out, err := t.Translate(ctx, in)
		Expect(err).NotTo(HaveOccurred())

		expectedOut, err := output.NewSinglePartitionedSnapshot(
			"mesh-discovery-1",
			labelutils.OwnershipLabels(),
			meshServices,
			meshWorkloads,
			meshes,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(out).To(Equal(expectedOut))
	})
})
