package translation_test

import (
	"context"

	"github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1beta2sets "github.com/solo-io/external-apis/pkg/api/appmesh/appmesh.k8s.aws/v1beta2/sets"
	appsv1sets "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1/sets"
	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/output/discovery"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	settingsv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1alpha2"
	. "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation"
	mock_destination "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/destination/mocks"
	mock_translator_internal "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/internal/mocks"
	mock_mesh "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/mesh/mocks"
	mock_workload "github.com/solo-io/gloo-mesh/pkg/mesh-discovery/translation/workload/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/labelutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {

	var (
		ctl *gomock.Controller
		ctx context.Context

		mockDependencyFactory     *mock_translator_internal.MockDependencyFactory
		mockMeshTranslator        *mock_mesh.MockTranslator
		mockWorkloadTranslator    *mock_workload.MockTranslator
		mockDestinationTranslator *mock_destination.MockTranslator
	)

	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDependencyFactory = mock_translator_internal.NewMockDependencyFactory(ctl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctl)
		mockWorkloadTranslator = mock_workload.NewMockTranslator(ctl)
		mockDestinationTranslator = mock_destination.NewMockTranslator(ctl)
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
		endpoints := corev1sets.NewEndpointsSet(&corev1.Endpoints{})
		deployments := appsv1sets.NewDeploymentSet(&appsv1.Deployment{})
		replicaSets := appsv1sets.NewReplicaSetSet(&appsv1.ReplicaSet{})
		daemonSets := appsv1sets.NewDaemonSetSet(&appsv1.DaemonSet{})
		statefulSets := appsv1sets.NewStatefulSetSet(&appsv1.StatefulSet{})
		inRemote := input.NewDiscoveryInputSnapshot(
			"mesh-discovery-remote",
			appMeshes,
			configMaps,
			services,
			pods,
			endpoints,
			nodes,
			deployments,
			replicaSets,
			daemonSets,
			statefulSets,
		)
		settings := &settingsv1alpha2.DiscoverySettings{}

		mockDependencyFactory.EXPECT().MakeMeshTranslator(ctx).Return(mockMeshTranslator)
		mockDependencyFactory.EXPECT().MakeWorkloadTranslator(ctx, inRemote).Return(mockWorkloadTranslator)
		mockDependencyFactory.EXPECT().MakeDestinationTranslator().Return(mockDestinationTranslator)

		labeledMeta := metav1.ObjectMeta{Labels: labelutils.ClusterLabels("cluster")}

		meshes := v1alpha2sets.NewMeshSet(&v1alpha2.Mesh{ObjectMeta: labeledMeta})
		workloads := v1alpha2sets.NewWorkloadSet(&v1alpha2.Workload{ObjectMeta: labeledMeta})
		destinations := v1alpha2sets.NewDestinationSet(&v1alpha2.Destination{ObjectMeta: labeledMeta})

		mockMeshTranslator.EXPECT().TranslateMeshes(inRemote, settings).Return(meshes)
		mockWorkloadTranslator.EXPECT().TranslateWorkloads(deployments, daemonSets, statefulSets, meshes).Return(workloads)
		mockDestinationTranslator.EXPECT().TranslateDestinations(ctx, services, pods, nodes, workloads, meshes, endpoints).Return(destinations)

		out, err := t.Translate(ctx, inRemote, settings)
		Expect(err).NotTo(HaveOccurred())

		expectedOut, err := discovery.NewSinglePartitionedSnapshot(
			"mesh-discovery-1",
			labelutils.OwnershipLabels(),
			destinations,
			workloads,
			meshes,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(out).To(Equal(expectedOut))
	})
})
