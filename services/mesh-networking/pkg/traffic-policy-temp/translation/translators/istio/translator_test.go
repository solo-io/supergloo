package istio_translator_test

import (
	proto_types "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	mock_selection "github.com/solo-io/service-mesh-hub/pkg/common/kube/selection/mocks"
	istio_networking_types "istio.io/api/networking/v1alpha3"
	istio_client_networking_types "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators/istio"
)

var _ = Describe("Istio Traffic Policy Translator", func() {
	var (
		ctrl      *gomock.Controller
		istioMesh = &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
					Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{},
				},
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	When("a service has no policies applying to it", func() {
		It("should not error", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				nil,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.DestinationRules).To(Equal([]*istio_client_networking_types.DestinationRule{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.DestinationRule{
					Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
					TrafficPolicy: &istio_networking_types.TrafficPolicy{
						Tls: &istio_networking_types.TLSSettings{
							Mode: istio_networking_types.TLSSettings_ISTIO_MUTUAL,
						},
					},
				},
			}}))
			Expect(result.VirtualServices).To(HaveLen(0))
		})
	})

	When("there are policies that need to be translated", func() {
		It("should yield a virtual service", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref:               &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.DestinationRules).To(Equal([]*istio_client_networking_types.DestinationRule{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.DestinationRule{
					Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
					TrafficPolicy: &istio_networking_types.TrafficPolicy{
						Tls: &istio_networking_types.TLSSettings{
							Mode: istio_networking_types.TLSSettings_ISTIO_MUTUAL,
						},
					},
				},
			}}))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		When("no destination is specified on the policy", func() {
			When("no ports are set on the service", func() {
				It("should report an error", func() {
					resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
					policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref:               &smh_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{},
						},
					}
					serviceBeingTranslated := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: container_runtime.GetWriteNamespace(),
						},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
								Ref: &smh_core_types.ResourceRef{
									Name:      "kube-svc",
									Namespace: "application-namespace",
								},
							},
						},
					}

					translator := NewIstioTrafficPolicyTranslator(resourceSelector)
					result, errs := translator.Translate(
						serviceBeingTranslated,
						[]*smh_discovery.MeshService{serviceBeingTranslated},
						istioMesh,
						policies,
					)
					Expect(errs).To(HaveLen(1))
					Expect(errs[0].Policy).To(Equal(policies[0]))
					Expect(errs[0].TranslatorErrors).To(HaveLen(1))
					Expect(errs[0].TranslatorErrors[0].TranslatorId).To(Equal(translator.Name()))
					Expect(errs[0].TranslatorErrors[0].ErrorMessage).To(ContainSubstring(NoSpecifiedPortError(serviceBeingTranslated).Error()))
					// we want to have 1 result destination rule; to make sure we don't delete it on an error

					Expect(result.VirtualServices).To(BeNil())
					Expect(result.DestinationRules).NotTo(BeNil())
				})
			})

			When("multiple ports are set on the service", func() {
				It("should report an error", func() {
					resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
					policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref:               &smh_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{},
						},
					}
					serviceBeingTranslated := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: container_runtime.GetWriteNamespace(),
						},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
								Ref: &smh_core_types.ResourceRef{
									Name:      "kube-svc",
									Namespace: "application-namespace",
								},
								Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
									{
										Name: "port-1",
									},
									{
										Name: "port-2",
									},
								},
							},
						},
					}

					translator := NewIstioTrafficPolicyTranslator(resourceSelector)
					result, errs := translator.Translate(
						serviceBeingTranslated,
						[]*smh_discovery.MeshService{serviceBeingTranslated},
						istioMesh,
						policies,
					)

					Expect(errs).To(HaveLen(1))
					Expect(errs[0].Policy).To(Equal(policies[0]))
					Expect(errs[0].TranslatorErrors).To(HaveLen(1))
					Expect(errs[0].TranslatorErrors[0].TranslatorId).To(Equal(translator.Name()))
					Expect(errs[0].TranslatorErrors[0].ErrorMessage).To(ContainSubstring(NoSpecifiedPortError(serviceBeingTranslated).Error()))

					Expect(result.VirtualServices).To(BeNil())
					Expect(result.DestinationRules).NotTo(BeNil())
				})
			})
		})

		It("should translate retries", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts:      5,
							PerTryTimeout: &proto_types.Duration{Seconds: 2},
						},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						Retries: &istio_networking_types.HTTPRetry{
							Attempts:      5,
							PerTryTimeout: &proto_types.Duration{Seconds: 2},
						},
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		It("should translate CORS policy", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						CorsPolicy: &smh_networking_types.TrafficPolicySpec_CorsPolicy{
							AllowOrigins: []*smh_networking_types.TrafficPolicySpec_StringMatch{
								{MatchType: &smh_networking_types.TrafficPolicySpec_StringMatch_Exact{Exact: "exact"}},
								{MatchType: &smh_networking_types.TrafficPolicySpec_StringMatch_Prefix{Prefix: "prefix"}},
								{MatchType: &smh_networking_types.TrafficPolicySpec_StringMatch_Regex{Regex: "regex"}},
							},
							AllowMethods:     []string{"GET", "POST"},
							AllowHeaders:     []string{"Header1", "Header2"},
							ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
							MaxAge:           &proto_types.Duration{Seconds: 1},
							AllowCredentials: &proto_types.BoolValue{Value: false},
						},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						CorsPolicy: &istio_networking_types.CorsPolicy{
							AllowOrigins: []*istio_networking_types.StringMatch{
								{MatchType: &istio_networking_types.StringMatch_Exact{Exact: "exact"}},
								{MatchType: &istio_networking_types.StringMatch_Prefix{Prefix: "prefix"}},
								{MatchType: &istio_networking_types.StringMatch_Regex{Regex: "regex"}},
							},
							AllowMethods:     []string{"GET", "POST"},
							AllowHeaders:     []string{"Header1", "Header2"},
							ExposeHeaders:    []string{"ExposedHeader1", "ExposedHeader2"},
							MaxAge:           &proto_types.Duration{Seconds: 1},
							AllowCredentials: &proto_types.BoolValue{Value: false},
						},
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		It("should translate HeaderManipulation", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						HeaderManipulation: &smh_networking_types.TrafficPolicySpec_HeaderManipulation{
							AppendRequestHeaders:  map[string]string{"a": "b"},
							RemoveRequestHeaders:  []string{"3", "4"},
							AppendResponseHeaders: map[string]string{"foo": "bar"},
							RemoveResponseHeaders: []string{"1", "2"},
						},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						Headers: &istio_networking_types.Headers{
							Request: &istio_networking_types.Headers_HeaderOperations{
								Add:    map[string]string{"a": "b"},
								Remove: []string{"3", "4"},
							},
							Response: &istio_networking_types.Headers_HeaderOperations{
								Add:    map[string]string{"foo": "bar"},
								Remove: []string{"1", "2"},
							},
						},
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		When("translating mirror destinations", func() {
			Context("and the mirror destination is on the same cluster", func() {
				It("should translate the Mirror config properly", func() {
					resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
					destName := "name"
					destNamespace := "namespace"
					port := uint32(9080)
					destCluster := "test-cluster"
					policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
								Mirror: &smh_networking_types.TrafficPolicySpec_Mirror{
									Destination: &smh_core_types.ResourceRef{
										Name:      destName,
										Namespace: destNamespace,
										Cluster:   destCluster,
									},
									Percentage: 50,
									Port:       port,
								},
							},
						},
					}
					serviceBeingTranslated := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: container_runtime.GetWriteNamespace(),
						},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
								Ref: &smh_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
						},
					}
					otherService := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service-being-mirrored-to",
							Namespace: container_runtime.GetWriteNamespace(),
						},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
								Ref: &smh_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
						},
					}
					allServices := []*smh_discovery.MeshService{serviceBeingTranslated, serviceBeingTranslated}

					resourceSelector.EXPECT().
						FindMeshServiceByRefSelector(
							allServices,
							policies[0].TrafficPolicySpec.Mirror.Destination.Name,
							policies[0].TrafficPolicySpec.Mirror.Destination.Namespace,
							policies[0].TrafficPolicySpec.Mirror.Destination.Cluster,
						).
						Return(otherService)

					translator := NewIstioTrafficPolicyTranslator(resourceSelector)
					result, errs := translator.Translate(
						serviceBeingTranslated,
						allServices,
						istioMesh,
						policies,
					)
					Expect(errs).To(HaveLen(0))
					Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
						ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
						Spec: istio_networking_types.VirtualService{
							Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
							Http: []*istio_networking_types.HTTPRoute{{
								Mirror: &istio_networking_types.Destination{
									Host: destName,
									Port: &istio_networking_types.PortSelector{
										Number: port,
									},
								},
								MirrorPercentage: &istio_networking_types.Percent{Value: 50.0},
								Route: []*istio_networking_types.HTTPRouteDestination{{
									Destination: &istio_networking_types.Destination{
										Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
										Port: &istio_networking_types.PortSelector{
											Number: 8000,
										},
									},
								}},
							}},
						},
					}}))
				})
			})

			Context("the mirror destination is on a remote cluster", func() {
				It("should translate the mirror config properly", func() {
					resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
					destName := "name"
					destNamespace := "namespace"
					port := uint32(9080)
					destCluster := "test-cluster"
					policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
								Mirror: &smh_networking_types.TrafficPolicySpec_Mirror{
									Destination: &smh_core_types.ResourceRef{
										Name:      destName,
										Namespace: destNamespace,
										Cluster:   destCluster + "-remote-version",
									},
									Percentage: 50,
									Port:       port,
								},
							},
						},
					}
					serviceBeingTranslated := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: container_runtime.GetWriteNamespace(),
						},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
								Ref: &smh_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
						},
					}
					otherService := &smh_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service-being-mirrored-to",
							Namespace: container_runtime.GetWriteNamespace(),
						},
						Spec: smh_discovery_types.MeshServiceSpec{
							Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
								Ref: &smh_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
							Federation: &smh_discovery_types.MeshServiceSpec_Federation{
								MulticlusterDnsName: "multicluster.dns.name",
							},
						},
					}
					allServices := []*smh_discovery.MeshService{serviceBeingTranslated, serviceBeingTranslated}

					resourceSelector.EXPECT().
						FindMeshServiceByRefSelector(
							allServices,
							policies[0].TrafficPolicySpec.Mirror.Destination.Name,
							policies[0].TrafficPolicySpec.Mirror.Destination.Namespace,
							policies[0].TrafficPolicySpec.Mirror.Destination.Cluster,
						).
						Return(otherService)

					translator := NewIstioTrafficPolicyTranslator(resourceSelector)
					result, errs := translator.Translate(
						serviceBeingTranslated,
						allServices,
						istioMesh,
						policies,
					)
					Expect(errs).To(HaveLen(0))
					Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
						ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
						Spec: istio_networking_types.VirtualService{
							Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
							Http: []*istio_networking_types.HTTPRoute{{
								Mirror: &istio_networking_types.Destination{
									Host: otherService.Spec.Federation.MulticlusterDnsName,
									Port: &istio_networking_types.PortSelector{
										Number: port,
									},
								},
								MirrorPercentage: &istio_networking_types.Percent{Value: 50.0},
								Route: []*istio_networking_types.HTTPRouteDestination{{
									Destination: &istio_networking_types.Destination{
										Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
										Port: &istio_networking_types.PortSelector{
											Number: 8000,
										},
									},
								}},
							}},
						},
					}}))
				})
			})
		})

		It("should translate fault injections", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						FaultInjection: &smh_networking_types.TrafficPolicySpec_FaultInjection{
							FaultInjectionType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
								Delay: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay{
									HttpDelayType: &smh_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
										FixedDelay: &proto_types.Duration{Seconds: 2},
									},
								},
							},
							Percentage: 50,
						},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						Fault: &istio_networking_types.HTTPFaultInjection{
							Delay: &istio_networking_types.HTTPFaultInjection_Delay{
								HttpDelayType: &istio_networking_types.HTTPFaultInjection_Delay_FixedDelay{FixedDelay: &proto_types.Duration{Seconds: 2}},
								Percentage:    &istio_networking_types.Percent{Value: 50},
							},
						},
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		It("should translate retries", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						Retries: &smh_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts:      5,
							PerTryTimeout: &proto_types.Duration{Seconds: 2},
						},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						Retries: &istio_networking_types.HTTPRetry{
							Attempts:      5,
							PerTryTimeout: &proto_types.Duration{Seconds: 2},
						},
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		It("should translate HeaderMatchers", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{{
							Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{Method: smh_core_types.HttpMethodValue_GET},
							Headers: []*smh_networking_types.TrafficPolicySpec_HeaderMatcher{
								{
									Name:        "name1",
									Value:       "value1",
									Regex:       false,
									InvertMatch: false,
								},
								{
									Name:        "name2",
									Value:       "*",
									Regex:       true,
									InvertMatch: false,
								},
								{
									Name:        "name3",
									Value:       "[a-z]+",
									Regex:       true,
									InvertMatch: true,
								},
							},
						}},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						Match: []*istio_networking_types.HTTPMatchRequest{{
							Method: &istio_networking_types.StringMatch{
								MatchType: &istio_networking_types.StringMatch_Exact{
									Exact: "GET",
								},
							},
							Headers: map[string]*istio_networking_types.StringMatch{
								"name1": {
									MatchType: &istio_networking_types.StringMatch_Exact{
										Exact: "value1",
									},
								},
								"name2": {
									MatchType: &istio_networking_types.StringMatch_Regex{
										Regex: "*",
									},
								},
							},
							WithoutHeaders: map[string]*istio_networking_types.StringMatch{
								"name3": {
									MatchType: &istio_networking_types.StringMatch_Regex{
										Regex: "[a-z]+",
									},
								},
							},
						}},
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		It("should translate query param matchers", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{{
							Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{Method: smh_core_types.HttpMethodValue_GET},
							QueryParameters: []*smh_networking_types.TrafficPolicySpec_QueryParameterMatcher{
								{
									Name:  "qp1",
									Value: "qpv1",
									Regex: false,
								},
								{
									Name:  "qp2",
									Value: "qpv2",
									Regex: true,
								},
							},
						}},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     8000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{{
						Match: []*istio_networking_types.HTTPMatchRequest{{
							Method: &istio_networking_types.StringMatch{
								MatchType: &istio_networking_types.StringMatch_Exact{
									Exact: "GET",
								},
							},
							QueryParams: map[string]*istio_networking_types.StringMatch{
								"qp1": {
									MatchType: &istio_networking_types.StringMatch_Exact{Exact: "qpv1"},
								},
								"qp2": {
									MatchType: &istio_networking_types.StringMatch_Regex{Regex: "qpv2"},
								},
							},
						}},
						Route: []*istio_networking_types.HTTPRouteDestination{{
							Destination: &istio_networking_types.Destination{
								Host: serviceBeingTranslated.Spec.KubeService.Ref.Name,
								Port: &istio_networking_types.PortSelector{
									Number: 8000,
								},
							},
						}},
					}},
				},
			}}))
		})

		When("translating traffic shifts", func() {
			It("should translate traffic shifts without subsets", func() {
				resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
				destName := "name"
				destNamespace := "namespace"
				multiClusterDnsName := "multicluster-dns-name"
				port := uint32(9080)
				destCluster := "remote-cluster-1"
				policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
						TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
							TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
								Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
									{
										Destination: &smh_core_types.ResourceRef{
											Name:      destName,
											Namespace: destNamespace,
											Cluster:   destCluster,
										},
										Weight: 50,
										Port:   port,
									},
								},
							},
						},
					},
				}
				serviceBeingTranslated := &smh_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "kube-svc",
								Namespace: "application-namespace",
							},
							Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &smh_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				otherService := &smh_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service-being-shifted-to",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      destName,
								Namespace: destNamespace,
								Cluster:   destCluster,
							},
							Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &smh_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				allServices := []*smh_discovery.MeshService{serviceBeingTranslated, otherService}

				resourceSelector.EXPECT().
					FindMeshServiceByRefSelector(
						allServices,
						policies[0].TrafficPolicySpec.TrafficShift.Destinations[0].Destination.Name,
						policies[0].TrafficPolicySpec.TrafficShift.Destinations[0].Destination.Namespace,
						policies[0].TrafficPolicySpec.TrafficShift.Destinations[0].Destination.Cluster,
					).
					Return(otherService)

				translator := NewIstioTrafficPolicyTranslator(resourceSelector)
				result, errs := translator.Translate(
					serviceBeingTranslated,
					allServices,
					istioMesh,
					policies,
				)
				Expect(errs).To(HaveLen(0))
				Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
					ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
					Spec: istio_networking_types.VirtualService{
						Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
						Http: []*istio_networking_types.HTTPRoute{{
							Route: []*istio_networking_types.HTTPRouteDestination{{
								Destination: &istio_networking_types.Destination{
									Host: multiClusterDnsName,
									Port: &istio_networking_types.PortSelector{
										Number: port,
									},
								},
								Weight: 50,
							}},
						}},
					},
				}}))
			})

			It("should translate traffic shifts with subsets", func() {
				resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
				destName := "name"
				destNamespace := "namespace"
				declaredSubset := map[string]string{"env": "dev", "version": "v1"}
				expectedSubsetName := "env-dev_version-v1"
				port := uint32(9080)
				multiClusterDnsName := "multi-cluster-dns-name"
				clusterName := "test-cluster"
				policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
						TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
							TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
								Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
									{
										Destination: &smh_core_types.ResourceRef{
											Name:      destName,
											Namespace: destNamespace,
											Cluster:   clusterName,
										},
										Weight: 50,
										Subset: declaredSubset,
										Port:   port,
									},
								},
							},
						},
					},
				}
				serviceBeingTranslated := &smh_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "kube-svc",
								Namespace: "application-namespace",
								Cluster:   clusterName,
							},
							Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &smh_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				otherService := &smh_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service-being-shifted-to",
						Namespace: container_runtime.GetWriteNamespace(),
					},
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      destName,
								Namespace: destNamespace,
								Cluster:   clusterName,
							},
							Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &smh_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				allServices := []*smh_discovery.MeshService{serviceBeingTranslated, otherService}

				resourceSelector.EXPECT().
					FindMeshServiceByRefSelector(
						allServices,
						policies[0].TrafficPolicySpec.TrafficShift.Destinations[0].Destination.Name,
						policies[0].TrafficPolicySpec.TrafficShift.Destinations[0].Destination.Namespace,
						policies[0].TrafficPolicySpec.TrafficShift.Destinations[0].Destination.Cluster,
					).
					Return(otherService)

				translator := NewIstioTrafficPolicyTranslator(resourceSelector)
				result, errs := translator.Translate(
					serviceBeingTranslated,
					allServices,
					istioMesh,
					policies,
				)
				Expect(errs).To(HaveLen(0))
				Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
					ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
					Spec: istio_networking_types.VirtualService{
						Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
						Http: []*istio_networking_types.HTTPRoute{{
							Route: []*istio_networking_types.HTTPRouteDestination{{
								Destination: &istio_networking_types.Destination{
									Host:   destName,
									Subset: expectedSubsetName,
									Port: &istio_networking_types.PortSelector{
										Number: port,
									},
								},
								Weight: 50,
							}},
						}},
					},
				}}))
			})
		})

		It("should deterministically order HTTPRoutes according to decreasing specificity", func() {
			resourceSelector := mock_selection.NewMockResourceSelector(ctrl)
			sourceNamespace := "source-namespace"
			policies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &smh_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
									Exact: "exact-path",
								},
							},
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/prefix",
								},
								Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{
									Method: smh_core_types.HttpMethodValue_GET,
								},
							},
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
									Exact: "exact-path",
								},
								Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{
									Method: smh_core_types.HttpMethodValue_GET,
								},
							},
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
									Exact: "exact-path",
								},
								Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{
									Method: smh_core_types.HttpMethodValue_PUT,
								},
							},
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Regex{
									Regex: "www*",
								},
							},
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/",
								},
								Headers: []*smh_networking_types.TrafficPolicySpec_HeaderMatcher{
									{
										Name:        "set-cookie",
										Value:       "foo=bar",
										InvertMatch: true,
									},
								},
							},
							{
								PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/",
								},
								Headers: []*smh_networking_types.TrafficPolicySpec_HeaderMatcher{
									{
										Name:        "content-type",
										Value:       "text/html",
										InvertMatch: false,
									},
								},
							},
						},
					},
				},
			}
			serviceBeingTranslated := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: selection.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
							Cluster:   sourceNamespace,
						},
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Port:     9000,
							Name:     "test-port",
							Protocol: "tcp",
						}},
					},
				},
			}
			defaultRoute := []*istio_networking_types.HTTPRouteDestination{
				{
					Destination: &istio_networking_types.Destination{
						Host: serviceBeingTranslated.Spec.GetKubeService().GetRef().GetName(),
						Port: &istio_networking_types.PortSelector{
							Number: serviceBeingTranslated.Spec.KubeService.Ports[0].Port,
						},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*smh_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: selection.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
				Spec: istio_networking_types.VirtualService{
					Hosts: []string{serviceBeingTranslated.Spec.KubeService.Ref.Name},
					Http: []*istio_networking_types.HTTPRoute{
						{
							Match: []*istio_networking_types.HTTPMatchRequest{
								{

									Headers: map[string]*istio_networking_types.StringMatch{
										"content-type": {MatchType: &istio_networking_types.StringMatch_Exact{Exact: "text/html"}},
									},
									Uri: &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Prefix{Prefix: "/"}},
								},
							},
							Route: defaultRoute,
						},
						{
							Match: []*istio_networking_types.HTTPMatchRequest{
								{

									Uri:    &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: "exact-path"}},
									Method: &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: "PUT"}},
								},
							},
							Route: defaultRoute,
						},
						{
							Match: []*istio_networking_types.HTTPMatchRequest{
								{

									Uri:    &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: "exact-path"}},
									Method: &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: "GET"}},
								},
							},
							Route: defaultRoute,
						},
						{
							Match: []*istio_networking_types.HTTPMatchRequest{
								{

									Uri: &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: "exact-path"}},
								},
							},
							Route: defaultRoute,
						},
						{
							Match: []*istio_networking_types.HTTPMatchRequest{
								{

									Uri: &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Regex{Regex: "www*"}},
								},
							},
							Route: defaultRoute,
						},
						{
							Match: []*istio_networking_types.HTTPMatchRequest{
								{

									Uri:    &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Prefix{Prefix: "/prefix"}},
									Method: &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Exact{Exact: "GET"}},
								},
							},
							Route: defaultRoute,
						},
						{
							Match: []*istio_networking_types.HTTPMatchRequest{
								{

									WithoutHeaders: map[string]*istio_networking_types.StringMatch{
										"set-cookie": {MatchType: &istio_networking_types.StringMatch_Exact{Exact: "foo=bar"}},
									},
									Uri: &istio_networking_types.StringMatch{MatchType: &istio_networking_types.StringMatch_Prefix{Prefix: "/"}},
								},
							},
							Route: defaultRoute,
						},
					},
				},
			}}))
		})
	})
})
