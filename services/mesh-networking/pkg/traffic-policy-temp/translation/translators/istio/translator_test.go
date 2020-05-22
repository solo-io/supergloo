package istio_translator_test

import (
	proto_types "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators"
	istio_networking_types "istio.io/api/networking/v1alpha3"
	istio_client_networking_types "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/translators/istio"
)

var _ = Describe("Istio Traffic Policy Translator", func() {
	var (
		ctrl      *gomock.Controller
		istioMesh = &zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{
					Istio: &zephyr_discovery_types.MeshSpec_IstioMesh{},
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
					},
				},
			}

			translator := NewIstioTrafficPolicyTranslator(resourceSelector)
			result, errs := translator.Translate(
				serviceBeingTranslated,
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				nil,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.DestinationRules).To(Equal([]*istio_client_networking_types.DestinationRule{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref:               &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{},
				},
			}
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.DestinationRules).To(Equal([]*istio_client_networking_types.DestinationRule{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
					resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
					policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref:               &zephyr_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{},
						},
					}
					serviceBeingTranslated := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: env.GetWriteNamespace(),
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
								Ref: &zephyr_core_types.ResourceRef{
									Name:      "kube-svc",
									Namespace: "application-namespace",
								},
							},
						},
					}

					translator := NewIstioTrafficPolicyTranslator(resourceSelector)
					result, errs := translator.Translate(
						serviceBeingTranslated,
						[]*zephyr_discovery.MeshService{serviceBeingTranslated},
						istioMesh,
						policies,
					)
					Expect(errs).To(Equal([]*mesh_translation.TranslationError{{
						Policy: policies[0],
						TranslatorErrors: []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{{
							TranslatorId: translator.Name(),
							ErrorMessage: NoSpecifiedPortError(serviceBeingTranslated).Error(),
						}},
					}}))
					Expect(result).To(BeNil())
				})
			})

			When("multiple ports are set on the service", func() {
				It("should report an error", func() {
					resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
					policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref:               &zephyr_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{},
						},
					}
					serviceBeingTranslated := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: env.GetWriteNamespace(),
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
								Ref: &zephyr_core_types.ResourceRef{
									Name:      "kube-svc",
									Namespace: "application-namespace",
								},
								Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
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
						[]*zephyr_discovery.MeshService{serviceBeingTranslated},
						istioMesh,
						policies,
					)
					Expect(errs).To(Equal([]*mesh_translation.TranslationError{{
						Policy: policies[0],
						TranslatorErrors: []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{{
							TranslatorId: translator.Name(),
							ErrorMessage: NoSpecifiedPortError(serviceBeingTranslated).Error(),
						}},
					}}))
					Expect(result).To(BeNil())
				})
			})
		})

		It("should translate retries", func() {
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts:      5,
							PerTryTimeout: &proto_types.Duration{Seconds: 2},
						},
					},
				},
			}
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						CorsPolicy: &zephyr_networking_types.TrafficPolicySpec_CorsPolicy{
							AllowOrigins: []*zephyr_networking_types.TrafficPolicySpec_StringMatch{
								{MatchType: &zephyr_networking_types.TrafficPolicySpec_StringMatch_Exact{Exact: "exact"}},
								{MatchType: &zephyr_networking_types.TrafficPolicySpec_StringMatch_Prefix{Prefix: "prefix"}},
								{MatchType: &zephyr_networking_types.TrafficPolicySpec_StringMatch_Regex{Regex: "regex"}},
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
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						HeaderManipulation: &zephyr_networking_types.TrafficPolicySpec_HeaderManipulation{
							AppendRequestHeaders:  map[string]string{"a": "b"},
							RemoveRequestHeaders:  []string{"3", "4"},
							AppendResponseHeaders: map[string]string{"foo": "bar"},
							RemoveResponseHeaders: []string{"1", "2"},
						},
					},
				},
			}
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
					resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
					destName := "name"
					destNamespace := "namespace"
					port := uint32(9080)
					destCluster := "test-cluster"
					policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
								Mirror: &zephyr_networking_types.TrafficPolicySpec_Mirror{
									Destination: &zephyr_core_types.ResourceRef{
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
					serviceBeingTranslated := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: env.GetWriteNamespace(),
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
								Ref: &zephyr_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
						},
					}
					otherService := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service-being-mirrored-to",
							Namespace: env.GetWriteNamespace(),
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
								Ref: &zephyr_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
						},
					}
					allServices := []*zephyr_discovery.MeshService{serviceBeingTranslated, serviceBeingTranslated}

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
						ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
					resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
					destName := "name"
					destNamespace := "namespace"
					port := uint32(9080)
					destCluster := "test-cluster"
					policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
							TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
								Mirror: &zephyr_networking_types.TrafficPolicySpec_Mirror{
									Destination: &zephyr_core_types.ResourceRef{
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
					serviceBeingTranslated := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service",
							Namespace: env.GetWriteNamespace(),
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
								Ref: &zephyr_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
						},
					}
					otherService := &zephyr_discovery.MeshService{
						ObjectMeta: k8s_meta_types.ObjectMeta{
							Name:      "mesh-service-being-mirrored-to",
							Namespace: env.GetWriteNamespace(),
						},
						Spec: zephyr_discovery_types.MeshServiceSpec{
							Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
							KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
								Ref: &zephyr_core_types.ResourceRef{
									Name:      destName,
									Namespace: destNamespace,
									Cluster:   destCluster,
								},
								Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
									Port:     8000,
									Name:     "test-port",
									Protocol: "tcp",
								}},
							},
							Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
								MulticlusterDnsName: "multicluster.dns.name",
							},
						},
					}
					allServices := []*zephyr_discovery.MeshService{serviceBeingTranslated, serviceBeingTranslated}

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
						ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
							FaultInjectionType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_{
								Delay: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay{
									HttpDelayType: &zephyr_networking_types.TrafficPolicySpec_FaultInjection_Delay_FixedDelay{
										FixedDelay: &proto_types.Duration{Seconds: 2},
									},
								},
							},
							Percentage: 50,
						},
					},
				},
			}
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						Retries: &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{
							Attempts:      5,
							PerTryTimeout: &proto_types.Duration{Seconds: 2},
						},
					},
				},
			}
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{{
							Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
							Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
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
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{{
							Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
							QueryParameters: []*zephyr_networking_types.TrafficPolicySpec_QueryParameterMatcher{
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
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
				resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
				destName := "name"
				destNamespace := "namespace"
				multiClusterDnsName := "multicluster-dns-name"
				port := uint32(9080)
				destCluster := "remote-cluster-1"
				policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
						TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
							TrafficShift: &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
								Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
									{
										Destination: &zephyr_core_types.ResourceRef{
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
				serviceBeingTranslated := &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
							Ref: &zephyr_core_types.ResourceRef{
								Name:      "kube-svc",
								Namespace: "application-namespace",
							},
							Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				otherService := &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service-being-shifted-to",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
							Ref: &zephyr_core_types.ResourceRef{
								Name:      destName,
								Namespace: destNamespace,
								Cluster:   destCluster,
							},
							Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				allServices := []*zephyr_discovery.MeshService{serviceBeingTranslated, otherService}

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
					ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
				resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
				destName := "name"
				destNamespace := "namespace"
				declaredSubset := map[string]string{"env": "dev", "version": "v1"}
				expectedSubsetName := "env-dev_version-v1"
				port := uint32(9080)
				multiClusterDnsName := "multi-cluster-dns-name"
				clusterName := "test-cluster"
				policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
						TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
							TrafficShift: &zephyr_networking_types.TrafficPolicySpec_MultiDestination{
								Destinations: []*zephyr_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
									{
										Destination: &zephyr_core_types.ResourceRef{
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
				serviceBeingTranslated := &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
							Ref: &zephyr_core_types.ResourceRef{
								Name:      "kube-svc",
								Namespace: "application-namespace",
								Cluster:   clusterName,
							},
							Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				otherService := &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-service-being-shifted-to",
						Namespace: env.GetWriteNamespace(),
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
						KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
							Ref: &zephyr_core_types.ResourceRef{
								Name:      destName,
								Namespace: destNamespace,
								Cluster:   clusterName,
							},
							Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Port:     8000,
								Name:     "test-port",
								Protocol: "tcp",
							}},
						},
						Federation: &zephyr_discovery_types.MeshServiceSpec_Federation{
							MulticlusterDnsName: multiClusterDnsName,
						},
					},
				}
				allServices := []*zephyr_discovery.MeshService{serviceBeingTranslated, otherService}

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
					ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			sourceNamespace := "source-namespace"
			policies := []*zephyr_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{Name: "policy-1"},
					TrafficPolicySpec: &zephyr_networking_types.TrafficPolicySpec{
						HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
									Exact: "exact-path",
								},
							},
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/prefix",
								},
								Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{
									Method: zephyr_core_types.HttpMethodValue_GET,
								},
							},
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
									Exact: "exact-path",
								},
								Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{
									Method: zephyr_core_types.HttpMethodValue_GET,
								},
							},
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Exact{
									Exact: "exact-path",
								},
								Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{
									Method: zephyr_core_types.HttpMethodValue_PUT,
								},
							},
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Regex{
									Regex: "www*",
								},
							},
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/",
								},
								Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
									{
										Name:        "set-cookie",
										Value:       "foo=bar",
										InvertMatch: true,
									},
								},
							},
							{
								PathSpecifier: &zephyr_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
									Prefix: "/",
								},
								Headers: []*zephyr_networking_types.TrafficPolicySpec_HeaderMatcher{
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
			serviceBeingTranslated := &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(istioMesh.ObjectMeta),
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "kube-svc",
							Namespace: "application-namespace",
							Cluster:   sourceNamespace,
						},
						Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
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
				[]*zephyr_discovery.MeshService{serviceBeingTranslated},
				istioMesh,
				policies,
			)
			Expect(errs).To(HaveLen(0))
			Expect(result.VirtualServices).To(Equal([]*istio_client_networking_types.VirtualService{{
				ObjectMeta: clients.ResourceRefToObjectMeta(serviceBeingTranslated.Spec.KubeService.Ref),
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
