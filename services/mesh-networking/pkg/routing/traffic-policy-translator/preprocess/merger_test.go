package preprocess_test

import (
	"context"

	types1 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_v1alpha1_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/mesh-projects/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/mesh-projects/pkg/selector"
	mock_selector "github.com/solo-io/mesh-projects/pkg/selector/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		httpMatcher1 := &networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{
			Method: &networking_v1alpha1_types.TrafficPolicySpec_HttpMethod{Method: core_types.HttpMethodValue_GET},
		}
		httpMatcher2 := &networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{
			Method: &networking_v1alpha1_types.TrafficPolicySpec_HttpMethod{Method: core_types.HttpMethodValue_POST},
		}
		httpMatcher3 := &networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{
			Method: &networking_v1alpha1_types.TrafficPolicySpec_HttpMethod{Method: core_types.HttpMethodValue_PUT},
		}
		tp1 := networking_v1alpha1.TrafficPolicy{
			Spec: networking_v1alpha1_types.TrafficPolicySpec{
				SourceSelector: &core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &core_types.ServiceSelector{
					ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
							Services: []*core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
							},
						},
					},
				},
				RequestTimeout:      &types1.Duration{Seconds: 1},
				HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher1},
			},
			Status: networking_v1alpha1_types.TrafficPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp2 := networking_v1alpha1.TrafficPolicy{
			Spec: networking_v1alpha1_types.TrafficPolicySpec{
				SourceSelector: &core_types.WorkloadSelector{
					Namespaces: destNamespaces2,
					Labels:     destLabels2,
				},
				DestinationSelector: &core_types.ServiceSelector{
					ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
							Services: []*core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
								{Name: meshServiceName2, Namespace: meshServiceNamespace2},
							},
						},
					},
				},
				RequestTimeout:      &types1.Duration{Seconds: 1},
				HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher2},
			},
			Status: networking_v1alpha1_types.TrafficPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp3 := networking_v1alpha1.TrafficPolicy{
			Spec: networking_v1alpha1_types.TrafficPolicySpec{
				SourceSelector: &core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &core_types.ServiceSelector{
					ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
							Services: []*core_types.ResourceRef{
								{Name: meshServiceName2, Namespace: meshServiceNamespace2},
							},
						},
					},
				},
				Retries:             &networking_v1alpha1_types.TrafficPolicySpec_RetryPolicy{Attempts: 2},
				HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher1, httpMatcher2, httpMatcher3},
			},
			Status: networking_v1alpha1_types.TrafficPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp4 := networking_v1alpha1.TrafficPolicy{
			Spec: networking_v1alpha1_types.TrafficPolicySpec{
				SourceSelector: &core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &core_types.ServiceSelector{
					ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
							Services: []*core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
							},
						},
					},
				},
				FaultInjection: &networking_v1alpha1_types.TrafficPolicySpec_FaultInjection{
					Percentage: 50,
				},
				HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher3},
			},
			Status: networking_v1alpha1_types.TrafficPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		tp5 := networking_v1alpha1.TrafficPolicy{
			Spec: networking_v1alpha1_types.TrafficPolicySpec{
				SourceSelector: &core_types.WorkloadSelector{
					Namespaces: destNamespaces1,
					Labels:     destLabels1,
				},
				DestinationSelector: &core_types.ServiceSelector{
					ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
							Services: []*core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
							},
						},
					},
				},
				FaultInjection: &networking_v1alpha1_types.TrafficPolicySpec_FaultInjection{
					Percentage: 50,
				},
				HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher1},
			},
			Status: networking_v1alpha1_types.TrafficPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_ACCEPTED,
					Message: "",
				},
			},
		}
		ignoredTP := networking_v1alpha1.TrafficPolicy{
			Spec: networking_v1alpha1_types.TrafficPolicySpec{
				DestinationSelector: &core_types.ServiceSelector{
					ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
						ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
							Services: []*core_types.ResourceRef{
								{Name: meshServiceName1, Namespace: meshServiceNamespace1},
								{Name: meshServiceName2, Namespace: meshServiceNamespace2},
							},
						},
					},
				},
			},
			Status: networking_v1alpha1_types.TrafficPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_CONFLICT,
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
		meshServices := []*v1alpha1.MeshService{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:        meshServiceName1,
					Namespace:   meshServiceNamespace1,
					ClusterName: clusterName1,
				},
				Spec: types.MeshServiceSpec{
					Mesh: &core_types.ResourceRef{
						Name:      meshName1,
						Namespace: meshNamespace1,
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:        meshServiceName2,
					Namespace:   meshServiceNamespace2,
					ClusterName: clusterName2,
				},
				Spec: types.MeshServiceSpec{
					Mesh: &core_types.ResourceRef{
						Name:      meshName2,
						Namespace: meshNamespace2,
					},
				},
			},
		}
		/*** GetMeshServicesByServiceSelector() ***/
		trafficPolicyList := &networking_v1alpha1.TrafficPolicyList{
			Items: []networking_v1alpha1.TrafficPolicy{tp1, tp2, tp3, tp4, tp5, ignoredTP}}
		mockTrafficPolicyClient.EXPECT().List(ctx).Return(trafficPolicyList, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp1.Spec.GetDestinationSelector()).
			Return([]*v1alpha1.MeshService{meshServices[0]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp2.Spec.GetDestinationSelector()).
			Return([]*v1alpha1.MeshService{meshServices[0], meshServices[1]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp3.Spec.GetDestinationSelector()).
			Return([]*v1alpha1.MeshService{meshServices[1]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp4.Spec.GetDestinationSelector()).
			Return([]*v1alpha1.MeshService{meshServices[0]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, tp5.Spec.GetDestinationSelector()).
			Return([]*v1alpha1.MeshService{meshServices[0]}, nil)
		mockResourceSelector.EXPECT().GetMeshServicesByServiceSelector(ctx, ignoredTP.Spec.GetDestinationSelector()).
			Return([]*v1alpha1.MeshService{meshServices[0], meshServices[1]}, nil)
		/*** buildKeyForMeshService ***/
		mesh1 := &v1alpha1.Mesh{Spec: types.MeshSpec{Cluster: &core_types.ResourceRef{Name: meshClusterName1}}}
		mesh2 := &v1alpha1.Mesh{Spec: types.MeshSpec{Cluster: &core_types.ResourceRef{Name: meshClusterName2}}}
		mockMeshClient.EXPECT().Get(ctx, client.ObjectKey{Name: meshName1, Namespace: meshNamespace1}).Return(mesh1, nil).Times(6)
		mockMeshClient.EXPECT().Get(ctx, client.ObjectKey{Name: meshName2, Namespace: meshNamespace2}).Return(mesh2, nil).Times(4)
		mergedTrafficPolicy1 := []*networking_v1alpha1.TrafficPolicy{
			{
				Spec: networking_v1alpha1_types.TrafficPolicySpec{
					SourceSelector: &core_types.WorkloadSelector{
						Namespaces: destNamespaces1,
						Labels:     destLabels1,
					},
					HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher1},
					RequestTimeout:      &types1.Duration{Seconds: 1},
					FaultInjection: &networking_v1alpha1_types.TrafficPolicySpec_FaultInjection{
						Percentage: 50,
					},
				},
			},
			{
				Spec: networking_v1alpha1_types.TrafficPolicySpec{
					SourceSelector: &core_types.WorkloadSelector{
						Namespaces: destNamespaces2,
						Labels:     destLabels2,
					},
					HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher2},
					RequestTimeout:      &types1.Duration{Seconds: 1},
				},
			},
			{
				Spec: networking_v1alpha1_types.TrafficPolicySpec{
					SourceSelector: &core_types.WorkloadSelector{
						Namespaces: destNamespaces1,
						Labels:     destLabels1,
					},
					HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher3},
					FaultInjection: &networking_v1alpha1_types.TrafficPolicySpec_FaultInjection{
						Percentage: 50,
					},
				},
			},
		}
		mergedTrafficPolicy2 := []*networking_v1alpha1.TrafficPolicy{
			{
				Spec: networking_v1alpha1_types.TrafficPolicySpec{
					SourceSelector: &core_types.WorkloadSelector{
						Namespaces: destNamespaces2,
						Labels:     destLabels2,
					},
					HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher2},
					RequestTimeout:      &types1.Duration{Seconds: 1},
				},
			},
			{
				Spec: networking_v1alpha1_types.TrafficPolicySpec{
					SourceSelector: &core_types.WorkloadSelector{
						Namespaces: destNamespaces1,
						Labels:     destLabels1,
					},
					HttpRequestMatchers: []*networking_v1alpha1_types.TrafficPolicySpec_HttpMatcher{httpMatcher1, httpMatcher2, httpMatcher3},
					Retries:             &networking_v1alpha1_types.TrafficPolicySpec_RetryPolicy{Attempts: 2},
				},
			},
		}
		policiesByMeshService, err := trafficPolicyMerger.MergeTrafficPoliciesForMeshServices(ctx, meshServices)
		Expect(err).To(BeNil())
		Expect(policiesByMeshService).To(HaveKeyWithValue(meshServiceKey1, mergedTrafficPolicy1))
		Expect(policiesByMeshService).To(HaveKeyWithValue(meshServiceKey2, mergedTrafficPolicy2))
	})
})
