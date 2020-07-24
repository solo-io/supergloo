package mirror_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/decorators/trafficpolicy/mirror"
	mock_hostutils "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils/mocks"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("MirrorDecorator", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		mirrorDecorator           trafficpolicy.VirtualServiceDecorator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
	})

	It("should decorate mirror", func() {
		meshServices := v1alpha2sets.NewMeshServiceSet(
			&discoveryv1alpha2.MeshService{
				Spec: discoveryv1alpha2.MeshServiceSpec{
					Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
						KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, meshServices)
		originalService := &discoveryv1alpha2.MeshService{
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				Mirror: &v1alpha2.TrafficPolicySpec_Mirror{
					DestinationType: &v1alpha2.TrafficPolicySpec_Mirror_KubeService{
						KubeService: &v1.ClusterObjectRef{
							Name:        "mirror",
							Namespace:   "namespace",
							ClusterName: "local-cluster",
						},
					},
					Percentage: 50,
					// Not specifying port should default to the single port on the Mirror destination
				},
			},
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(originalService.Spec.GetKubeService().Ref.ClusterName, appliedPolicy.Spec.Mirror.GetKubeService()).
			Return(localHostname)

		output := &v1alpha3.HTTPRoute{}
		expectedMirror := &v1alpha3.Destination{
			Host: localHostname,
		}
		expectedMirrorPercentage := &v1alpha3.Percent{
			Value: appliedPolicy.Spec.Mirror.Percentage,
		}
		err := mirrorDecorator.ApplyToVirtualService(
			appliedPolicy,
			originalService,
			output,
			registerField,
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Mirror).To(Equal(expectedMirror))
		Expect(output.MirrorPercentage).To(Equal(expectedMirrorPercentage))
	})

	It("should throw error if mirror destination has multiple ports but port is not specified in TrafficPolicy", func() {
		meshServices := v1alpha2sets.NewMeshServiceSet(
			&discoveryv1alpha2.MeshService{
				Spec: discoveryv1alpha2.MeshServiceSpec{
					Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
						KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
								{
									Port:     9081,
									Name:     "http2",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, meshServices)
		originalService := &discoveryv1alpha2.MeshService{
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				Mirror: &v1alpha2.TrafficPolicySpec_Mirror{
					DestinationType: &v1alpha2.TrafficPolicySpec_Mirror_KubeService{
						KubeService: &v1.ClusterObjectRef{
							Name:        "mirror",
							Namespace:   "namespace",
							ClusterName: "local-cluster",
						},
					},
					Percentage: 50,
					// Not specifying port should result in error
				},
			},
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(originalService.Spec.GetKubeService().Ref.ClusterName, appliedPolicy.Spec.Mirror.GetKubeService()).
			Return(localHostname)

		output := &v1alpha3.HTTPRoute{}
		err := mirrorDecorator.ApplyToVirtualService(
			appliedPolicy,
			originalService,
			output,
			registerField,
		)
		Expect(err.Error()).To(ContainSubstring("must provide port for mirror destination service"))
	})

	It("should not set mirror if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		meshServices := v1alpha2sets.NewMeshServiceSet(
			&discoveryv1alpha2.MeshService{
				Spec: discoveryv1alpha2.MeshServiceSpec{
					Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
						KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1alpha2.MeshServiceSpec_KubeService_KubeServicePort{
								{
									Port:     9080,
									Name:     "http1",
									Protocol: "http",
								},
							},
						},
					},
				},
			})
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, meshServices)
		originalService := &discoveryv1alpha2.MeshService{
			Spec: discoveryv1alpha2.MeshServiceSpec{
				Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
					KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		appliedPolicy := &discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
			Spec: &v1alpha2.TrafficPolicySpec{
				Mirror: &v1alpha2.TrafficPolicySpec_Mirror{
					DestinationType: &v1alpha2.TrafficPolicySpec_Mirror_KubeService{
						KubeService: &v1.ClusterObjectRef{
							Name:        "mirror",
							Namespace:   "namespace",
							ClusterName: "local-cluster",
						},
					},
					Percentage: 50,
					// Not specifying port should default to the single port on the Mirror destination
				},
			},
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(originalService.Spec.GetKubeService().Ref.ClusterName, appliedPolicy.Spec.Mirror.GetKubeService()).
			Return(localHostname)

		output := &v1alpha3.HTTPRoute{}
		err := mirrorDecorator.ApplyToVirtualService(
			appliedPolicy,
			originalService,
			output,
			registerField,
		)

		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.Fault).To(BeNil())
	})
})
