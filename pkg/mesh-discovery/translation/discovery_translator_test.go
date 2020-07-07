package translation

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/snapshot/output"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	mock_mesh "github.com/solo-io/smh/pkg/mesh-discovery/translation/mesh/mocks"
	mock_meshservice "github.com/solo-io/smh/pkg/mesh-discovery/translation/meshservice/mocks"
	mock_meshworkload "github.com/solo-io/smh/pkg/mesh-discovery/translation/meshworkload/mocks"
	"github.com/solo-io/smh/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {

	var (
		ctl *gomock.Controller
		ctx context.Context

		mockDependencyFactory      *MockdependencyFactory
		mockMeshTranslator         *mock_mesh.MockTranslator
		mockMeshworkloadTranslator *mock_meshworkload.MockTranslator
		mockMeshserviceTranslator  *mock_meshservice.MockTranslator
	)

	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDependencyFactory = NewMockdependencyFactory(ctl)
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
		deployments := appsv1sets.NewDeploymentSet(&appsv1.Deployment{})
		replicaSets := appsv1sets.NewReplicaSetSet(&appsv1.ReplicaSet{})
		daemonSets := appsv1sets.NewDaemonSetSet(&appsv1.DaemonSet{})
		statefulSets := appsv1sets.NewStatefulSetSet(&appsv1.StatefulSet{})

		in := input.NewSnapshot(
			"mesh-discovery",
			configMaps,
			services,
			pods,
			deployments,
			replicaSets,
			daemonSets,
			statefulSets,
		)

		mockDependencyFactory.EXPECT().makeMeshTranslator(ctx, configMaps).Return(mockMeshTranslator)
		mockDependencyFactory.EXPECT().makeMeshWorkloadTranslator(ctx, pods, replicaSets).Return(mockMeshworkloadTranslator)
		mockDependencyFactory.EXPECT().makeMeshServiceTranslator().Return(mockMeshserviceTranslator)

		labeledMeta := metav1.ObjectMeta{Labels: labelutils.ClusterLabels("cluster")}

		meshes := v1alpha1sets.NewMeshSet(&v1alpha1.Mesh{ObjectMeta: labeledMeta})
		meshWorkloads := v1alpha1sets.NewMeshWorkloadSet(&v1alpha1.MeshWorkload{ObjectMeta: labeledMeta})
		meshServices := v1alpha1sets.NewMeshServiceSet(&v1alpha1.MeshService{ObjectMeta: labeledMeta})

		mockMeshTranslator.EXPECT().TranslateMeshes(deployments).Return(meshes)
		mockMeshworkloadTranslator.EXPECT().TranslateMeshWorkloads(deployments, daemonSets, statefulSets, meshes).Return(meshWorkloads)
		mockMeshserviceTranslator.EXPECT().TranslateMeshServices(services, meshWorkloads).Return(meshServices)

		out, err := t.Translate(ctx, in)
		Expect(err).NotTo(HaveOccurred())

		expectedOut, err := output.NewLabelPartitionedSnapshot(
			"mesh-discovery",
			labelutils.ClusterLabelKey,
			meshServices,
			meshWorkloads,
			meshes,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(out).To(Equal(expectedOut))
	})
})
