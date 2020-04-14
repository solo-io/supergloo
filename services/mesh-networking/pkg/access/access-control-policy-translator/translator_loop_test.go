package acp_translator_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/contextutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	access_control_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator"
	mock_access_control_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	mock_zephyr_networking2 "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/networking"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		acpController             *mock_zephyr_networking2.MockAccessControlPolicyEventWatcher
		MeshServiceEventWatcher   *mock_zephyr_discovery.MockMeshServiceEventWatcher
		meshClient                *mock_core.MockMeshClient
		accessControlPolicyClient *mock_zephyr_networking.MockAccessControlPolicyClient
		resourceSelector          *mock_selector.MockResourceSelector
		mockMeshTranslator1       *mock_access_control_policy_translator.MockAcpMeshTranslator
		mockMeshTranslator2       *mock_access_control_policy_translator.MockAcpMeshTranslator
		meshTranslators           []*mock_access_control_policy_translator.MockAcpMeshTranslator

		// captured event handlers
		acpHandler         *networking_controller.AccessControlPolicyEventHandlerFuncs
		meshServiceHandler *discovery_controller.MeshServiceEventHandlerFuncs
		acpTranslator      access_control_policy_translator.AcpTranslatorLoop
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		acpController = mock_zephyr_networking2.NewMockAccessControlPolicyEventWatcher(ctrl)
		MeshServiceEventWatcher = mock_zephyr_discovery.NewMockMeshServiceEventWatcher(ctrl)
		meshClient = mock_core.NewMockMeshClient(ctrl)
		accessControlPolicyClient = mock_zephyr_networking.NewMockAccessControlPolicyClient(ctrl)
		resourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		mockMeshTranslator1 = mock_access_control_policy_translator.NewMockAcpMeshTranslator(ctrl)
		mockMeshTranslator2 = mock_access_control_policy_translator.NewMockAcpMeshTranslator(ctrl)
		meshTranslators = []*mock_access_control_policy_translator.MockAcpMeshTranslator{
			mockMeshTranslator1,
			mockMeshTranslator2,
		}
		acpTranslator = access_control_policy_translator.NewAcpTranslatorLoop(
			acpController,
			MeshServiceEventWatcher,
			meshClient,
			accessControlPolicyClient,
			resourceSelector,
			[]access_control_policy_translator.AcpMeshTranslator{
				mockMeshTranslator1, mockMeshTranslator2,
			},
		)
		acpController.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *networking_controller.AccessControlPolicyEventHandlerFuncs) error {
				acpHandler = eventHandler
				return nil
			})
		MeshServiceEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *discovery_controller.MeshServiceEventHandlerFuncs) error {

				meshServiceHandler = eventHandler
				return nil
			})
		acpTranslator.Start(ctx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var acp = func() *networking_v1alpha1.AccessControlPolicy {
		return &networking_v1alpha1.AccessControlPolicy{
			Spec: networking_types.AccessControlPolicySpec{
				DestinationSelector: &core_types.ServiceSelector{},
			},
		}
	}

	Describe("It should handle AccessControlPolicy events", func() {
		It("should handle AccessControlPolicy create", func() {
			acp := acp()
			matchingMeshServices := []*discovery_v1alpha1.MeshService{
				{
					Spec: discovery_types.MeshServiceSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-name-1",
							Namespace: "mesh-namespace-1",
						},
					},
				},
				{
					Spec: discovery_types.MeshServiceSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-name-2",
							Namespace: "mesh-namespace-2",
						},
					},
				},
			}
			meshesForService := []*discovery_v1alpha1.Mesh{
				{ObjectMeta: metav1.ObjectMeta{Name: "mesh-name-1", Namespace: "mesh-namespace-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "mesh-name-22", Namespace: "mesh-namespace-22"}},
			}
			resourceSelector.
				EXPECT().
				GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
				Return(matchingMeshServices, nil)
			var expectedTargetServices []access_control_policy_translator.TargetService
			for i, meshService := range matchingMeshServices {
				meshClient.
					EXPECT().
					GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
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
			var capturedACPWithStatus *networking_v1alpha1.AccessControlPolicy
			accessControlPolicyClient.
				EXPECT().
				UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, acp *networking_v1alpha1.AccessControlPolicy) error {
					capturedACPWithStatus = acp
					return nil
				})
			expectedStatus := networking_types.AccessControlPolicyStatus{
				TranslationStatus: &core_types.Status{
					State: core_types.Status_ACCEPTED,
				},
				TranslatorErrors: nil,
			}
			err := acpHandler.OnCreate(acp)
			Expect(err).ToNot(HaveOccurred())
			Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
		})

		It("should aggregate list of translator errors", func() {
			acp := acp()
			matchingMeshServices := []*discovery_v1alpha1.MeshService{
				{
					Spec: discovery_types.MeshServiceSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-name-1",
							Namespace: "mesh-namespace-1",
						},
					},
				},
				{
					Spec: discovery_types.MeshServiceSpec{
						Mesh: &core_types.ResourceRef{
							Name:      "mesh-name-2",
							Namespace: "mesh-namespace-2",
						},
					},
				},
			}
			meshesForService := []*discovery_v1alpha1.Mesh{
				{ObjectMeta: metav1.ObjectMeta{Name: "mesh-name-1", Namespace: "mesh-namespace-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "mesh-name-22", Namespace: "mesh-namespace-22"}},
			}
			resourceSelector.
				EXPECT().
				GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
				Return(matchingMeshServices, nil)
			var expectedTargetServices []access_control_policy_translator.TargetService
			for i, meshService := range matchingMeshServices {
				meshClient.
					EXPECT().
					GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
					Return(meshesForService[i], nil)
				expectedTargetServices = append(
					expectedTargetServices,
					access_control_policy_translator.TargetService{
						MeshService: meshService,
						Mesh:        meshesForService[i],
					},
				)
			}
			var translatorErrors []*networking_types.AccessControlPolicyStatus_TranslatorError
			for _, meshTranslator := range meshTranslators {
				translatorErr := &networking_types.AccessControlPolicyStatus_TranslatorError{
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
			var capturedACPWithStatus *networking_v1alpha1.AccessControlPolicy
			accessControlPolicyClient.
				EXPECT().
				UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, acp *networking_v1alpha1.AccessControlPolicy) error {
					capturedACPWithStatus = acp
					return nil
				})
			expectedStatus := networking_types.AccessControlPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_PROCESSING_ERROR,
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
			meshService := &discovery_v1alpha1.MeshService{
				Spec: discovery_types.MeshServiceSpec{
					Mesh: &core_types.ResourceRef{
						Name:      "mesh-name-1",
						Namespace: "mesh-namespace-1",
					},
				},
			}
			mesh := &discovery_v1alpha1.Mesh{
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
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
				GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
				Return(mesh, nil).
				Times(5)
			acpList := &networking_v1alpha1.AccessControlPolicyList{
				Items: []networking_v1alpha1.AccessControlPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-name-1"},
						Spec: networking_types.AccessControlPolicySpec{
							DestinationSelector: &core_types.ServiceSelector{
								ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
									Matcher: &core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-1"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-name-2"},
						Spec: networking_types.AccessControlPolicySpec{
							DestinationSelector: &core_types.ServiceSelector{
								ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
									Matcher: &core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-2"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-name-3"},
						Spec: networking_types.AccessControlPolicySpec{
							DestinationSelector: &core_types.ServiceSelector{
								ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
									Matcher: &core_types.ServiceSelector_Matcher{
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
			var capturedACPsWithStatus []*networking_v1alpha1.AccessControlPolicy
			for _, acp := range acpList.Items {
				acp := acp
				resourceSelector.
					EXPECT().
					GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
					Return([]*discovery_v1alpha1.MeshService{meshService}, nil)
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
					DoAndReturn(func(ctx context.Context, acp *networking_v1alpha1.AccessControlPolicy) error {
						capturedACPsWithStatus = append(capturedACPsWithStatus, acp)
						return nil
					})
			}
			err := meshServiceHandler.OnCreate(meshService)
			Expect(err).ToNot(HaveOccurred())
			expectedStatus := networking_types.AccessControlPolicyStatus{
				TranslationStatus: &core_types.Status{
					State: core_types.Status_ACCEPTED,
				},
				TranslatorErrors: nil,
			}
			for _, capturedACPWithStatus := range capturedACPsWithStatus {
				Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
			}
		})

		It("should aggregate translator errors for each applicable ACP for MeshService", func() {
			meshService := &discovery_v1alpha1.MeshService{
				Spec: discovery_types.MeshServiceSpec{
					Mesh: &core_types.ResourceRef{
						Name:      "mesh-name-1",
						Namespace: "mesh-namespace-1",
					},
				},
			}
			mesh := &discovery_v1alpha1.Mesh{
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
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
				GetMesh(ctx, clients.ResourceRefToObjectKey(meshService.Spec.GetMesh())).
				Return(mesh, nil).
				Times(5)
			acpList := &networking_v1alpha1.AccessControlPolicyList{
				Items: []networking_v1alpha1.AccessControlPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-name-1"},
						Spec: networking_types.AccessControlPolicySpec{
							DestinationSelector: &core_types.ServiceSelector{
								ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
									Matcher: &core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-1"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-name-2"},
						Spec: networking_types.AccessControlPolicySpec{
							DestinationSelector: &core_types.ServiceSelector{
								ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
									Matcher: &core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-2"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-name-3"},
						Spec: networking_types.AccessControlPolicySpec{
							DestinationSelector: &core_types.ServiceSelector{
								ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
									Matcher: &core_types.ServiceSelector_Matcher{
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
			var newTranslatorError = func() *networking_types.AccessControlPolicyStatus_TranslatorError {
				return &networking_types.AccessControlPolicyStatus_TranslatorError{
					TranslatorId: "translator-id",
					ErrorMessage: "",
				}
			}
			var capturedACPsWithStatus []*networking_v1alpha1.AccessControlPolicy
			for _, acp := range acpList.Items {
				acp := acp
				resourceSelector.
					EXPECT().
					GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
					Return([]*discovery_v1alpha1.MeshService{meshService}, nil)
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
					DoAndReturn(func(ctx context.Context, acp *networking_v1alpha1.AccessControlPolicy) error {
						capturedACPsWithStatus = append(capturedACPsWithStatus, acp)
						return nil
					})
			}
			err := meshServiceHandler.OnCreate(meshService)
			Expect(err).ToNot(HaveOccurred())
			var expectedTranslatorErrors []*networking_types.AccessControlPolicyStatus_TranslatorError
			for range meshTranslators {
				expectedTranslatorErrors = append(expectedTranslatorErrors, newTranslatorError())
			}
			expectedStatus := networking_types.AccessControlPolicyStatus{
				TranslationStatus: &core_types.Status{
					State:   core_types.Status_PROCESSING_ERROR,
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
