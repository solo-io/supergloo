package translation_test

import (
	"context"

	"github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	v1beta2sets "github.com/solo-io/external-apis/pkg/api/appmesh/appmesh.k8s.aws/v1beta2/sets"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/output/discovery"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation"
	mock_translator_internal "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/internal/mocks"
	mock_mesh "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/mocks"
	mock_traffictarget "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/traffictarget/mocks"
	mock_workload "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {

	var (
		ctl *gomock.Controller
		ctx context.Context

		mockDependencyFactory       *mock_translator_internal.MockDependencyFactory
		mockMeshTranslator          *mock_mesh.MockTranslator
		mockWorkloadTranslator      *mock_workload.MockTranslator
		mockTrafficTargetTranslator *mock_traffictarget.MockTranslator
	)

	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDependencyFactory = mock_translator_internal.NewMockDependencyFactory(ctl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctl)
		mockWorkloadTranslator = mock_workload.NewMockTranslator(ctl)
		mockTrafficTargetTranslator = mock_traffictarget.NewMockTranslator(ctl)
	})

	AfterEach(func() {
		ctl.Finish()
	})
	It("translates", func() {
		t := NewTranslator(mockDependencyFactory)

		appMeshes := v1beta2sets.NewMeshSet(&v1beta2.Mesh{})
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
			appMeshes,
			configMaps,
			services,
			pods,
			nodes,
			deployments,
			replicaSets,
			daemonSets,
			statefulSets,
		)

		mockDependencyFactory.EXPECT().MakeMeshTranslator(ctx).Return(mockMeshTranslator)
		mockDependencyFactory.EXPECT().MakeWorkloadTranslator(ctx, in).Return(mockWorkloadTranslator)
		mockDependencyFactory.EXPECT().MakeTrafficTargetTranslator(ctx).Return(mockTrafficTargetTranslator)

		labeledMeta := metav1.ObjectMeta{Labels: labelutils.ClusterLabels("cluster")}

		meshes := v1alpha2sets.NewMeshSet(&v1alpha2.Mesh{ObjectMeta: labeledMeta})
		workloads := v1alpha2sets.NewWorkloadSet(&v1alpha2.Workload{ObjectMeta: labeledMeta})
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(&v1alpha2.TrafficTarget{ObjectMeta: labeledMeta})

		mockMeshTranslator.EXPECT().TranslateMeshes(in).Return(meshes)
		mockWorkloadTranslator.EXPECT().TranslateWorkloads(deployments, daemonSets, statefulSets, meshes).Return(workloads)
		mockTrafficTargetTranslator.EXPECT().TranslateTrafficTargets(services, workloads, meshes).Return(trafficTargets)

		out, err := t.Translate(ctx, in)
		Expect(err).NotTo(HaveOccurred())

		expectedOut, err := discovery.NewSinglePartitionedSnapshot(
			"mesh-discovery-1",
			labelutils.OwnershipLabels(),
			trafficTargets,
			workloads,
			meshes,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(out).To(Equal(expectedOut))
	})
})
