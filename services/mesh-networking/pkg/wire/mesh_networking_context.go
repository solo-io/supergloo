package wire

import (
	"context"
	"time"

	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/compute-target/k8s"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-enforcer"
	access_control_policy "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/access/access-control-policy-translator"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/compute-target/snapshot"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/decider"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/resolver"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator"
	cert_manager "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/security/cert-manager"
)

// just used to package everything up for wire
type MeshNetworkingContext struct {
	MultiClusterDeps              mc_manager.MultiClusterDependencies
	MeshNetworkingClusterHandler  mc_manager.AsyncManagerHandler
	TrafficPolicyTranslator       traffic_policy_translator.TrafficPolicyTranslatorLoop
	MeshNetworkingSnapshotContext *MeshNetworkingSnapshotContext
	AccessControlPolicyTranslator access_control_policy.AcpTranslatorLoop
	GlobalAccessPolicyEnforcer    access_control_enforcer.AccessPolicyEnforcerLoop
	FederationResolver            resolver.FederationResolver
}

func MeshNetworkingContextProvider(
	multiClusterDeps mc_manager.MultiClusterDependencies,
	meshNetworkingClusterHandler mc_manager.AsyncManagerHandler,
	trafficPolicyTranslator traffic_policy_translator.TrafficPolicyTranslatorLoop,
	meshNetworkingSnapshotContext *MeshNetworkingSnapshotContext,
	accessControlPolicyTranslator access_control_policy.AcpTranslatorLoop,
	GlobalAccessPolicyEnforcer access_control_enforcer.AccessPolicyEnforcerLoop,
	federationResolver resolver.FederationResolver,
) MeshNetworkingContext {
	return MeshNetworkingContext{
		MultiClusterDeps:              multiClusterDeps,
		MeshNetworkingClusterHandler:  meshNetworkingClusterHandler,
		TrafficPolicyTranslator:       trafficPolicyTranslator,
		MeshNetworkingSnapshotContext: meshNetworkingSnapshotContext,
		AccessControlPolicyTranslator: accessControlPolicyTranslator,
		GlobalAccessPolicyEnforcer:    GlobalAccessPolicyEnforcer,
		FederationResolver:            federationResolver,
	}
}

type MeshNetworkingSnapshotContext struct {
	MeshWorkloadEventWatcher          zephyr_discovery_controller.MeshWorkloadEventWatcher
	MeshServiceEventWatcher           zephyr_discovery_controller.MeshServiceEventWatcher
	VirtualMeshEventWatcher           zephyr_networking_controller.VirtualMeshEventWatcher
	SnapshotValidator                 snapshot.MeshNetworkingSnapshotValidator
	VMCSRSnapshotListener             cert_manager.VMCSRSnapshotListener
	FederationDeciderSnapshotListener decider.FederationDeciderSnapshotListener
}

func MeshNetworkingSnapshotContextProvider(
	meshWorkloadEventWatcher zephyr_discovery_controller.MeshWorkloadEventWatcher,
	meshServiceEventWatcher zephyr_discovery_controller.MeshServiceEventWatcher,
	virtualMeshEventWatcher zephyr_networking_controller.VirtualMeshEventWatcher,
	snapshotValidator snapshot.MeshNetworkingSnapshotValidator,
	vmcsrSnapshotListener cert_manager.VMCSRSnapshotListener,
	federationDeciderSnapshotListener decider.FederationDeciderSnapshotListener,
) *MeshNetworkingSnapshotContext {
	return &MeshNetworkingSnapshotContext{
		MeshWorkloadEventWatcher:          meshWorkloadEventWatcher,
		MeshServiceEventWatcher:           meshServiceEventWatcher,
		VirtualMeshEventWatcher:           virtualMeshEventWatcher,
		SnapshotValidator:                 snapshotValidator,
		VMCSRSnapshotListener:             vmcsrSnapshotListener,
		FederationDeciderSnapshotListener: federationDeciderSnapshotListener,
	}
}

func (m *MeshNetworkingSnapshotContext) StartListening(ctx context.Context) error {
	listenerGenerator, err := snapshot.NewMeshNetworkingSnapshotGenerator(
		ctx,
		m.SnapshotValidator,
		m.MeshServiceEventWatcher,
		m.VirtualMeshEventWatcher,
		m.MeshWorkloadEventWatcher,
	)
	if err != nil {
		return err
	}
	listenerGenerator.RegisterListener(m.VMCSRSnapshotListener)
	listenerGenerator.RegisterListener(m.FederationDeciderSnapshotListener)
	go func() { listenerGenerator.StartPushingSnapshots(ctx, time.Second) }()
	return nil
}
