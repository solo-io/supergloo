package mirror_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.gloomesh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.gloomesh.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.gloomesh.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/mirror"
	mock_hostutils "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils/mocks"
	"github.com/solo-io/go-utils/testutils"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, trafficTargets)
		originalService := &discoveryv1alpha2.TrafficTarget{
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
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
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
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

		expectedMirror := &v1alpha3.Destination{
			Host: localHostname,
		}
		expectedMirrorPercentage := &v1alpha3.Percent{
			Value: appliedPolicy.Spec.Mirror.Percentage,
		}
		err := mirrorDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, originalService, nil, output, registerField)

		Expect(err).ToNot(HaveOccurred())
		Expect(output.Mirror).To(Equal(expectedMirror))
		Expect(output.MirrorPercentage).To(Equal(expectedMirrorPercentage))
	})

	It("should decorate mirror for federated TrafficTarget", func() {
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, trafficTargets)
		originalService := &discoveryv1alpha2.TrafficTarget{
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
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
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
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

		sourceMeshInstallation := &discoveryv1alpha2.MeshSpec_MeshInstallation{
			Cluster: "federated-cluster-name",
		}
		globalHostname := "name.namespace.svc.local-cluster.global"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(sourceMeshInstallation.GetCluster(), appliedPolicy.Spec.Mirror.GetKubeService()).
			Return(globalHostname)

		expectedMirror := &v1alpha3.Destination{
			Host: globalHostname,
		}
		expectedMirrorPercentage := &v1alpha3.Percent{
			Value: appliedPolicy.Spec.Mirror.Percentage,
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
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, trafficTargets)
		originalService := &discoveryv1alpha2.TrafficTarget{
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
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
		appliedPolicyMissingPort := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
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
		appliedPolicyNonexistentPort := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
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
					Port:       1,
				},
			},
		}

		localHostname := "name.namespace.svc.cluster.local"
		mockClusterDomainRegistry.
			EXPECT().
			GetDestinationServiceFQDN(originalService.Spec.GetKubeService().Ref.ClusterName, appliedPolicyMissingPort.Spec.Mirror.GetKubeService()).
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
		trafficTargets := v1alpha2sets.NewTrafficTargetSet(
			&discoveryv1alpha2.TrafficTarget{
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "mirror",
								Namespace:   "namespace",
								ClusterName: "local-cluster",
							},
							Ports: []*discoveryv1alpha2.TrafficTargetSpec_KubeService_KubeServicePort{
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
		mirrorDecorator = mirror.NewMirrorDecorator(mockClusterDomainRegistry, trafficTargets)
		originalService := &discoveryv1alpha2.TrafficTarget{
			Spec: discoveryv1alpha2.TrafficTargetSpec{
				Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
					KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
						Ref: &v1.ClusterObjectRef{
							ClusterName: "local-cluster",
						},
					},
				},
			},
		}
		appliedPolicy := &discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy{
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

		err := mirrorDecorator.ApplyTrafficPolicyToVirtualService(appliedPolicy, originalService, nil, output, registerField)

		Expect(err).To(testutils.HaveInErrorChain(testErr))
		Expect(output.Fault).To(BeNil())
	})
})
