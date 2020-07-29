package istio

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	v1beta1sets "github.com/solo-io/external-apis/pkg/api/istio/security.istio.io/v1beta1/sets"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input/test"
	istio2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output/istio"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	mock_istio "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/internal/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh"
	mock_mesh "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice"
	mock_meshservice "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	multiclusterv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IstioNetworkingTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		ctxWithValue              context.Context
		mockReporter              *mock_reporting.MockReporter
		mockMeshServiceTranslator *mock_meshservice.MockTranslator
		mockMeshTranslator        *mock_mesh.MockTranslator
		mockDependencyFactory     *mock_istio.MockDependencyFactory
		translator                Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		ctxWithValue = contextutils.WithLogger(context.TODO(), "istio-translator-0")
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockMeshServiceTranslator = mock_meshservice.NewMockTranslator(ctrl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctrl)
		mockDependencyFactory = mock_istio.NewMockDependencyFactory(ctrl)
		translator = &istioTranslator{dependencies: mockDependencyFactory}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		in := test.NewInputSnapshotBuilder("").
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
			AddMeshServices([]*discoveryv1alpha2.MeshService{
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

		// MeshService translation
		msVirtualServices := []*v1alpha3.VirtualService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ms-virtual-service-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ms-virtual-service-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		msDestinationRules := []*v1alpha3.DestinationRule{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ms-destination-rule-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ms-destination-rule-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		msAuthPolicies := []*v1beta1.AuthorizationPolicy{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ms-authorization-policy-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "ms-authorization-policy-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		mockDependencyFactory.
			EXPECT().
			MakeMeshServiceTranslator(in.KubernetesClusters()).
			Return(mockMeshServiceTranslator)
		for i := range in.MeshServices().List() {
			mockMeshServiceTranslator.
				EXPECT().
				Translate(in, in.MeshServices().List()[i], mockReporter).
				Return(meshservice.Outputs{
					VirtualService:      msVirtualServices[i],
					DestinationRule:     msDestinationRules[i],
					AuthorizationPolicy: msAuthPolicies[i],
				})
		}

		// Mesh translation
		mGateways := []*v1alpha3.Gateway{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-gateway-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-gateway-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		mEnvoyFilters := []*v1alpha3.EnvoyFilter{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-envoy-filter-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-envoy-filter-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		mDestinationRules := []*v1alpha3.DestinationRule{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-destination-rule-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-destination-rule-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		mServiceEntries := []*v1alpha3.ServiceEntry{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-service-entry-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-service-entry-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		mAuthPolicies := []*v1beta1.AuthorizationPolicy{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-authorization-policy-1",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "m-authorization-policy-2",
					Labels: metautils.TranslatedObjectLabels(),
				},
			},
		}
		mockDependencyFactory.
			EXPECT().
			MakeMeshTranslator(ctxWithValue, in.KubernetesClusters()).
			Return(mockMeshTranslator)
		for i := range in.Meshes().List() {
			mockMeshTranslator.
				EXPECT().
				Translate(in, in.Meshes().List()[i], mockReporter).
				Return(mesh.Outputs{
					Gateways:              v1alpha3sets.NewGatewaySet(mGateways[i]),
					EnvoyFilters:          v1alpha3sets.NewEnvoyFilterSet(mEnvoyFilters[i]),
					DestinationRules:      v1alpha3sets.NewDestinationRuleSet(mDestinationRules[i]),
					ServiceEntries:        v1alpha3sets.NewServiceEntrySet(mServiceEntries[i]),
					AuthorizationPolicies: v1beta1sets.NewAuthorizationPolicySet(mAuthPolicies[i]),
				})
		}

		expectedOutput, err := istio2.NewBuilder("istio-networking-1").
			AddDestinationRules(append(msDestinationRules, mDestinationRules...)).
			AddEnvoyFilters(mEnvoyFilters).
			AddGateways(mGateways).
			AddServiceEntries(mServiceEntries).
			AddVirtualServices(msVirtualServices).
			AddAuthorizationPolicies(append(msAuthPolicies, mAuthPolicies...)).
			BuildSinglePartitionedSnapshot(metautils.TranslatedObjectLabels())
		Expect(err).ToNot(HaveOccurred())

		output, err := translator.Translate(ctx, in, mockReporter)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(Equal(expectedOutput))
	})
})
