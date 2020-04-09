package access_control_policy_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_interactive "github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	mock_resource_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing/mocks"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	access_control_policy "github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/access-control-policy"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking/mocks"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("AccessControlPolicy", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		mockKubeLoader         *cli_mocks.MockKubeLoader
		mockMeshServiceClient  *mock_core.MockMeshServiceClient
		mockMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockACPClient          *mock_zephyr_networking.MockAccessControlPolicyClient
		mockInteractivePrompt  *mock_interactive.MockInteractivePrompt
		mockResourcePrinter    *mock_resource_printing.MockResourcePrinter
		meshctl                *cli_test.MockMeshctl
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockACPClient = mock_zephyr_networking.NewMockAccessControlPolicyClient(ctrl)
		mockInteractivePrompt = mock_interactive.NewMockInteractivePrompt(ctrl)
		mockResourcePrinter = mock_resource_printing.NewMockResourcePrinter(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				MeshWorkloadClient:        mockMeshWorkloadClient,
				MeshServiceClient:         mockMeshServiceClient,
				AccessControlPolicyClient: mockACPClient,
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

	It("should interactively create AccessControlPolicy using service acount refs as identity selector", func() {
		mockInteractivePrompt.
			EXPECT().
			SelectValue(gomock.Any(), []string{
				access_control_policy.MatcherSelectorOptionName,
				access_control_policy.RefSelectorOptionName,
			}).
			Return(access_control_policy.RefSelectorOptionName, nil)
		// select service accounts
		saNames := []string{"sa1", "sa2"}
		saDisplayNames := []string{"sa1.namespace1.cluster1", "sa2.namespace2.cluster2"}
		workloadList := &v1alpha1.MeshWorkloadList{
			Items: []v1alpha1.MeshWorkload{
				{
					Spec: types2.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Cluster: "cluster1",
						},
						KubeController: &types2.MeshWorkloadSpec_KubeController{
							ServiceAccountName: saNames[0],
							KubeControllerRef: &core_types.ResourceRef{
								Namespace: "namespace1",
							},
						},
					},
				},
				{
					Spec: types2.MeshWorkloadSpec{
						Mesh: &core_types.ResourceRef{
							Cluster: "cluster2",
						},
						KubeController: &types2.MeshWorkloadSpec_KubeController{
							ServiceAccountName: saNames[1],
							KubeControllerRef: &core_types.ResourceRef{
								Namespace: "namespace2",
							},
						},
					},
				},
			},
		}
		mockMeshWorkloadClient.
			EXPECT().
			List(ctx, gomock.Any()).
			Return(workloadList, nil)
		mockInteractivePrompt.
			EXPECT().
			SelectMultipleValues(gomock.Any(), saDisplayNames).
			Return([]string{saDisplayNames[1]}, nil)
		expectedIdentitySelector := &core_types.IdentitySelector{
			IdentitySelectorType: &core_types.IdentitySelector_ServiceAccountRefs_{
				ServiceAccountRefs: &core_types.IdentitySelector_ServiceAccountRefs{
					ServiceAccounts: []*core_types.ResourceRef{
						{Name: saNames[1], Namespace: "namespace2", Cluster: "cluster2"},
					},
				},
			},
		}
		// mesh services
		meshServiceNames := []string{"ms1", "ms2"}
		meshServiceDisplayNames := []string{"ms1.namespace1.cluster1", "ms2.namespace2.cluster2"}
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
		mockInteractivePrompt.
			EXPECT().
			SelectMultipleValues(gomock.Any(), meshServiceDisplayNames).
			Return([]string{meshServiceDisplayNames[0], meshServiceDisplayNames[1]}, nil)
		expectedTargetSelector := &core_types.ServiceSelector{
			ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
				ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
					Services: []*core_types.ResourceRef{
						{Name: meshServiceNames[0], Namespace: "namespace1"},
						{Name: meshServiceNames[1], Namespace: "namespace2"},
					},
				},
			},
		}
		// user inputs path
		mockInteractivePrompt.
			EXPECT().
			PromptValue(gomock.Any(), "").
			Return("/", nil)
		// user finishes inputting paths
		mockInteractivePrompt.
			EXPECT().
			PromptValue(gomock.Any(), "").
			Return("", nil)
		mockInteractivePrompt.
			EXPECT().
			SelectMultipleValues(gomock.Any(), access_control_policy.AllowedMethods).
			Return([]string{access_control_policy.AllowedMethods[0]}, nil)
		// user inputs port
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(gomock.Any(), "", gomock.Any()).
			Return("8080", nil)
		// user finishes inputting ports
		mockInteractivePrompt.
			EXPECT().
			PromptValueWithValidator(gomock.Any(), "", gomock.Any()).
			Return("", nil)

		expectedACP := &v1alpha12.AccessControlPolicy{
			TypeMeta: k8s_meta_v1.TypeMeta{
				Kind: "AccessControlPolicy"},
			Spec: types.AccessControlPolicySpec{
				SourceSelector:      expectedIdentitySelector,
				DestinationSelector: expectedTargetSelector,
				AllowedPaths:        []string{"/"},
				AllowedMethods:      []core_types.HttpMethodValue{core_types.HttpMethodValue_GET},
				AllowedPorts:        []uint32{8080},
			},
		}

		mockResourcePrinter.
			EXPECT().
			Print(gomock.Any(), expectedACP, resource_printing.JSONFormat).
			Return(nil)

		_, err := meshctl.Invoke("create --dry-run accesscontrolpolicy -o json")
		Expect(err).ToNot(HaveOccurred())
	})
})
