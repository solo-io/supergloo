package traffic_policy_translator_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	mock_traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/mocks"
	mock_processor "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/smh/discovery"
	mock_smh_networking "github.com/solo-io/service-mesh-hub/test/mocks/smh/networking"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Translator", func() {
	var (
		ctx                           context.Context
		ctrl                          *gomock.Controller
		mockPreprocessor              *mock_processor.MockTrafficPolicyPreprocessor
		mockIstioTranslator           *mock_traffic_policy_translator.MockTrafficPolicyMeshTranslator
		mockMeshServiceClient         *mock_core.MockMeshServiceClient
		mockMeshClient                *mock_core.MockMeshClient
		mockTrafficPolicyClient       *mock_smh_networking_clients.MockTrafficPolicyClient
		mockTrafficPolicyEventWatcher *mock_smh_networking.MockTrafficPolicyEventWatcher
		mockMeshServiceEventWatcher   *mock_smh_discovery.MockMeshServiceEventWatcher
		trafficPolicyEventHandler     *smh_networking_controller.TrafficPolicyEventHandlerFuncs
		meshServiceEventHandler       *smh_discovery_controller.MeshServiceEventHandlerFuncs
		translator                    traffic_policy_translator.TrafficPolicyTranslatorLoop
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockTrafficPolicyClient = mock_smh_networking_clients.NewMockTrafficPolicyClient(ctrl)
		mockPreprocessor = mock_processor.NewMockTrafficPolicyPreprocessor(ctrl)
		mockIstioTranslator = mock_traffic_policy_translator.NewMockTrafficPolicyMeshTranslator(ctrl)
		mockTrafficPolicyEventWatcher = mock_smh_networking.NewMockTrafficPolicyEventWatcher(ctrl)
		mockMeshServiceEventWatcher = mock_smh_discovery.NewMockMeshServiceEventWatcher(ctrl)
		translator = traffic_policy_translator.NewTrafficPolicyTranslatorLoop(
			mockPreprocessor,
			[]traffic_policy_translator.TrafficPolicyMeshTranslator{
				mockIstioTranslator,
			},
			mockMeshClient,
			mockMeshServiceClient,
			mockTrafficPolicyClient,
			mockTrafficPolicyEventWatcher,
			mockMeshServiceEventWatcher,
		)
		mockTrafficPolicyEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_networking_controller.TrafficPolicyEventHandlerFuncs) error {
				trafficPolicyEventHandler = eventHandler
				return nil
			})
		mockMeshServiceEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *smh_discovery_controller.MeshServiceEventHandlerFuncs) error {
				meshServiceEventHandler = eventHandler
				return nil
			})
		translator.Start(ctx)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("should handle TrafficPolicy events", func() {
		var (
			triggeringTP *smh_networking.TrafficPolicy
		)

		BeforeEach(func() {
			triggeringTP = &smh_networking.TrafficPolicy{
				Status: smh_networking_types.TrafficPolicyStatus{
					TranslationStatus: &smh_core_types.Status{
						State:   smh_core_types.Status_ACCEPTED,
						Message: "",
					},
				},
			}
		})

		It("handle create for TrafficPolicy", func() {
			meshServiceMCKey := selection.MeshServiceId{
				Name:      "name",
				Namespace: "namespace",
			}
			mergedTPsByMeshService := map[selection.MeshServiceId][]*smh_networking.TrafficPolicy{
				meshServiceMCKey: {},
			}
			meshServiceObjectKey := client.ObjectKey{Name: meshServiceMCKey.Name, Namespace: meshServiceMCKey.Namespace}
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshService := &smh_discovery.MeshService{
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: &smh_core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
				},
			}
			mesh := &smh_discovery.Mesh{
				Spec: smh_discovery_types.MeshSpec{
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
				},
			}
			mockMeshServiceClient.
				EXPECT().
				GetMeshService(ctx, meshServiceObjectKey).
				Return(meshService, nil)
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, meshObjKey).
				Return(mesh, nil)
			mockPreprocessor.
				EXPECT().
				PreprocessTrafficPolicy(ctx, triggeringTP).
				Return(mergedTPsByMeshService, nil)
			mockIstioTranslator.
				EXPECT().
				TranslateTrafficPolicy(contextutils.WithLogger(ctx, ""), meshService, mesh, mergedTPsByMeshService[meshServiceMCKey]).
				Return(nil)
			mockIstioTranslator.
				EXPECT().
				Name().
				Return("")
			mockTrafficPolicyClient.EXPECT().UpdateTrafficPolicyStatus(ctx, triggeringTP).Return(nil)
			trafficPolicyEventHandler.OnCreate(triggeringTP)
		})

		It("should return translator specific error statuses", func() {
			meshServiceMCKey := selection.MeshServiceId{
				Name:      "name",
				Namespace: "namespace",
			}
			mergedTPsByMeshService := map[selection.MeshServiceId][]*smh_networking.TrafficPolicy{
				meshServiceMCKey: {},
			}
			meshServiceObjectKey := client.ObjectKey{Name: meshServiceMCKey.Name, Namespace: meshServiceMCKey.Namespace}
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshService := &smh_discovery.MeshService{
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: &smh_core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
				},
			}
			mesh := &smh_discovery.Mesh{
				Spec: smh_discovery_types.MeshSpec{
					MeshType: &smh_discovery_types.MeshSpec_Istio1_6_{},
				},
			}
			mockMeshServiceClient.
				EXPECT().
				GetMeshService(ctx, meshServiceObjectKey).
				Return(meshService, nil)
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, meshObjKey).
				Return(mesh, nil)
			mockPreprocessor.
				EXPECT().
				PreprocessTrafficPolicy(ctx, triggeringTP).
				Return(mergedTPsByMeshService, nil)
			translatorError := &smh_networking_types.TrafficPolicyStatus_TranslatorError{
				TranslatorId: istio_translator.TranslatorId,
				ErrorMessage: "error message",
			}
			mockIstioTranslator.
				EXPECT().
				TranslateTrafficPolicy(contextutils.WithLogger(ctx, ""), meshService, mesh, mergedTPsByMeshService[meshServiceMCKey]).
				Return(translatorError)
			mockIstioTranslator.
				EXPECT().
				Name().
				Return("")
			expectedMeshTypeStatuses := []*smh_networking_types.TrafficPolicyStatus_TranslatorError{translatorError}

			expectedTP := &smh_networking.TrafficPolicy{}
			expectedTP.Status.TranslationStatus = &smh_core_types.Status{
				State:   smh_core_types.Status_PROCESSING_ERROR,
				Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
			}
			expectedTP.Status.TranslatorErrors = expectedMeshTypeStatuses
			mockTrafficPolicyClient.EXPECT().UpdateTrafficPolicyStatus(ctx, expectedTP).Return(nil)

			trafficPolicyEventHandler.OnCreate(triggeringTP)
		})
	})

	Describe("should handle MeshService events", func() {
		var (
			triggerMeshService *smh_discovery.MeshService
		)

		BeforeEach(func() {
			triggerMeshService = &smh_discovery.MeshService{
				Status: smh_discovery_types.MeshServiceStatus{
					FederationStatus: &smh_core_types.Status{
						State:   smh_core_types.Status_ACCEPTED,
						Message: "",
					},
				},
			}
		})

		It("should upsert policy resources for MeshService", func() {
			meshServiceMCKey := selection.MeshServiceId{
				Name:      "name",
				Namespace: "namespace",
			}
			meshServiceObjectKey := client.ObjectKey{Name: meshServiceMCKey.Name, Namespace: meshServiceMCKey.Namespace}
			mergedTPsByMeshService := map[selection.MeshServiceId][]*smh_networking.TrafficPolicy{meshServiceMCKey: {}}
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshService := &smh_discovery.MeshService{
				Spec: smh_discovery_types.MeshServiceSpec{
					Mesh: &smh_core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
				},
			}
			mesh := &smh_discovery.Mesh{
				Spec: smh_discovery_types.MeshSpec{
					MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
				},
			}
			mockMeshServiceClient.
				EXPECT().
				GetMeshService(ctx, meshServiceObjectKey).
				Return(meshService, nil)
			mockMeshClient.
				EXPECT().
				GetMesh(ctx, meshObjKey).
				Return(mesh, nil)
			mockPreprocessor.
				EXPECT().
				PreprocessTrafficPoliciesForMeshService(ctx, triggerMeshService).
				Return(mergedTPsByMeshService, nil)
			mockIstioTranslator.
				EXPECT().
				TranslateTrafficPolicy(contextutils.WithLogger(ctx, ""), meshService, mesh, mergedTPsByMeshService[meshServiceMCKey]).
				Return(nil)
			mockIstioTranslator.
				EXPECT().
				Name().
				Return("")

			meshServiceEventHandler.OnCreate(triggerMeshService)
		})

	})
})
