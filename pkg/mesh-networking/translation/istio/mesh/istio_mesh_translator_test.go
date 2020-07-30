package mesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istiov1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	v1beta1sets "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/sets"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh"
	mock_access "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/access/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/failoverservice"
	mock_failoverservice "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/failoverservice/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/federation"
	mock_federation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/federation/mocks"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IstioMeshTranslator", func() {
	var (
		ctrl                          *gomock.Controller
		ctx                           context.Context
		mockFederationTranslator      *mock_federation.MockTranslator
		mockAccessTranslator          *mock_access.MockTranslator
		mockFailoverServiceTranslator *mock_failoverservice.MockTranslator
		mockReporter                  *mock_reporting.MockReporter
		in                            input.Snapshot
		istioMeshTranslator           mesh.Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockFederationTranslator = mock_federation.NewMockTranslator(ctrl)
		mockAccessTranslator = mock_access.NewMockTranslator(ctrl)
		mockFailoverServiceTranslator = mock_failoverservice.NewMockTranslator(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		in = input.NewInputSnapshotManualBuilder("").Build()
		istioMeshTranslator = mesh.NewTranslator(ctx, mockFederationTranslator, mockAccessTranslator, mockFailoverServiceTranslator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		istioMesh := &discoveryv1alpha2.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mesh-1",
				Namespace: "mesh-namespace-1",
			},
			Spec: discoveryv1alpha2.MeshSpec{
				MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
					Istio: &discoveryv1alpha2.MeshSpec_Istio{
						Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
							Cluster:   "cluster-1",
							Namespace: "istio-system",
						},
					},
				},
			},
			Status: discoveryv1alpha2.MeshStatus{
				AppliedFailoverServices: []*discoveryv1alpha2.MeshStatus_AppliedFailoverService{
					{},
				},
				AppliedVirtualMeshes: []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
					{},
				},
			},
		}

		expectedOutputs := mesh.Outputs{
			Gateways: istiov1alpha3sets.NewGatewaySet(
				&v1alpha3.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "gateway"}},
			),
			EnvoyFilters: istiov1alpha3sets.NewEnvoyFilterSet(
				&v1alpha3.EnvoyFilter{ObjectMeta: metav1.ObjectMeta{Name: "vm-envoy-filter"}},
				&v1alpha3.EnvoyFilter{ObjectMeta: metav1.ObjectMeta{Name: "fs-envoy-filter"}},
			),
			DestinationRules: istiov1alpha3sets.NewDestinationRuleSet(
				&v1alpha3.DestinationRule{ObjectMeta: metav1.ObjectMeta{Name: "destination-rule"}},
			),
			ServiceEntries: istiov1alpha3sets.NewServiceEntrySet(
				&v1alpha3.ServiceEntry{ObjectMeta: metav1.ObjectMeta{Name: "vm-service-entry"}},
				&v1alpha3.ServiceEntry{ObjectMeta: metav1.ObjectMeta{Name: "fs-service-entry"}},
			),
			AuthorizationPolicies: v1beta1sets.NewAuthorizationPolicySet(
				&v1beta1.AuthorizationPolicy{ObjectMeta: metav1.ObjectMeta{Name: "authorization-policy"}},
			),
		}

		mockFederationTranslator.
			EXPECT().
			Translate(in, istioMesh, istioMesh.Status.AppliedVirtualMeshes[0], mockReporter).
			Return(federation.Outputs{
				Gateway:          expectedOutputs.Gateways.List()[0],
				EnvoyFilter:      expectedOutputs.EnvoyFilters.List()[0],
				DestinationRules: expectedOutputs.DestinationRules,
				ServiceEntries:   istiov1alpha3sets.NewServiceEntrySet(expectedOutputs.ServiceEntries.List()[0]),
			})

		mockAccessTranslator.
			EXPECT().
			Translate(istioMesh, istioMesh.Status.AppliedVirtualMeshes[0]).
			Return(expectedOutputs.AuthorizationPolicies)

		mockFailoverServiceTranslator.
			EXPECT().
			Translate(in, istioMesh, istioMesh.Status.AppliedFailoverServices[0], mockReporter).
			Return(failoverservice.Outputs{
				EnvoyFilters:   istiov1alpha3sets.NewEnvoyFilterSet(expectedOutputs.EnvoyFilters.List()[1]),
				ServiceEntries: istiov1alpha3sets.NewServiceEntrySet(expectedOutputs.ServiceEntries.List()[1]),
			})

		outputs := istioMeshTranslator.Translate(in, istioMesh, mockReporter)
		Expect(outputs).To(Equal(expectedOutputs))
	})
})
