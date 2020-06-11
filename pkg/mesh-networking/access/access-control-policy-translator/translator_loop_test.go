package acp_translator_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/common/kube/selection/mocks"
	access_control_policy_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-policy-translator"
	mock_access_control_policy_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-policy-translator/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/smh/discovery"
	mock_smh_networking2 "github.com/solo-io/service-mesh-hub/test/mocks/smh/networking"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		acpEventWatcher           *mock_smh_networking2.MockAccessControlPolicyEventWatcher
		MeshServiceEventWatcher   *mock_smh_discovery.MockMeshServiceEventWatcher
		meshClient                *mock_core.MockMeshClient
		accessControlPolicyClient *mock_smh_networking.MockAccessControlPolicyClient
		resourceSelector          *mock_selector.MockResourceSelector
		mockMeshTranslator1       *mock_access_control_policy_translator.MockAcpMeshTranslator
		mockMeshTranslator2       *mock_access_control_policy_translator.MockAcpMeshTranslator
		meshTranslators           []*mock_access_control_policy_translator.MockAcpMeshTranslator

		// captured event handlers
		acpHandler         *smh_networking_controller.AccessControlPolicyEventHandlerFuncs
		meshServiceHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs
		acpTranslator      access_control_policy_translator.AcpTranslatorLoop
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		acpEventWatcher = mock_smh_networking2.NewMockAccessControlPolicyEventWatcher(ctrl)
		MeshServiceEventWatcher = mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		meshClient = mock_core.NewMockMeshClient(ctrl)
		accessControlPolicyClient = mock_smh_networking.NewMockAccessControlPolicyClient(ctrl)
		resourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		mockMeshTranslator1 = mock_access_control_policy_translator.NewMockAcpMeshTranslator(ctrl)
		mockMeshTranslator2 = mock_access_control_policy_translator.NewMockAcpMeshTranslator(ctrl)
		meshTranslators = []*mock_access_control_policy_translator.MockAcpMeshTranslator{
			mockMeshTranslator1,
			mockMeshTranslator2,
		}
		acpTranslator = access_control_policy_translator.NewAcpTranslatorLoop(
			acpEventWatcher,
			MeshServiceEventWatcher,
			meshClient,
			accessControlPolicyClient,
			resourceSelector,
			[]access_control_policy_translator.AcpMeshTranslator{
				mockMeshTranslator1, mockMeshTranslator2,
			},
		)
		acpEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_networking_controller.AccessControlPolicyEventHandlerFuncs) error {
				acpHandler = eventHandler
				return nil
			})
		MeshServiceEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {

				meshServiceHandler = eventHandler
				return nil
			})
		acpTranslator.Start(ctx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var acp = func() *smh_networking.AccessControlPolicy {
		return &smh_networking.AccessControlPolicy{
			Spec: smh_networking_types.AccessControlPolicySpec{
				DestinationSelector: &smh_core_types.ServiceSelector{},
			},
		}
	}

	Describe("It should handle AccessControlPolicy events", func() {
		It("should handle AccessControlPolicy create", func() {
			acp := acp()
			matchingMeshServices := []*smh_discovery.MeshService{
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-name-1",
							Namespace: "mesh-namespace-1",
						},
					},
				},
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-name-2",
							Namespace: "mesh-namespace-2",
						},
					},
				},
			}
			meshesForService := []*smh_discovery.Mesh{
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-1", Namespace: "mesh-namespace-1"}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-22", Namespace: "mesh-namespace-22"}},
			}
			resourceSelector.
				EXPECT().
				GetAllMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
				Return(matchingMeshServices, nil)
			var expectedTargetServices []access_control_policy_translator.TargetService
			for i, meshService := range matchingMeshServices {
				meshClient.
					EXPECT().
					GetMesh(ctx, selection.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
					Return(meshesForService[i], nil)
				expectedTargetServices = append(
					expectedTargetServices,
					access_control_policy_translator.TargetService{
						MeshService: meshService,
						Mesh:        meshesForService[i],
					},
				)
			}
			for _, meshTranslator := range meshTranslators {
				meshTranslator.EXPECT().Translate(contextutils.WithLogger(ctx, ""), expectedTargetServices, acp).Return(nil)
				meshTranslator.EXPECT().Name().Return("")
			}
			var capturedACPWithStatus *smh_networking.AccessControlPolicy
			accessControlPolicyClient.
				EXPECT().
				UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, acp *smh_networking.AccessControlPolicy) error {
					capturedACPWithStatus = acp
					return nil
				})
			expectedStatus := smh_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &smh_core_types.Status{
					State: smh_core_types.Status_ACCEPTED,
				},
				TranslatorErrors: nil,
			}
			err := acpHandler.OnCreate(acp)
			Expect(err).ToNot(HaveOccurred())
			Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
		})

		It("should aggregate list of translator errors", func() {
			acp := acp()
			matchingMeshServices := []*smh_discovery.MeshService{
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-name-1",
							Namespace: "mesh-namespace-1",
						},
					},
				},
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh-name-2",
							Namespace: "mesh-namespace-2",
						},
					},
				},
			}
			meshesForService := []*smh_discovery.Mesh{
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-1", Namespace: "mesh-namespace-1"}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-22", Namespace: "mesh-namespace-22"}},
			}
			resourceSelector.
				EXPECT().
				GetAllMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
				Return(matchingMeshServices, nil)
			var expectedTargetServices []access_control_policy_translator.TargetService
			for i, meshService := range matchingMeshServices {
				meshClient.
					EXPECT().
					GetMesh(ctx, selection.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
					Return(meshesForService[i], nil)
				expectedTargetServices = append(
					expectedTargetServices,
					access_control_policy_translator.TargetService{
						MeshService: meshService,
						Mesh:        meshesForService[i],
					},
				)
			}
			var translatorErrors []*smh_networking_types.AccessControlPolicyStatus_TranslatorError
			for _, meshTranslator := range meshTranslators {
				translatorErr := &smh_networking_types.AccessControlPolicyStatus_TranslatorError{
					TranslatorId: "translator-id",
					ErrorMessage: "",
				}
				translatorErrors = append(translatorErrors, translatorErr)
				meshTranslator.
					EXPECT().
					Translate(contextutils.WithLogger(ctx, ""), expectedTargetServices, acp).
					Return(translatorErr)
				meshTranslator.EXPECT().Name().Return("")
			}
			var capturedACPWithStatus *smh_networking.AccessControlPolicy
			accessControlPolicyClient.
				EXPECT().
				UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, acp *smh_networking.AccessControlPolicy) error {
					capturedACPWithStatus = acp
					return nil
				})
			expectedStatus := smh_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
				},
				TranslatorErrors: translatorErrors,
			}
			err := acpHandler.OnCreate(acp)
			Expect(err).ToNot(HaveOccurred())
			Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
		})
	})

	Describe("It should handle MeshService events", func() {
		It("should handle MeshService create", func() {
			meshService := &smh_discovery.MeshService{
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: &smh_core_types.ResourceRef{
						Name:      "mesh-name-1",
						Namespace: "mesh-namespace-1",
					},
				},
			}
			mesh := &smh_discovery.Mesh{
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "cluster-name",
					},
				},
			}
			targetService := access_control_policy_translator.TargetService{
				MeshService: meshService,
				Mesh:        mesh,
			}
			meshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
				Return(mesh, nil).
				Times(5)
			acpList := &smh_networking.AccessControlPolicyList{
				Items: []smh_networking.AccessControlPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-1"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: &smh_core_types.ServiceSelector{
								ServiceSelectorType: &smh_core_types.ServiceSelector_Matcher_{
									Matcher: &smh_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-1"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-2"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: &smh_core_types.ServiceSelector{
								ServiceSelectorType: &smh_core_types.ServiceSelector_Matcher_{
									Matcher: &smh_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-2"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-3"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: &smh_core_types.ServiceSelector{
								ServiceSelectorType: &smh_core_types.ServiceSelector_Matcher_{
									Matcher: &smh_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-3"},
									},
								},
							},
						},
					},
				},
			}
			accessControlPolicyClient.
				EXPECT().
				ListAccessControlPolicy(ctx).
				Return(acpList, nil)
			var capturedACPsWithStatus []*smh_networking.AccessControlPolicy
			for _, acp := range acpList.Items {
				acp := acp
				resourceSelector.
					EXPECT().
					GetAllMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
					Return([]*smh_discovery.MeshService{meshService}, nil)
				for _, meshTranslator := range meshTranslators {
					meshTranslator.
						EXPECT().
						Translate(contextutils.WithLogger(ctx, ""), []access_control_policy_translator.TargetService{targetService}, &acp).
						Return(nil)
					meshTranslator.EXPECT().Name().Return("")
				}
				accessControlPolicyClient.
					EXPECT().
					UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, acp *smh_networking.AccessControlPolicy) error {
						capturedACPsWithStatus = append(capturedACPsWithStatus, acp)
						return nil
					})
			}
			err := meshServiceHandler.OnCreate(meshService)
			Expect(err).ToNot(HaveOccurred())
			expectedStatus := smh_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &smh_core_types.Status{
					State: smh_core_types.Status_ACCEPTED,
				},
				TranslatorErrors: nil,
			}
			for _, capturedACPWithStatus := range capturedACPsWithStatus {
				Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
			}
		})

		It("should aggregate translator errors for each applicable ACP for MeshService", func() {
			meshService := &smh_discovery.MeshService{
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: &smh_core_types.ResourceRef{
						Name:      "mesh-name-1",
						Namespace: "mesh-namespace-1",
					},
				},
			}
			mesh := &smh_discovery.Mesh{
				Spec: smh_discovery_types.MeshSpec{
					Cluster: &smh_core_types.ResourceRef{
						Name: "cluster-name",
					},
				},
			}
			targetService := access_control_policy_translator.TargetService{
				MeshService: meshService,
				Mesh:        mesh,
			}
			meshClient.
				EXPECT().
				GetMesh(ctx, selection.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
				Return(mesh, nil).
				Times(5)
			acpList := &smh_networking.AccessControlPolicyList{
				Items: []smh_networking.AccessControlPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-1"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: &smh_core_types.ServiceSelector{
								ServiceSelectorType: &smh_core_types.ServiceSelector_Matcher_{
									Matcher: &smh_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-1"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-2"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: &smh_core_types.ServiceSelector{
								ServiceSelectorType: &smh_core_types.ServiceSelector_Matcher_{
									Matcher: &smh_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-2"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-3"},
						Spec: smh_networking_types.AccessControlPolicySpec{
							DestinationSelector: &smh_core_types.ServiceSelector{
								ServiceSelectorType: &smh_core_types.ServiceSelector_Matcher_{
									Matcher: &smh_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-3"},
									},
								},
							},
						},
					},
				},
			}
			accessControlPolicyClient.
				EXPECT().
				ListAccessControlPolicy(ctx).
				Return(acpList, nil)
			var newTranslatorError = func() *smh_networking_types.AccessControlPolicyStatus_TranslatorError {
				return &smh_networking_types.AccessControlPolicyStatus_TranslatorError{
					TranslatorId: "translator-id",
					ErrorMessage: "",
				}
			}
			var capturedACPsWithStatus []*smh_networking.AccessControlPolicy
			for _, acp := range acpList.Items {
				acp := acp
				resourceSelector.
					EXPECT().
					GetAllMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
					Return([]*smh_discovery.MeshService{meshService}, nil)
				for _, meshTranslator := range meshTranslators {
					meshTranslator.
						EXPECT().
						Translate(contextutils.WithLogger(ctx, ""), []access_control_policy_translator.TargetService{targetService}, &acp).
						Return(newTranslatorError())
					meshTranslator.EXPECT().Name().Return("")
				}
				accessControlPolicyClient.
					EXPECT().
					UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, acp *smh_networking.AccessControlPolicy) error {
						capturedACPsWithStatus = append(capturedACPsWithStatus, acp)
						return nil
					})
			}
			err := meshServiceHandler.OnCreate(meshService)
			Expect(err).ToNot(HaveOccurred())
			var expectedTranslatorErrors []*smh_networking_types.AccessControlPolicyStatus_TranslatorError
			for range meshTranslators {
				expectedTranslatorErrors = append(expectedTranslatorErrors, newTranslatorError())
			}
			expectedStatus := smh_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &smh_core_types.Status{
					State:   smh_core_types.Status_PROCESSING_ERROR,
					Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
				},
				TranslatorErrors: expectedTranslatorErrors,
			}
			for _, capturedACPWithStatus := range capturedACPsWithStatus {
				Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
			}
		})
	})
})
