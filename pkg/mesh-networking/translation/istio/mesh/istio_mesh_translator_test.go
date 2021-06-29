package mesh_test

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	mock_mtls "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls/mocks"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh"
	mock_access "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/access/mocks"
	mock_federation "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/federation/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IstioMeshTranslator", func() {
	var (
		ctrl                     *gomock.Controller
		ctx                      context.Context
		mockMtlsTranslator       *mock_mtls.MockTranslator
		mockFederationTranslator *mock_federation.MockTranslator
		mockAccessTranslator     *mock_access.MockTranslator
		mockReporter             *mock_reporting.MockReporter
		in                       input.LocalSnapshot
		istioMeshTranslator      mesh.Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMtlsTranslator = mock_mtls.NewMockTranslator(ctrl)
		mockFederationTranslator = mock_federation.NewMockTranslator(ctrl)
		mockAccessTranslator = mock_access.NewMockTranslator(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		in = input.NewInputLocalSnapshotManualBuilder("").Build()
		istioMeshTranslator = mesh.NewTranslator(ctx, mockMtlsTranslator, mockFederationTranslator, mockAccessTranslator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		outputs := istio.NewBuilder(context.TODO(), "")
		localOutputs := local.NewBuilder(context.TODO(), "")

		istioMesh := &discoveryv1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mesh-1",
				Namespace: "mesh-namespace-1",
			},
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{
						Installation: &discoveryv1.MeshInstallation{
							Cluster:   "cluster-1",
							Namespace: "istio-system",
						},
					},
				},
			},
			Status: discoveryv1.MeshStatus{
				AppliedVirtualMesh: &discoveryv1.MeshStatus_AppliedVirtualMesh{},
			},
		}

		mockMtlsTranslator.
			EXPECT().
			Translate(istioMesh, istioMesh.Status.AppliedVirtualMesh, outputs, localOutputs, mockReporter)

		mockFederationTranslator.
			EXPECT().
			Translate(in, istioMesh, istioMesh.Status.AppliedVirtualMesh, outputs, mockReporter)

		mockAccessTranslator.
			EXPECT().
			Translate(istioMesh, istioMesh.Status.AppliedVirtualMesh, outputs)

		istioMeshTranslator.Translate(in, istioMesh, outputs, localOutputs, mockReporter)
	})
})
