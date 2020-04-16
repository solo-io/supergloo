package acp_translator_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	access_control_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator"
	mock_access_control_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	mock_zephyr_networking2 "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/networking"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		acpEventWatcher           *mock_zephyr_networking2.MockAccessControlPolicyEventWatcher
		MeshServiceEventWatcher   *mock_zephyr_discovery.MockMeshServiceEventWatcher
		meshClient                *mock_core.MockMeshClient
		accessControlPolicyClient *mock_zephyr_networking.MockAccessControlPolicyClient
		resourceSelector          *mock_selector.MockResourceSelector
		mockMeshTranslator1       *mock_access_control_policy_translator.MockAcpMeshTranslator
		mockMeshTranslator2       *mock_access_control_policy_translator.MockAcpMeshTranslator
		meshTranslators           []*mock_access_control_policy_translator.MockAcpMeshTranslator

		// captured event handlers
		acpHandler         *zephyr_networking_controller.AccessControlPolicyEventHandlerFuncs
		meshServiceHandler *zephyr_discovery_controller.MeshServiceEventHandlerFuncs
		acpTranslator      access_control_policy_translator.AcpTranslatorLoop
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		acpEventWatcher = mock_zephyr_networking2.NewMockAccessControlPolicyEventWatcher(ctrl)
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
			DoAndReturn(func(ctx context.Context, eventHandler *zephyr_networking_controller.AccessControlPolicyEventHandlerFuncs) error {
				acpHandler = eventHandler
				return nil
			})
		MeshServiceEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *zephyr_discovery_controller.MeshServiceEventHandlerFuncs) error {

				meshServiceHandler = eventHandler
				return nil
			})
		acpTranslator.Start(ctx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var acp = func() *zephyr_networking.AccessControlPolicy {
		return &zephyr_networking.AccessControlPolicy{
			Spec: zephyr_networking_types.AccessControlPolicySpec{
				DestinationSelector: &zephyr_core_types.ServiceSelector{},
			},
		}
	}

	Describe("It should handle AccessControlPolicy events", func() {
		It("should handle AccessControlPolicy create", func() {
			acp := acp()
			matchingMeshServices := []*zephyr_discovery.MeshService{
				{
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: &zephyr_core_types.ResourceRef{
							Name:      "mesh-name-1",
							Namespace: "mesh-namespace-1",
						},
					},
				},
				{
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: &zephyr_core_types.ResourceRef{
							Name:      "mesh-name-2",
							Namespace: "mesh-namespace-2",
						},
					},
				},
			}
			meshesForService := []*zephyr_discovery.Mesh{
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-1", Namespace: "mesh-namespace-1"}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-22", Namespace: "mesh-namespace-22"}},
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
			var capturedACPWithStatus *zephyr_networking.AccessControlPolicy
			accessControlPolicyClient.
				EXPECT().
				UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, acp *zephyr_networking.AccessControlPolicy) error {
					capturedACPWithStatus = acp
					return nil
				})
			expectedStatus := zephyr_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State: zephyr_core_types.Status_ACCEPTED,
				},
				TranslatorErrors: nil,
			}
			err := acpHandler.OnCreate(acp)
			Expect(err).ToNot(HaveOccurred())
			Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
		})

		It("should aggregate list of translator errors", func() {
			acp := acp()
			matchingMeshServices := []*zephyr_discovery.MeshService{
				{
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: &zephyr_core_types.ResourceRef{
							Name:      "mesh-name-1",
							Namespace: "mesh-namespace-1",
						},
					},
				},
				{
					Spec: zephyr_discovery_types.MeshServiceSpec{
						Mesh: &zephyr_core_types.ResourceRef{
							Name:      "mesh-name-2",
							Namespace: "mesh-namespace-2",
						},
					},
				},
			}
			meshesForService := []*zephyr_discovery.Mesh{
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-1", Namespace: "mesh-namespace-1"}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "mesh-name-22", Namespace: "mesh-namespace-22"}},
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
			var translatorErrors []*zephyr_networking_types.AccessControlPolicyStatus_TranslatorError
			for _, meshTranslator := range meshTranslators {
				translatorErr := &zephyr_networking_types.AccessControlPolicyStatus_TranslatorError{
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
			var capturedACPWithStatus *zephyr_networking.AccessControlPolicy
			accessControlPolicyClient.
				EXPECT().
				UpdateAccessControlPolicyStatus(ctx, gomock.Any()).
				DoAndReturn(func(ctx context.Context, acp *zephyr_networking.AccessControlPolicy) error {
					capturedACPWithStatus = acp
					return nil
				})
			expectedStatus := zephyr_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
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
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "mesh-name-1",
						Namespace: "mesh-namespace-1",
					},
				},
			}
			mesh := &zephyr_discovery.Mesh{
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
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
			acpList := &zephyr_networking.AccessControlPolicyList{
				Items: []zephyr_networking.AccessControlPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-1"},
						Spec: zephyr_networking_types.AccessControlPolicySpec{
							DestinationSelector: &zephyr_core_types.ServiceSelector{
								ServiceSelectorType: &zephyr_core_types.ServiceSelector_Matcher_{
									Matcher: &zephyr_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-1"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-2"},
						Spec: zephyr_networking_types.AccessControlPolicySpec{
							DestinationSelector: &zephyr_core_types.ServiceSelector{
								ServiceSelectorType: &zephyr_core_types.ServiceSelector_Matcher_{
									Matcher: &zephyr_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-2"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-3"},
						Spec: zephyr_networking_types.AccessControlPolicySpec{
							DestinationSelector: &zephyr_core_types.ServiceSelector{
								ServiceSelectorType: &zephyr_core_types.ServiceSelector_Matcher_{
									Matcher: &zephyr_core_types.ServiceSelector_Matcher{
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
			var capturedACPsWithStatus []*zephyr_networking.AccessControlPolicy
			for _, acp := range acpList.Items {
				acp := acp
				resourceSelector.
					EXPECT().
					GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
					Return([]*zephyr_discovery.MeshService{meshService}, nil)
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
					DoAndReturn(func(ctx context.Context, acp *zephyr_networking.AccessControlPolicy) error {
						capturedACPsWithStatus = append(capturedACPsWithStatus, acp)
						return nil
					})
			}
			err := meshServiceHandler.OnCreate(meshService)
			Expect(err).ToNot(HaveOccurred())
			expectedStatus := zephyr_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State: zephyr_core_types.Status_ACCEPTED,
				},
				TranslatorErrors: nil,
			}
			for _, capturedACPWithStatus := range capturedACPsWithStatus {
				Expect(capturedACPWithStatus.Status).To(Equal(expectedStatus))
			}
		})

		It("should aggregate translator errors for each applicable ACP for MeshService", func() {
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "mesh-name-1",
						Namespace: "mesh-namespace-1",
					},
				},
			}
			mesh := &zephyr_discovery.Mesh{
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
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
			acpList := &zephyr_networking.AccessControlPolicyList{
				Items: []zephyr_networking.AccessControlPolicy{
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-1"},
						Spec: zephyr_networking_types.AccessControlPolicySpec{
							DestinationSelector: &zephyr_core_types.ServiceSelector{
								ServiceSelectorType: &zephyr_core_types.ServiceSelector_Matcher_{
									Matcher: &zephyr_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-1"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-2"},
						Spec: zephyr_networking_types.AccessControlPolicySpec{
							DestinationSelector: &zephyr_core_types.ServiceSelector{
								ServiceSelectorType: &zephyr_core_types.ServiceSelector_Matcher_{
									Matcher: &zephyr_core_types.ServiceSelector_Matcher{
										Namespaces: []string{"dest-namespace-2"},
									},
								},
							},
						},
					},
					{
						ObjectMeta: k8s_meta_types.ObjectMeta{Name: "acp-name-3"},
						Spec: zephyr_networking_types.AccessControlPolicySpec{
							DestinationSelector: &zephyr_core_types.ServiceSelector{
								ServiceSelectorType: &zephyr_core_types.ServiceSelector_Matcher_{
									Matcher: &zephyr_core_types.ServiceSelector_Matcher{
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
			var newTranslatorError = func() *zephyr_networking_types.AccessControlPolicyStatus_TranslatorError {
				return &zephyr_networking_types.AccessControlPolicyStatus_TranslatorError{
					TranslatorId: "translator-id",
					ErrorMessage: "",
				}
			}
			var capturedACPsWithStatus []*zephyr_networking.AccessControlPolicy
			for _, acp := range acpList.Items {
				acp := acp
				resourceSelector.
					EXPECT().
					GetMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector()).
					Return([]*zephyr_discovery.MeshService{meshService}, nil)
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
					DoAndReturn(func(ctx context.Context, acp *zephyr_networking.AccessControlPolicy) error {
						capturedACPsWithStatus = append(capturedACPsWithStatus, acp)
						return nil
					})
			}
			err := meshServiceHandler.OnCreate(meshService)
			Expect(err).ToNot(HaveOccurred())
			var expectedTranslatorErrors []*zephyr_networking_types.AccessControlPolicyStatus_TranslatorError
			for range meshTranslators {
				expectedTranslatorErrors = append(expectedTranslatorErrors, newTranslatorError())
			}
			expectedStatus := zephyr_networking_types.AccessControlPolicyStatus{
				TranslationStatus: &zephyr_core_types.Status{
					State:   zephyr_core_types.Status_PROCESSING_ERROR,
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
