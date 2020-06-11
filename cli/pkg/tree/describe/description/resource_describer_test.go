package description_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/describe/description"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/common/kube/selection/mocks"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Resource describer", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("DescribeService", func() {
		It("can find all the config that applies to a service", func() {
			trafficPolicyClient := mock_smh_networking.NewMockTrafficPolicyClient(ctrl)
			accessControlPolicyClient := mock_smh_networking.NewMockAccessControlPolicyClient(ctrl)
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			serviceName, serviceNs, serviceCluster := "test-svc", "test-ns", "test-cluster"
			correctServiceSelector := &smh_core_types.ServiceSelector{
				ServiceSelectorType: &smh_core_types.ServiceSelector_ServiceRefs_{
					ServiceRefs: &smh_core_types.ServiceSelector_ServiceRefs{
						Services: []*smh_core_types.ResourceRef{{Name: serviceName, Namespace: serviceNs, Cluster: serviceCluster}},
					},
				},
			}
			wrongServiceSelector := &smh_core_types.ServiceSelector{
				ServiceSelectorType: &smh_core_types.ServiceSelector_ServiceRefs_{
					ServiceRefs: &smh_core_types.ServiceSelector_ServiceRefs{
						Services: []*smh_core_types.ResourceRef{{Name: "other-name", Namespace: "other-ns", Cluster: serviceCluster}},
					},
				},
			}
			describedMeshService := &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "expected-mesh-service",
					Namespace: container_runtime.GetWriteNamespace(),
				},
			}

			accessControlPolices := &smh_networking.AccessControlPolicyList{
				Items: []smh_networking.AccessControlPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-1"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: wrongServiceSelector,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-2"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: correctServiceSelector,
						},
					},
				},
			}
			trafficPolicies := &smh_networking.TrafficPolicyList{
				Items: []smh_networking.TrafficPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp-1"},
						Spec: smh_networking_types.TrafficPolicySpec{
							DestinationSelector: wrongServiceSelector,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp-2"},
						Spec: smh_networking_types.TrafficPolicySpec{
							DestinationSelector: correctServiceSelector,
						},
					},
				},
			}

			resourceSelector.EXPECT().
				GetAllMeshServiceByRefSelector(ctx, serviceName, serviceNs, serviceCluster).
				Return(describedMeshService, nil)
			accessControlPolicyClient.EXPECT().
				ListAccessControlPolicy(ctx).
				Return(accessControlPolices, nil)
			trafficPolicyClient.EXPECT().
				ListTrafficPolicy(ctx).
				Return(trafficPolicies, nil)
			resourceSelector.EXPECT().
				GetAllMeshServicesByServiceSelector(ctx, wrongServiceSelector).
				Return([]*smh_discovery.MeshService{}, nil).
				Times(2)
			resourceSelector.EXPECT().
				GetAllMeshServicesByServiceSelector(ctx, correctServiceSelector).
				Return([]*smh_discovery.MeshService{describedMeshService}, nil).
				Times(2)

			describer := description.NewResourceDescriber(trafficPolicyClient, accessControlPolicyClient, resourceSelector)
			explorationResult, err := describer.DescribeService(ctx, description.FullyQualifiedKubeResource{
				Name:        serviceName,
				Namespace:   serviceNs,
				ClusterName: serviceCluster,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(explorationResult).To(Equal(&description.DescriptionResult{
				Policies: &description.Policies{
					AccessControlPolicies: []*smh_networking.AccessControlPolicy{&accessControlPolices.Items[1]},
					TrafficPolicies:       []*smh_networking.TrafficPolicy{&trafficPolicies.Items[1]},
				},
			}))
		})
	})

	Describe("DescribeWorkload", func() {
		It("can find all the config that applies to a service", func() {
			trafficPolicyClient := mock_smh_networking.NewMockTrafficPolicyClient(ctrl)
			accessControlPolicyClient := mock_smh_networking.NewMockAccessControlPolicyClient(ctrl)
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)

			controllerName, controllerNs, controllerCluster := "controller-name", "controller-ns", "controller-cluster"
			describedMeshWorkload := &smh_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: container_runtime.GetWriteNamespace(),
				},
			}
			wrongIdentitySelector := &smh_core_types.IdentitySelector{}
			correctIdentitySelector := &smh_core_types.IdentitySelector{
				IdentitySelectorType: &smh_core_types.IdentitySelector_ServiceAccountRefs_{},
			}
			wrongWorkloadSelector := &smh_core_types.WorkloadSelector{}
			correctWorkloadSelector := &smh_core_types.WorkloadSelector{
				Namespaces: []string{"doesn't-matter"},
			}
			accessControlPolices := &smh_networking.AccessControlPolicyList{
				Items: []smh_networking.AccessControlPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-1"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							SourceSelector: wrongIdentitySelector,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-2"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							SourceSelector: correctIdentitySelector,
						},
					},
				},
			}
			trafficPolicies := &smh_networking.TrafficPolicyList{
				Items: []smh_networking.TrafficPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp-1"},
						Spec: smh_networking_types.TrafficPolicySpec{
							SourceSelector: wrongWorkloadSelector,
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp-2"},
						Spec: smh_networking_types.TrafficPolicySpec{
							SourceSelector: correctWorkloadSelector,
						},
					},
				},
			}

			resourceSelector.EXPECT().
				GetMeshWorkloadByRefSelector(ctx, controllerName, controllerNs, controllerCluster).
				Return(describedMeshWorkload, nil)
			accessControlPolicyClient.EXPECT().
				ListAccessControlPolicy(ctx).
				Return(accessControlPolices, nil)
			trafficPolicyClient.EXPECT().
				ListTrafficPolicy(ctx).
				Return(trafficPolicies, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByIdentitySelector(ctx, wrongIdentitySelector).
				Return([]*smh_discovery.MeshWorkload{}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByIdentitySelector(ctx, correctIdentitySelector).
				Return([]*smh_discovery.MeshWorkload{describedMeshWorkload}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByWorkloadSelector(ctx, wrongWorkloadSelector).
				Return([]*smh_discovery.MeshWorkload{}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByWorkloadSelector(ctx, correctWorkloadSelector).
				Return([]*smh_discovery.MeshWorkload{describedMeshWorkload}, nil)

			describer := description.NewResourceDescriber(trafficPolicyClient, accessControlPolicyClient, resourceSelector)
			explorationResult, err := describer.DescribeWorkload(ctx, description.FullyQualifiedKubeResource{
				Name:        controllerName,
				Namespace:   controllerNs,
				ClusterName: controllerCluster,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(explorationResult).To(Equal(&description.DescriptionResult{
				Policies: &description.Policies{
					AccessControlPolicies: []*smh_networking.AccessControlPolicy{&accessControlPolices.Items[1]},
					TrafficPolicies:       []*smh_networking.TrafficPolicy{&trafficPolicies.Items[1]},
				},
			}))
		})
	})
})
