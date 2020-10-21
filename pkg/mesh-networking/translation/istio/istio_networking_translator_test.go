package istio

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio"
	mock_istio_output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio/mocks"
	mock_local_output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/local/mocks"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	mock_extensions "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/extensions/mocks"
	mock_istio "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/internal/mocks"
	mock_mesh "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/mocks"
	mock_traffictarget "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/traffictarget/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	multiclusterv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IstioNetworkingTranslator", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		ctxWithValue                context.Context
		mockIstioExtensions         *mock_extensions.MockIstioExtensions
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
		mockIstioExtensions = mock_extensions.NewMockIstioExtensions(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockTrafficTargetTranslator = mock_traffictarget.NewMockTranslator(ctrl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctrl)
		mockDependencyFactory = mock_istio.NewMockDependencyFactory(ctrl)
		mockIstioOutputs = mock_istio_output.NewMockBuilder(ctrl)
		mockLocalOutputs = mock_local_output.NewMockBuilder(ctrl)
		translator = &istioTranslator{dependencies: mockDependencyFactory, extensions: mockIstioExtensions}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		in := input.NewInputSnapshotManualBuilder("").
			AddKubernetesClusters([]*multiclusterv1alpha1.KubernetesCluster{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-cluster",
						Namespace: "namespace",
					},
				},
			}).
			AddMeshes([]*discoveryv1alpha2.Mesh{
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
			AddWorkloads([]*discoveryv1alpha2.Workload{
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
			AddTrafficTargets([]*discoveryv1alpha2.TrafficTarget{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-service-1",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-service-2",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
			}).Build()

		contextMatcher := gomock.Any()
		mockDependencyFactory.
			EXPECT().
			MakeTrafficTargetTranslator(contextMatcher, in.KubernetesClusters(), in.TrafficTargets(), in.FailoverServices()).
			Return(mockTrafficTargetTranslator)

		istioOutputsMatcher := gomock.AssignableToTypeOf(istio.NewBuilder(ctx, ""))

		for _, trafficTarget := range in.TrafficTargets().List() {
			mockTrafficTargetTranslator.
				EXPECT().
				Translate(in, trafficTarget, istioOutputsMatcher, mockReporter)
			mockIstioExtensions.EXPECT().PatchTrafficTargetOutputs(contextMatcher, trafficTarget, istioOutputsMatcher)
			mockIstioOutputs.EXPECT().Merge(istioOutputsMatcher)
		}

		for _, workload := range in.Workloads().List() {
			mockIstioExtensions.EXPECT().PatchWorkloadOutputs(contextMatcher, workload, istioOutputsMatcher)
			mockIstioOutputs.EXPECT().Merge(istioOutputsMatcher)
		}

		mockDependencyFactory.
			EXPECT().
			MakeMeshTranslator(ctxWithValue, in.KubernetesClusters(), in.Secrets(), in.Workloads(), in.TrafficTargets(), in.FailoverServices()).
			Return(mockMeshTranslator)
		for _, mesh := range in.Meshes().List() {
			mockMeshTranslator.
				EXPECT().
				Translate(in, mesh, istioOutputsMatcher, mockLocalOutputs, mockReporter)
			mockIstioExtensions.EXPECT().PatchMeshOutputs(contextMatcher, mesh, istioOutputsMatcher)
			mockIstioOutputs.EXPECT().Merge(istioOutputsMatcher)
		}

		translator.Translate(ctx, in, mockIstioOutputs, mockLocalOutputs, mockReporter)
	})
})
