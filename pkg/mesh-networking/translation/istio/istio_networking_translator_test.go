package istio

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	mock_output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/istio/mocks"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
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
		mockReporter                *mock_reporting.MockReporter
		mockOutputs                 *mock_output.MockBuilder
		mockTrafficTargetTranslator *mock_traffictarget.MockTranslator
		mockMeshTranslator          *mock_mesh.MockTranslator
		mockDependencyFactory       *mock_istio.MockDependencyFactory
		translator                  Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		ctxWithValue = contextutils.WithLogger(context.TODO(), "istio-translator-0")
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockTrafficTargetTranslator = mock_traffictarget.NewMockTranslator(ctrl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctrl)
		mockDependencyFactory = mock_istio.NewMockDependencyFactory(ctrl)
		mockOutputs = mock_output.NewMockBuilder(ctrl)
		translator = &istioTranslator{dependencies: mockDependencyFactory}
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

		mockDependencyFactory.
			EXPECT().
			MakeTrafficTargetTranslator(in.KubernetesClusters(), in.TrafficTargets()).
			Return(mockTrafficTargetTranslator)
		for i := range in.TrafficTargets().List() {
			mockTrafficTargetTranslator.
				EXPECT().
				Translate(in, in.TrafficTargets().List()[i], mockOutputs, mockReporter)
		}

		mockDependencyFactory.
			EXPECT().
			MakeMeshTranslator(ctxWithValue, in.KubernetesClusters(), in.Secrets(), in.Workloads(), in.TrafficTargets()).
			Return(mockMeshTranslator)
		for i := range in.Meshes().List() {
			mockMeshTranslator.
				EXPECT().
				Translate(in, in.Meshes().List()[i], mockOutputs, mockReporter)
		}

		translator.Translate(ctx, in, mockOutputs, mockReporter)
	})
})
