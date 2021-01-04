package istio

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	mock_istio_output "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio/mocks"
	mock_local_output "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local/mocks"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	mock_extensions "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/extensions/mocks"
	mock_istio "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/internal/mocks"
	mock_mesh "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mocks"
	mock_traffictarget "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/traffictarget/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	multiclusterv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IstioNetworkingTranslator", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		ctxWithValue                context.Context
		mockIstioExtender           *mock_extensions.MockIstioExtender
		mockReporter                *mock_reporting.MockReporter
		mockIstioOutputs            *mock_istio_output.MockBuilder
		mockLocalOutputs            *mock_local_output.MockBuilder
		mockTrafficTargetTranslator *mock_traffictarget.MockTranslator
		mockMeshTranslator          *mock_mesh.MockTranslator
		mockDependencyFactory       *mock_istio.MockDependencyFactory
		translator                  Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		ctxWithValue = contextutils.WithLogger(context.TODO(), "istio-translator-0")
		mockIstioExtender = mock_extensions.NewMockIstioExtender(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockTrafficTargetTranslator = mock_traffictarget.NewMockTranslator(ctrl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctrl)
		mockDependencyFactory = mock_istio.NewMockDependencyFactory(ctrl)
		mockIstioOutputs = mock_istio_output.NewMockBuilder(ctrl)
		mockLocalOutputs = mock_local_output.NewMockBuilder(ctrl)
		translator = &istioTranslator{dependencies: mockDependencyFactory, extender: mockIstioExtender}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		in := input.NewInputLocalSnapshotManualBuilder("").
			AddMulticlusterSoloIov1Alpha1KubernetesClusters([]*multiclusterv1alpha1.KubernetesCluster{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-cluster",
						Namespace: "namespace",
					},
				},
			}).
			AddDiscoveryMeshGlooSoloIov1Alpha2Meshes([]*discoveryv1alpha2.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mesh-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mesh-2",
					},
				},
			}).
			AddDiscoveryMeshGlooSoloIov1Alpha2Workloads([]*discoveryv1alpha2.Workload{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-workload-1",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-workload-2",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
			}).
			AddDiscoveryMeshGlooSoloIov1Alpha2TrafficTargets([]*discoveryv1alpha2.TrafficTarget{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "traffic-target-1",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "traffic-target-2",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
			}).Build()

		contextMatcher := gomock.Any()
		mockDependencyFactory.
			EXPECT().
			MakeTrafficTargetTranslator(contextMatcher, nil, in.MulticlusterSoloIov1Alpha1KubernetesClusters(), in.DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets(), in.NetworkingMeshGlooSoloIov1Alpha2FailoverServices()).
			Return(mockTrafficTargetTranslator)

		for _, trafficTarget := range in.DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets().List() {
			mockTrafficTargetTranslator.
				EXPECT().
				Translate(in, trafficTarget, mockIstioOutputs, mockReporter)
		}

		mockDependencyFactory.
			EXPECT().
			MakeMeshTranslator(ctxWithValue, nil, in.MulticlusterSoloIov1Alpha1KubernetesClusters(), in.V1Secrets(), in.DiscoveryMeshGlooSoloIov1Alpha2Workloads(), in.DiscoveryMeshGlooSoloIov1Alpha2TrafficTargets(), in.NetworkingMeshGlooSoloIov1Alpha2FailoverServices()).
			Return(mockMeshTranslator)
		for _, mesh := range in.NetworkingMeshGlooSoloIov1Alpha2VirtualMeshes().List() {
			mockMeshTranslator.
				EXPECT().
				Translate(in, mesh, mockIstioOutputs, mockLocalOutputs, mockReporter)
		}

		mockIstioExtender.EXPECT().PatchOutputs(contextMatcher, in, mockIstioOutputs)

		translator.Translate(ctx, in, nil, mockIstioOutputs, mockLocalOutputs, mockReporter)
	})
})
