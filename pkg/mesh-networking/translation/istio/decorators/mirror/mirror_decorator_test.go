package mirror_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mirror"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/go-utils/testutils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"istio.io/api/networking/v1alpha3"
)

var _ = Describe("MirrorDecorator", func() {
	var (
		ctrl                      *gomock.Controller
		mockClusterDomainRegistry *mock_hostutils.MockClusterDomainRegistry
		mirrorDecorator           decorators.TrafficPolicyVirtualServiceDecorator
		output                    *v1alpha3.HTTPRoute
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClusterDomainRegistry = mock_hostutils.NewMockClusterDomainRegistry(ctrl)
		output = &v1alpha3.HTTPRoute{}
	})

	It("should decorate mirror", func() {
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, destinations)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					Mirror: &v1.TrafficPolicySpec_Policy_Mirror{
						DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
							KubeService: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
						},
						Percentage: 50,
						// Not specifying port should default to the single port on the Mirror destination
					},
				},
			},
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.Spec.GetKubeService().Ref.ClusterName, appliedPolicy.Spec.GetPolicy().GetMirror().GetKubeService()).
			Return(localHostname)

		expectedMirror := &v1alpha3.Destination{
			Host: localHostname,
		}
		expectedMirrorPercentage := &v1alpha3.Percent{
			Value: appliedPolicy.Spec.Policy.Mirror.Percentage,
		}
		err := mirrorDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, originalService, nil, output, registerField)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Mirror).To(Equal(expectedMirror))
		Expect(output.MirrorPercentage).To(Equal(expectedMirrorPercentage))
	})

	It("should decorate mirror for federated Destination", func() {
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, destinations)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					Mirror: &v1.TrafficPolicySpec_Policy_Mirror{
						DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
							KubeService: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
						},
						Percentage: 50,
						// Not specifying port should default to the single port on the Mirror destination
					},
				},
			},
		}

		sourceMeshInstallation := &discoveryv1.MeshInstallation{
			Cluster: "federated-cluster-name",
		}
		globalHostname := "name.namespace.svc.local-cluster.global"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(sourceMeshInstallation.GetCluster(), appliedPolicy.Spec.Policy.Mirror.GetKubeService()).
			Return(globalHostname)

		expectedMirror := &v1alpha3.Destination{
			Host: globalHostname,
		}
		expectedMirrorPercentage := &v1alpha3.Percent{
			Value: appliedPolicy.Spec.Policy.Mirror.Percentage,
		}
		err := mirrorDecorator.ApplyTrafficPolicyToVirtualService(
			appliedPolicy,
			originalService,
			sourceMeshInstallation,
			output,
			registerField,
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Mirror).To(Equal(expectedMirror))
		Expect(output.MirrorPercentage).To(Equal(expectedMirrorPercentage))
	})

	It("should throw error if mirror destination has multiple ports but port is not specified in TrafficPolicy", func() {
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, destinations)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		registerField := func(fieldPtr, val interface{}) error {
			return nil
		}
		appliedPolicyMissingPort := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					Mirror: &v1.TrafficPolicySpec_Policy_Mirror{
						DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
							KubeService: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
						},
						Percentage: 50,
						// Not specifying port should result in error
					},
				},
			},
		}
		appliedPolicyNonexistentPort := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					Mirror: &v1.TrafficPolicySpec_Policy_Mirror{
						DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
							KubeService: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
						},
						Percentage: 50,
						Port:       1,
					},
				},
			},
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.Spec.GetKubeService().Ref.ClusterName, appliedPolicyMissingPort.Spec.Policy.Mirror.GetKubeService()).
			Return(localHostname).
			Times(2)

		err := mirrorDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicyMissingPort, originalService, nil, output, registerField)
		Expect(err.Error()).To(ContainSubstring("must provide port for mirror destination service"))

		err = mirrorDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicyNonexistentPort, originalService, nil, output, registerField)
		Expect(err.Error()).To(ContainSubstring("does not exist for mirror destination service"))
	})

	It("should not set mirror if error during field registration", func() {
		testErr := eris.New("registration error")
		registerField := func(fieldPtr, val interface{}) error {
			return testErr
		}
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, destinations)
		originalService := &discoveryv1.Destination{
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		appliedPolicy := &v1.AppliedTrafficPolicy{
			Spec: &v1.TrafficPolicySpec{
				Policy: &v1.TrafficPolicySpec_Policy{
					Mirror: &v1.TrafficPolicySpec_Policy_Mirror{
						DestinationType: &v1.TrafficPolicySpec_Policy_Mirror_KubeService{
							KubeService: &skv2corev1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
						},
						Percentage: 50,
						// Not specifying port should default to the single port on the Mirror destination
					},
				},
			},
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationFQDN(originalService.Spec.GetKubeService().Ref.ClusterName, appliedPolicy.Spec.Policy.Mirror.GetKubeService()).
			Return(localHostname)

		err := mirrorDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, originalService, nil, output, registerField)

		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.Fault).To(BeNil())
	})
})
