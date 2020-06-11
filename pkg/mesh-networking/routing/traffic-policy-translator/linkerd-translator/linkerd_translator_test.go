package linkerd_translator_test

import (
	"context"
	"time"

	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster/mocks"
	"k8s.io/apimachinery/pkg/api/resource"

	smi_config "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"

	"github.com/solo-io/service-mesh-hub/test/fakes"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	linkerd_networking "github.com/solo-io/service-mesh-hub/pkg/api/linkerd/v1alpha2"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	smi_networking "github.com/solo-io/service-mesh-hub/pkg/api/smi/split/v1alpha1"
	linkerd_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/routing/traffic-policy-translator/linkerd-translator"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("LinkerdTranslator", func() {
	var (
		ctrl                           *gomock.Controller
		linkerdTrafficPolicyTranslator linkerd_translator.LinkerdTranslator
		ctx                            context.Context
		mockDynamicClientGetter        *mock_multicluster.MockDynamicClientGetter
		mockMeshClient                 *mock_core.MockMeshClient

		clusterName       = "clusterName"
		meshObjKey        = client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
		meshServiceObjKey = client.ObjectKey{Name: "mesh-service-name", Namespace: "mesh-service-namespace"}
		kubeServiceObjKey = client.ObjectKey{Name: "kube-service-name", Namespace: "kube-service-namespace"}
		meshService       = &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:        meshServiceObjKey.Name,
				Namespace:   meshServiceObjKey.Namespace,
				ClusterName: clusterName,
			},
			Spec: smh_discovery_types.MeshServiceSpec{
				Mesh: &smh_core_types.ResourceRef{
					Name:      meshObjKey.Name,
					Namespace: meshObjKey.Namespace,
				},
				KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
					Ref: &smh_core_types.ResourceRef{
						Name:      kubeServiceObjKey.Name,
						Namespace: kubeServiceObjKey.Namespace,
						Cluster:   clusterName,
					},
					Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{
						{
							Port: 9080,
							Name: "http",
						},
					},
				},
			},
		}
		mesh = &smh_discovery.Mesh{
			Spec: smh_discovery_types.MeshSpec{
				Cluster: &smh_core_types.ResourceRef{
					Name: clusterName,
				},
				MeshType: &smh_discovery_types.MeshSpec_Linkerd{
					Linkerd: &smh_discovery_types.MeshSpec_LinkerdMesh{
						ClusterDomain: "cluster.domain",
					},
				},
			},
		}
		serviceProfileClient linkerd_networking.ServiceProfileClient
		trafficSplitClient   smi_networking.TrafficSplitClient
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockDynamicClientGetter = mock_multicluster.NewMockDynamicClientGetter(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		serviceProfileClient = linkerd_networking.NewServiceProfileClient(fakes.InMemoryClient())
		trafficSplitClient = smi_networking.NewTrafficSplitClient(fakes.InMemoryClient())
		linkerdTrafficPolicyTranslator = linkerd_translator.NewLinkerdTrafficPolicyTranslator(
			mockDynamicClientGetter,
			mockMeshClient,
			func(client client.Client) linkerd_networking.ServiceProfileClient {
				return serviceProfileClient
			},
			func(client client.Client) smi_networking.TrafficSplitClient {
				return trafficSplitClient
			},
		)
		mockMeshClient.EXPECT().GetMesh(ctx, meshObjKey).Return(mesh, nil)
		mockDynamicClientGetter.EXPECT().GetClientForCluster(ctx, clusterName).Return(nil, nil)

	})
	AfterEach(func() {
		ctrl.Finish()
	})
	Context("no relevant config provided", func() {

		trafficPolicy := []*smh_networking.TrafficPolicy{{
			Spec: smh_networking_types.TrafficPolicySpec{
				FaultInjection: &smh_networking_types.TrafficPolicySpec_FaultInjection{
					Percentage: 100,
				},
				HeaderManipulation: &smh_networking_types.TrafficPolicySpec_HeaderManipulation{
					AppendResponseHeaders: map[string]string{"foo": "bar"},
				},
			}},
		}

		It("does not create a service profile", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(BeNil())

			serviceProfiles, err := serviceProfileClient.ListServiceProfile(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceProfiles.Items).To(BeEmpty())
		})
	})

	Context("basic traffic policy", func() {
		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				Spec: smh_networking_types.TrafficPolicySpec{
					HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{
						{}, // one default matcher
					},
				},
			},
		}

		It("creates sp with the name and namespace matching the kube DNS name and namespace of the backing service, respectively", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(BeNil())

			serviceProfiles, err := serviceProfileClient.ListServiceProfile(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceProfiles.Items).To(HaveLen(1))

			sp := &serviceProfiles.Items[0]

			Expect(sp.Name).To(Equal("kube-service-name.kube-service-namespace.cluster.domain"))
			Expect(sp.Namespace).To(Equal(kubeServiceObjKey.Namespace))
		})
	})

	Context("prefix matcher provided", func() {

		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp"},
				Spec: smh_networking_types.TrafficPolicySpec{
					HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{
						{
							PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
								Prefix: "/prefix/",
							},
							Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{Method: smh_core_types.HttpMethodValue_GET},
						},
					},
				},
			},
		}

		It("creates sp with the paths converted to regex", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(BeNil())

			serviceProfiles, err := serviceProfileClient.ListServiceProfile(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceProfiles.Items).To(HaveLen(1))

			sp := &serviceProfiles.Items[0]

			Expect(sp.Spec.Routes).To(Equal([]*linkerd_config.RouteSpec{
				{
					Name: "tp.ns",
					Condition: &linkerd_config.RequestMatch{
						Any: []*linkerd_config.RequestMatch{
							{
								PathRegex: "/prefix/.*",
								Method:    "GET",
							},
						},
					},
				},
			}))
		})
	})

	Context("traffic shift provided", func() {

		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp"},
				Spec: smh_networking_types.TrafficPolicySpec{
					TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
						Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
							{
								Destination: &smh_core_types.ResourceRef{Name: "foo-svc", Namespace: meshService.Spec.KubeService.Ref.Namespace},
								Weight:      5,
							},
							{
								Destination: &smh_core_types.ResourceRef{Name: "bar-svc", Namespace: meshService.Spec.KubeService.Ref.Namespace},
								Weight:      15,
							},
						},
					},
				},
			},
		}

		It("creates traffic split with the corresponding split", func() {

			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(BeNil())

			trafficSplits, err := trafficSplitClient.ListTrafficSplit(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(trafficSplits.Items).To(HaveLen(1))

			ts := &trafficSplits.Items[0]

			makeQty := func(val int64) *resource.Quantity {
				// annoying, we need to do this because un/marshalling the quantity changes its Go representation
				q := resource.MustParse(resource.NewScaledQuantity(val, resource.Milli).String())
				return &q
			}

			Expect(ts.Spec.Backends).To(Equal([]smi_config.TrafficSplitBackend{
				{
					Service: "foo-svc",
					Weight:  makeQty(250),
				},
				{
					Service: "bar-svc",
					Weight:  makeQty(750),
				},
			}))

		})
	})

	Context("timeout provided", func() {

		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp"},
				Spec: smh_networking_types.TrafficPolicySpec{
					RequestTimeout: types.DurationProto(time.Minute),
				},
			},
		}

		It("creates sp with the corresponding timeout", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(BeNil())

			serviceProfiles, err := serviceProfileClient.ListServiceProfile(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceProfiles.Items).To(HaveLen(1))

			sp := &serviceProfiles.Items[0]

			Expect(sp.Spec.Routes).To(Equal([]*linkerd_config.RouteSpec{
				{
					Name: "tp.ns",
					Condition: &linkerd_config.RequestMatch{
						Any: []*linkerd_config.RequestMatch{
							{
								PathRegex: "/.*",
								Method:    "",
							},
						},
					},
					Timeout: time.Minute.String(),
				},
			}))
		})
	})

	Context("multiple policies defined", func() {

		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp1"},
				Spec: smh_networking_types.TrafficPolicySpec{
					HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{
						{
							PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
								Prefix: "/short",
							},
							Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{Method: smh_core_types.HttpMethodValue_GET},
						},
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp2"},
				Spec: smh_networking_types.TrafficPolicySpec{
					HttpRequestMatchers: []*smh_networking_types.TrafficPolicySpec_HttpMatcher{
						{
							PathSpecifier: &smh_networking_types.TrafficPolicySpec_HttpMatcher_Prefix{
								Prefix: "/longer",
							},
							Method: &smh_networking_types.TrafficPolicySpec_HttpMethod{Method: smh_core_types.HttpMethodValue_GET},
						},
					},
				},
			},
		}

		It("sorts the routes in the sp byt the length of the first path specifier", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(BeNil())

			serviceProfiles, err := serviceProfileClient.ListServiceProfile(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceProfiles.Items).To(HaveLen(1))

			sp := &serviceProfiles.Items[0]

			Expect(sp.Spec.Routes).To(Equal([]*linkerd_config.RouteSpec{
				{
					Name: "tp2.ns",
					Condition: &linkerd_config.RequestMatch{
						Any: []*linkerd_config.RequestMatch{
							{
								PathRegex: "/longer.*",
								Method:    "GET",
							},
						},
					},
				},
				{
					Name: "tp1.ns",
					Condition: &linkerd_config.RequestMatch{
						Any: []*linkerd_config.RequestMatch{
							{
								PathRegex: "/short.*",
								Method:    "GET",
							},
						},
					},
				},
			}))
		})
	})
	Context("cross-namespace defined in destination", func() {

		dest := &smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
			Destination: &smh_core_types.ResourceRef{Namespace: "another-namespace"},
			Subset:      map[string]string{},
		}
		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp"},
				Spec: smh_networking_types.TrafficPolicySpec{
					TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
						Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{dest},
					},
				},
			},
		}

		It("returns a translator error", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(Equal(&smh_networking_types.TrafficPolicyStatus_TranslatorError{
				TranslatorId: linkerd_translator.TranslatorId,
				ErrorMessage: multierror.Append(nil, linkerd_translator.CrossNamespaceSplitNotSupportedErr).Error(),
			}))
		})
	})
	Context("subsets defined in destination", func() {

		dest := &smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
			Destination: &smh_core_types.ResourceRef{Namespace: meshService.Spec.KubeService.Ref.Namespace},
			Subset:      map[string]string{},
		}
		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp"},
				Spec: smh_networking_types.TrafficPolicySpec{
					TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{
						Destinations: []*smh_networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{dest},
					},
				},
			},
		}

		It("returns a translator error", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(Equal(&smh_networking_types.TrafficPolicyStatus_TranslatorError{
				TranslatorId: linkerd_translator.TranslatorId,
				ErrorMessage: multierror.Append(nil, linkerd_translator.SubsetsNotSupportedErr(dest)).Error(),
			}))
		})
	})
	Context("multiple policies defined with traffic shift", func() {

		trafficPolicy := []*smh_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp1"},
				Spec: smh_networking_types.TrafficPolicySpec{
					TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: "ns", Name: "tp2"},
				Spec: smh_networking_types.TrafficPolicySpec{
					TrafficShift: &smh_networking_types.TrafficPolicySpec_MultiDestination{},
				},
			},
		}

		It("returns a translator error", func() {
			translatorError := linkerdTrafficPolicyTranslator.TranslateTrafficPolicy(
				ctx,
				meshService,
				mesh,
				trafficPolicy)
			Expect(translatorError).To(Equal(&smh_networking_types.TrafficPolicyStatus_TranslatorError{
				TranslatorId: linkerd_translator.TranslatorId,
				ErrorMessage: multierror.Append(nil, linkerd_translator.TrafficShiftRedefinedErr(meshService, []smh_core_types.ResourceRef{
					{Namespace: "ns", Name: "tp1"},
					{Namespace: "ns", Name: "tp2"},
				})).Error(),
			}))
		})
	})

})
