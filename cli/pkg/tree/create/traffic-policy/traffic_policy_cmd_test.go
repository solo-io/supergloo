package traffic_policy_test

import (
	"context"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/pkg/utils"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_interactive "github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	mock_resource_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing/mocks"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	traffic_policy "github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/traffic-policy"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking/mocks"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
)

var _ = Describe("TrafficPolicyCmd", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		mockKubeLoader          *cli_mocks.MockKubeLoader
		mockMeshServiceClient   *mock_core.MockMeshServiceClient
		mockTrafficPolicyClient *mock_zephyr_networking.MockTrafficPolicyClient
		mockInteractivePrompt   *mock_interactive.MockInteractivePrompt
		mockResourcePrinter     *mock_resource_printing.MockResourcePrinter
		meshctl                 *cli_test.MockMeshctl
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockTrafficPolicyClient = mock_zephyr_networking.NewMockTrafficPolicyClient(ctrl)
		mockInteractivePrompt = mock_interactive.NewMockInteractivePrompt(ctrl)
		mockResourcePrinter = mock_resource_printing.NewMockResourcePrinter(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				MeshServiceClient:   mockMeshServiceClient,
				TrafficPolicyClient: mockTrafficPolicyClient,
			},
			KubeLoader:        mockKubeLoader,
			Ctx:               ctx,
			InteractivePrompt: mockInteractivePrompt,
			Printers:          common.Printers{ResourcePrinter: mockResourcePrinter},
		}
		targetRestConfig := &rest.Config{}
		mockKubeLoader.EXPECT().GetRestConfigForContext("", "").Return(targetRestConfig, nil)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should interactively create TrafficPolicy", func() {
		// mesh services
		meshServiceNames := []string{"ms1", "ms2"}
		meshServiceDisplayNames := []string{"ms1.namespace1.cluster1", "ms2.namespace2.cluster2"}
		meshServiceDisplayNamesWithDoneOption := append([]string{traffic_policy.DoneSelectingOption}, meshServiceDisplayNames...)
		meshServiceList := &v1alpha1.MeshServiceList{
			Items: []v1alpha1.MeshService{
				{
					ObjectMeta: k8s_meta_v1.ObjectMeta{
						Name:      meshServiceNames[0],
						Namespace: "namespace1",
					},
					Spec: types2.MeshServiceSpec{
						KubeService: &types2.MeshServiceSpec_KubeService{
							Ref: &core_types.ResourceRef{Cluster: "cluster1"},
						},
					},
				},
				{
					ObjectMeta: k8s_meta_v1.ObjectMeta{
						Name:      meshServiceNames[1],
						Namespace: "namespace2",
					},
					Spec: types2.MeshServiceSpec{
						KubeService: &types2.MeshServiceSpec_KubeService{
							Ref: &core_types.ResourceRef{Cluster: "cluster2"},
						},
					},
				},
			},
		}
		mockMeshServiceClient.
			EXPECT().
			List(ctx).
			Return(meshServiceList, nil)
		// select sources
		expectedLabels := labels.Set(map[string]string{"k1": "v1", "k2": "v2"})
		expectedNamespaces := []string{"namespace1", "namespace2"}
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(gomock.Any(), "", gomock.Any()).
			Return("k1=v1, k2=v2", nil)
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(gomock.Any(), "", gomock.Any()).
			Return(strings.Join(expectedNamespaces, ","), nil)
		expectedWorkloadSelector := &core_types.WorkloadSelector{Labels: expectedLabels, Namespaces: expectedNamespaces}
		// select targets
		mockInteractivePrompt.
			EXPECT().
			SelectMultipleValues(gomock.Any(), meshServiceDisplayNames).
			Return([]string{meshServiceDisplayNames[0]}, nil)
		expectedTargetSelector := &core_types.ServiceSelector{
			ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
				ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
					Services: []*core_types.ResourceRef{
						{
							Name:      meshServiceList.Items[0].GetName(),
							Namespace: meshServiceList.Items[0].GetNamespace(),
						},
					},
				},
			},
		}
		// select TrafficShift
		mockInteractivePrompt.
			EXPECT().
			SelectValue(gomock.Any(), meshServiceDisplayNamesWithDoneOption).
			Return(meshServiceDisplayNames[1], nil)
		// select port
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("8080", nil)
		// select weight
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("1", nil)
		// select subsets
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", nil)
		// finish selecting
		mockInteractivePrompt.
			EXPECT().
			SelectValue(gomock.Any(), utils.RemoveString(meshServiceDisplayNamesWithDoneOption, meshServiceDisplayNames[1])).
			Return(traffic_policy.DoneSelectingOption, nil)
		expectedTrafficShift := &networking_types.TrafficPolicySpec_MultiDestination{
			Destinations: []*networking_types.TrafficPolicySpec_MultiDestination_WeightedDestination{
				{
					Destination: &core_types.ResourceRef{
						Name:      meshServiceList.Items[1].GetName(),
						Namespace: meshServiceList.Items[1].GetNamespace(),
					},
					Port:   8080,
					Weight: 1,
				},
			},
		}
		expectedTrafficPolicy := &networking_v1alpha1.TrafficPolicy{
			TypeMeta: k8s_meta_v1.TypeMeta{
				Kind: "TrafficPolicy",
			},
			Spec: networking_types.TrafficPolicySpec{
				SourceSelector:      expectedWorkloadSelector,
				DestinationSelector: expectedTargetSelector,
				TrafficShift:        expectedTrafficShift,
			},
		}

		mockResourcePrinter.
			EXPECT().
			Print(gomock.Any(), expectedTrafficPolicy, resource_printing.JSONFormat).
			Return(nil)

		_, err := meshctl.Invoke("create --dry-run trafficpolicy -o json")
		Expect(err).ToNot(HaveOccurred())
	})
})
