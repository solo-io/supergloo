package preprocess_test

import (
	"context"

	types1 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Merger", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		trafficPolicyMerger     preprocess.TrafficPolicyMerger
		mockResourceSelector    *mock_selector.MockResourceSelector
		mockMeshClient          *mock_core.MockMeshClient
		mockTrafficPolicyClient *mock_zephyr_networking.MockTrafficPolicyClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockResourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockTrafficPolicyClient = mock_zephyr_networking.NewMockTrafficPolicyClient(ctrl)
		trafficPolicyMerger = preprocess.NewTrafficPolicyMerger(
			mockResourceSelector,
			mockMeshClient,
			mockTrafficPolicyClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should merge TrafficPolicy specs by MeshService", func() {
		// destinations
		meshServiceName1 := "meshServiceName1"
		meshServiceNamespace1 := "meshServiceNamespace1"
		clusterName1 := "clusterName1"
		meshName1 := "meshname1"
		meshNamespace1 := "meshnamespace1"
		meshServiceName2 := "meshServiceName2"
		meshServiceNamespace2 := "meshServiceNamespace2"
		clusterName2 := "clusterName2"
		meshName2 := "meshname2"
		meshNamespace2 := "meshnamespace2"
		meshClusterName1 := "mesh-cluster-name-1"
		meshClusterName2 := "mesh-cluster-name-2"
		// sources
		destNamespaces1 := []string{"namespace1"}
		destLabels1 := map[string]string{"k1": "v1"}
		destNamespaces2 := []string{"namespace2"}
		destLabels2 := map[string]string{"k2": "v2"}
		httpMatcher1 := &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
			Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_GET},
		}
		httpMatcher2 := &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
			Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_POST},
		}
		httpMatcher3 := &zephyr_networking_types.TrafficPolicySpec_HttpMatcher{
			Method: &zephyr_networking_types.TrafficPolicySpec_HttpMethod{Method: zephyr_core_types.HttpMethodValue_PUT},
		}
		tp1 := zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				SourceSelector: &zephyr_core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &zephyr_core_types.ServiceSelector{
					ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
							Services: []*zephyr_core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
							},
						},
					},
				},
				RequestTimeout:      &types1.Duration{Seconds: 1},
				HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher1},
			},
			Status: zephyr_networking_types.TrafficPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp2 := zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				SourceSelector: &zephyr_core_types.WorkloadSelector{
					Namespaces: destNamespaces2,
					Labels:     destLabels2,
				},
				DestinationSelector: &zephyr_core_types.ServiceSelector{
					ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
							Services: []*zephyr_core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
								{Name: meshServiceName2, Namespace: meshServiceNamespace2},
							},
						},
					},
				},
				RequestTimeout:      &types1.Duration{Seconds: 1},
				HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher2},
			},
			Status: zephyr_networking_types.TrafficPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp3 := zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				SourceSelector: &zephyr_core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &zephyr_core_types.ServiceSelector{
					ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
							Services: []*zephyr_core_types.ResourceRef{
								{Name: meshServiceName2, Namespace: meshServiceNamespace2},
							},
						},
					},
				},
				Retries:             &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{Attempts: 2},
				HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher1, httpMatcher2, httpMatcher3},
			},
			Status: zephyr_networking_types.TrafficPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp4 := zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				SourceSelector: &zephyr_core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &zephyr_core_types.ServiceSelector{
					ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
							Services: []*zephyr_core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
							},
						},
					},
				},
				FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
					Percentage: 50,
				},
				HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher3},
			},
			Status: zephyr_networking_types.TrafficPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp5 := zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				SourceSelector: &zephyr_core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &zephyr_core_types.ServiceSelector{
					ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
							Services: []*zephyr_core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
							},
						},
					},
				},
				FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
					Percentage: 50,
				},
				HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher1},
			},
			Status: zephyr_networking_types.TrafficPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		ignoredTP := zephyr_networking.TrafficPolicy{
			Spec: zephyr_networking_types.TrafficPolicySpec{
				DestinationSelector: &zephyr_core_types.ServiceSelector{
					ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &zephyr_core_types.ServiceSelector_ServiceRefs{
							Services: []*zephyr_core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
								{Name: meshServiceName2, Namespace: meshServiceNamespace2},
							},
						},
					},
				},
			},
			Status: zephyr_networking_types.TrafficPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_CONFLICT,
					Message: "",
				},
			},
		}
		meshServiceKey1 := selector.MeshServiceId{
			Name:        meshServiceName1,
			Namespace:   meshServiceNamespace1,
			ClusterName: meshClusterName1,
		}
		meshServiceKey2 := selector.MeshServiceId{
			Name:        meshServiceName2,
			Namespace:   meshServiceNamespace2,
			ClusterName: meshClusterName2,
		}
		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:        meshServiceName1,
					Namespace:   meshServiceNamespace1,
					ClusterName: clusterName1,
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      meshName1,
						Namespace: meshNamespace1,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:        meshServiceName2,
					Namespace:   meshServiceNamespace2,
					ClusterName: clusterName2,
				},
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      meshName2,
						Namespace: meshNamespace2,
					},
				},
			},
		}
		/*** GetMeshServicesByServiceSelector() ***/
		trafficPolicyList := &zephyr_networking.TrafficPolicyList{
			Items: []zephyr_networking.TrafficPolicy{tp1, tp2, tp3, tp4, tp5, ignoredTP}}
		mockTrafficPolicyClient.EXPECT().ListTrafficPolicy(ctx).Return(trafficPolicyList, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp1.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{meshServices[0]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp2.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{meshServices[0], meshServices[1]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp3.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{meshServices[1]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp4.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{meshServices[0]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp5.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{meshServices[0]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, ignoredTP.Spec.GetDestinationSelector()).
			Return([]*zephyr_discovery.MeshService{meshServices[0], meshServices[1]}, nil)
		/*** buildKeyForMeshService ***/
		mesh1 := &zephyr_discovery.Mesh{Spec: zephyr_discovery_types.MeshSpec{Cluster: &zephyr_core_types.ResourceRef{Name: meshClusterName1}}}
		mesh2 := &zephyr_discovery.Mesh{Spec: zephyr_discovery_types.MeshSpec{Cluster: &zephyr_core_types.ResourceRef{Name: meshClusterName2}}}
		mockMeshClient.EXPECT().GetMesh(ctx, client.ObjectKey{Name: meshName1, Namespace: meshNamespace1}).Return(mesh1, nil).Times(6)
		mockMeshClient.EXPECT().GetMesh(ctx, client.ObjectKey{Name: meshName2, Namespace: meshNamespace2}).Return(mesh2, nil).Times(4)
		mergedTrafficPolicy1 := []*zephyr_networking.TrafficPolicy{
			{
				Spec: zephyr_networking_types.TrafficPolicySpec{
					SourceSelector: &zephyr_core_types.WorkloadSelector{
						Namespaces: destNamespaces1,
						Labels:     destLabels1,
					},
					HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher1},
					RequestTimeout:      &types1.Duration{Seconds: 1},
					FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
						Percentage: 50,
					},
				},
			},
			{
				Spec: zephyr_networking_types.TrafficPolicySpec{
					SourceSelector: &zephyr_core_types.WorkloadSelector{
						Namespaces: destNamespaces2,
						Labels:     destLabels2,
					},
					HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher2},
					RequestTimeout:      &types1.Duration{Seconds: 1},
				},
			},
			{
				Spec: zephyr_networking_types.TrafficPolicySpec{
					SourceSelector: &zephyr_core_types.WorkloadSelector{
						Namespaces: destNamespaces1,
						Labels:     destLabels1,
					},
					HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher3},
					FaultInjection: &zephyr_networking_types.TrafficPolicySpec_FaultInjection{
						Percentage: 50,
					},
				},
			},
		}
		mergedTrafficPolicy2 := []*zephyr_networking.TrafficPolicy{
			{
				Spec: zephyr_networking_types.TrafficPolicySpec{
					SourceSelector: &zephyr_core_types.WorkloadSelector{
						Namespaces: destNamespaces2,
						Labels:     destLabels2,
					},
					HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher2},
					RequestTimeout:      &types1.Duration{Seconds: 1},
				},
			},
			{
				Spec: zephyr_networking_types.TrafficPolicySpec{
					SourceSelector: &zephyr_core_types.WorkloadSelector{
						Namespaces: destNamespaces1,
						Labels:     destLabels1,
					},
					HttpRequestMatchers: []*zephyr_networking_types.TrafficPolicySpec_HttpMatcher{httpMatcher1, httpMatcher2, httpMatcher3},
					Retries:             &zephyr_networking_types.TrafficPolicySpec_RetryPolicy{Attempts: 2},
				},
			},
		}
		policiesByMeshService, err := trafficPolicyMerger.MergeTrafficPoliciesForMeshServices(ctx, meshServices)
		Expect(err).To(BeNil())
		Expect(policiesByMeshService).To(HaveKeyWithValue(meshServiceKey1, mergedTrafficPolicy1))
		Expect(policiesByMeshService).To(HaveKeyWithValue(meshServiceKey2, mergedTrafficPolicy2))
	})
})
