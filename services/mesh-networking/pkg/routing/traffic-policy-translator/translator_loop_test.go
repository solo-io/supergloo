package traffic_policy_translator_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	istio_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/istio-translator"
	mock_traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/mocks"
	mock_processor "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/networking"
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
		mockTrafficPolicyClient       *mock_zephyr_networking_clients.MockTrafficPolicyClient
		mockTrafficPolicyEventWatcher *mock_zephyr_networking.MockTrafficPolicyEventWatcher
		mockMeshServiceEventWatcher   *mock_zephyr_discovery.MockMeshServiceEventWatcher
		trafficPolicyEventHandler     *zephyr_networking_controller.TrafficPolicyEventHandlerFuncs
		meshServiceEventHandler       *zephyr_discovery_controller.MeshServiceEventHandlerFuncs
		translator                    traffic_policy_translator.TrafficPolicyTranslatorLoop
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		mockTrafficPolicyClient = mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		mockPreprocessor = mock_processor.NewMockTrafficPolicyPreprocessor(ctrl)
		mockIstioTranslator = mock_traffic_policy_translator.NewMockTrafficPolicyMeshTranslator(ctrl)
		mockTrafficPolicyEventWatcher = mock_zephyr_networking.NewMockTrafficPolicyEventWatcher(ctrl)
		mockMeshServiceEventWatcher = mock_zephyr_discovery.NewMockMeshServiceEventWatcher(ctrl)
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
			DoAndReturn(func(ctx context.Context, eventHandler *zephyr_networking_controller.TrafficPolicyEventHandlerFuncs) error {
				trafficPolicyEventHandler = eventHandler
				return nil
			})
		mockMeshServiceEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandler *zephyr_discovery_controller.MeshServiceEventHandlerFuncs) error {
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
			triggeringTP *zephyr_networking.TrafficPolicy
		)

		BeforeEach(func() {
			triggeringTP = &zephyr_networking.TrafficPolicy{
				Status: zephyr_networking_types.TrafficPolicyStatus{
					TranslationStatus: &zephyr_core_types.Status{
						State:   zephyr_core_types.Status_ACCEPTED,
						Message: "",
					},
				},
			}
		})

		It("handle create for TrafficPolicy", func() {
			meshServiceMCKey := selector.MeshServiceId{
				Name:      "name",
				Namespace: "namespace",
			}
			mergedTPsByMeshService := map[selector.MeshServiceId][]*zephyr_networking.TrafficPolicy{
				meshServiceMCKey: {},
			}
			meshServiceObjectKey := client.ObjectKey{Name: meshServiceMCKey.Name, Namespace: meshServiceMCKey.Namespace}
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
				},
			}
			mesh := &zephyr_discovery.Mesh{
				Spec: zephyr_discovery_types.MeshSpec{
					MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
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
			meshServiceMCKey := selector.MeshServiceId{
				Name:      "name",
				Namespace: "namespace",
			}
			mergedTPsByMeshService := map[selector.MeshServiceId][]*zephyr_networking.TrafficPolicy{
				meshServiceMCKey: {},
			}
			meshServiceObjectKey := client.ObjectKey{Name: meshServiceMCKey.Name, Namespace: meshServiceMCKey.Namespace}
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
				},
			}
			mesh := &zephyr_discovery.Mesh{
				Spec: zephyr_discovery_types.MeshSpec{
					MeshType: &zephyr_discovery_types.MeshSpec_Istio1_6_{},
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
			translatorError := &zephyr_networking_types.TrafficPolicyStatus_TranslatorError{
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
			expectedMeshTypeStatuses := []*zephyr_networking_types.TrafficPolicyStatus_TranslatorError{translatorError}

			expectedTP := &zephyr_networking.TrafficPolicy{}
			expectedTP.Status.TranslationStatus = &zephyr_core_types.Status{
				State:   zephyr_core_types.Status_PROCESSING_ERROR,
				Message: fmt.Sprintf("Error while translating TrafficPolicy, check Status.TranslatorErrors for details"),
			}
			expectedTP.Status.TranslatorErrors = expectedMeshTypeStatuses
			mockTrafficPolicyClient.EXPECT().UpdateTrafficPolicyStatus(ctx, expectedTP).Return(nil)

			trafficPolicyEventHandler.OnCreate(triggeringTP)
		})
	})

	Describe("should handle MeshService events", func() {
		var (
			triggerMeshService *zephyr_discovery.MeshService
		)

		BeforeEach(func() {
			triggerMeshService = &zephyr_discovery.MeshService{
				Status: zephyr_discovery_types.MeshServiceStatus{
					FederationStatus: &zephyr_core_types.Status{
						State:   zephyr_core_types.Status_ACCEPTED,
						Message: "",
					},
				},
			}
		})

		It("should upsert policy resources for MeshService", func() {
			meshServiceMCKey := selector.MeshServiceId{
				Name:      "name",
				Namespace: "namespace",
			}
			meshServiceObjectKey := client.ObjectKey{Name: meshServiceMCKey.Name, Namespace: meshServiceMCKey.Namespace}
			mergedTPsByMeshService := map[selector.MeshServiceId][]*zephyr_networking.TrafficPolicy{meshServiceMCKey: {}}
			meshObjKey := client.ObjectKey{Name: "mesh-name", Namespace: "mesh-namespace"}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      meshObjKey.Name,
						Namespace: meshObjKey.Namespace,
					},
				},
			}
			mesh := &zephyr_discovery.Mesh{
				Spec: zephyr_discovery_types.MeshSpec{
					MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{},
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
