package virtualmesh_test

import (
	"context"
	"strconv"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	mock_interactive "github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive/mocks"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	mock_resource_printing "github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing/mocks"
	cli_mocks "github.com/solo-io/service-mesh-hub/cli/pkg/mocks"
	cli_test "github.com/solo-io/service-mesh-hub/cli/pkg/test"
	"github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("VirtualMeshCmd", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		mockKubeLoader        *cli_mocks.MockKubeLoader
		mockMeshClient        *mock_core.MockMeshClient
		mockVirtualMeshClient *mock_zephyr_networking.MockVirtualMeshClient
		mockInteractivePrompt *mock_interactive.MockInteractivePrompt
		mockResourcePrinter   *mock_resource_printing.MockResourcePrinter
		meshctl               *cli_test.MockMeshctl
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockKubeLoader = cli_mocks.NewMockKubeLoader(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockVirtualMeshClient = mock_zephyr_networking.NewMockVirtualMeshClient(ctrl)
		mockInteractivePrompt = mock_interactive.NewMockInteractivePrompt(ctrl)
		mockResourcePrinter = mock_resource_printing.NewMockResourcePrinter(ctrl)
		meshctl = &cli_test.MockMeshctl{
			MockController: ctrl,
			Clients:        common.Clients{},
			KubeClients: common.KubeClients{
				VirtualMeshClient: mockVirtualMeshClient,
				MeshClient:        mockMeshClient,
			},
			Printers: common.Printers{
				ResourcePrinter: mockResourcePrinter,
			},
			KubeLoader:        mockKubeLoader,
			Ctx:               ctx,
			InteractivePrompt: mockInteractivePrompt,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("should operate interactively", func() {
		It("should interactively create client", func() {
			displayName := "display-name"
			targetRestConfig := &rest.Config{}
			meshList := &v1alpha1.MeshList{Items: []v1alpha1.Mesh{
				{ObjectMeta: metav1.ObjectMeta{Name: "name1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "name2"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "name3"}},
			}}
			expectedMeshRefs := []*types.ResourceRef{
				{Name: meshList.Items[0].GetName(), Namespace: env.GetWriteNamespace()},
				{Name: meshList.Items[2].GetName(), Namespace: env.GetWriteNamespace()},
			}
			expectedCA := &networking_types.VirtualMeshSpec_CertificateAuthority{
				Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
					Builtin: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
						TtlDays:         365,
						RsaKeySizeBytes: 4096,
						OrgName:         "orgName",
					},
				},
			}
			expectedVM := &networking_v1alpha1.VirtualMesh{
				TypeMeta: metav1.TypeMeta{Kind: "VirtualMesh"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      displayName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: networking_types.VirtualMeshSpec{
					DisplayName:          displayName,
					Meshes:               expectedMeshRefs,
					CertificateAuthority: expectedCA,
					Federation: &networking_types.VirtualMeshSpec_Federation{
						Mode: 0,
					},
					TrustModel: &networking_types.VirtualMeshSpec_Shared{
						Shared: &networking_types.VirtualMeshSpec_SharedTrust{},
					},
				},
			}

			mockKubeLoader.EXPECT().GetRestConfigForContext("", "").Return(targetRestConfig, nil)
			mockMeshClient.EXPECT().List(ctx).Return(meshList, nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(displayName, nil)
			mockInteractivePrompt.
				EXPECT().
				SelectMultipleValues(gomock.Any(), []string{meshList.Items[0].GetName(), meshList.Items[1].GetName(), meshList.Items[2].GetName()}).
				Return([]string{expectedMeshRefs[0].GetName(), expectedMeshRefs[1].GetName()}, nil)
			mockInteractivePrompt.
				EXPECT().
				SelectValue(gomock.Any(), gomock.Any()).
				Return("builtin", nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(strconv.Itoa(int(expectedCA.GetBuiltin().GetTtlDays())), nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(strconv.Itoa(int(expectedCA.GetBuiltin().GetRsaKeySizeBytes())), nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValue(gomock.Any(), gomock.Any()).
				Return(expectedCA.GetBuiltin().GetOrgName(), nil)

			mockVirtualMeshClient.EXPECT().Create(ctx, expectedVM).Return(nil)

			_, err := meshctl.Invoke("create virtualmesh")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should interactively generate client and output for dry-run", func() {
			displayName := "display-name"
			targetRestConfig := &rest.Config{}
			meshList := &v1alpha1.MeshList{Items: []v1alpha1.Mesh{
				{ObjectMeta: metav1.ObjectMeta{Name: "name1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "name2"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "name3"}},
			}}
			expectedMeshRefs := []*types.ResourceRef{
				{Name: meshList.Items[0].GetName(), Namespace: env.GetWriteNamespace()},
				{Name: meshList.Items[2].GetName(), Namespace: env.GetWriteNamespace()},
			}
			expectedCA := &networking_types.VirtualMeshSpec_CertificateAuthority{
				Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
					Builtin: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
						TtlDays:         365,
						RsaKeySizeBytes: 4096,
						OrgName:         "orgName",
					},
				},
			}
			expectedVM := &networking_v1alpha1.VirtualMesh{
				TypeMeta: metav1.TypeMeta{Kind: "VirtualMesh"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      displayName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: networking_types.VirtualMeshSpec{
					DisplayName:          displayName,
					Meshes:               expectedMeshRefs,
					CertificateAuthority: expectedCA,
					Federation: &networking_types.VirtualMeshSpec_Federation{
						Mode: 0,
					},
					TrustModel: &networking_types.VirtualMeshSpec_Shared{
						Shared: &networking_types.VirtualMeshSpec_SharedTrust{},
					},
				},
			}

			mockKubeLoader.EXPECT().GetRestConfigForContext("", "").Return(targetRestConfig, nil)
			mockMeshClient.EXPECT().List(ctx).Return(meshList, nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(displayName, nil)
			mockInteractivePrompt.
				EXPECT().
				SelectMultipleValues(gomock.Any(), []string{meshList.Items[0].GetName(), meshList.Items[1].GetName(), meshList.Items[2].GetName()}).
				Return([]string{expectedMeshRefs[0].GetName(), expectedMeshRefs[1].GetName()}, nil)
			mockInteractivePrompt.
				EXPECT().
				SelectValue(gomock.Any(), gomock.Any()).
				Return("builtin", nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(strconv.Itoa(int(expectedCA.GetBuiltin().GetTtlDays())), nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(strconv.Itoa(int(expectedCA.GetBuiltin().GetRsaKeySizeBytes())), nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValue(gomock.Any(), gomock.Any()).
				Return(expectedCA.GetBuiltin().GetOrgName(), nil)

			mockResourcePrinter.EXPECT().Print(gomock.Any(), expectedVM, resource_printing.JSONFormat).Return(nil)

			_, err := meshctl.Invoke("create --dry-run virtualmesh -o json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should interactively generate client and output for dry-run", func() {
			displayName := "display-name"
			targetRestConfig := &rest.Config{}
			meshList := &v1alpha1.MeshList{Items: []v1alpha1.Mesh{
				{ObjectMeta: metav1.ObjectMeta{Name: "name1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "name2"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "name3"}},
			}}
			expectedMeshRefs := []*types.ResourceRef{
				{Name: meshList.Items[0].GetName(), Namespace: env.GetWriteNamespace()},
				{Name: meshList.Items[2].GetName(), Namespace: env.GetWriteNamespace()},
			}
			expectedCA := &networking_types.VirtualMeshSpec_CertificateAuthority{
				Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
					Builtin: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
						TtlDays:         365,
						RsaKeySizeBytes: 4096,
						OrgName:         "orgName",
					},
				},
			}
			expectedVM := &networking_v1alpha1.VirtualMesh{
				TypeMeta: metav1.TypeMeta{Kind: "VirtualMesh"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      displayName,
					Namespace: env.GetWriteNamespace(),
				},
				Spec: networking_types.VirtualMeshSpec{
					DisplayName:          displayName,
					Meshes:               expectedMeshRefs,
					CertificateAuthority: expectedCA,
					Federation: &networking_types.VirtualMeshSpec_Federation{
						Mode: 0,
					},
					TrustModel: &networking_types.VirtualMeshSpec_Shared{
						Shared: &networking_types.VirtualMeshSpec_SharedTrust{},
					},
				},
			}

			mockKubeLoader.EXPECT().GetRestConfigForContext("", "").Return(targetRestConfig, nil)
			mockMeshClient.EXPECT().List(ctx).Return(meshList, nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(displayName, nil)
			mockInteractivePrompt.
				EXPECT().
				SelectMultipleValues(gomock.Any(), []string{meshList.Items[0].GetName(), meshList.Items[1].GetName(), meshList.Items[2].GetName()}).
				Return([]string{expectedMeshRefs[0].GetName(), expectedMeshRefs[1].GetName()}, nil)
			mockInteractivePrompt.
				EXPECT().
				SelectValue(gomock.Any(), gomock.Any()).
				Return("builtin", nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(strconv.Itoa(int(expectedCA.GetBuiltin().GetTtlDays())), nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValueWithValidator(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(strconv.Itoa(int(expectedCA.GetBuiltin().GetRsaKeySizeBytes())), nil)
			mockInteractivePrompt.
				EXPECT().
				PromptValue(gomock.Any(), gomock.Any()).
				Return(expectedCA.GetBuiltin().GetOrgName(), nil)

			mockResourcePrinter.EXPECT().Print(gomock.Any(), expectedVM, resource_printing.JSONFormat).Return(nil)

			_, err := meshctl.Invoke("create --dry-run virtualmesh -o json")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
