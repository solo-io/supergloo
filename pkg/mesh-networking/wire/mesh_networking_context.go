package wire

import (
	"context"
	"time"

	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	smh_networking_controller "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/controller"
	mc_manager "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/resolver"
	networking_snapshot "github.com/solo-io/service-mesh-hub/pkg/common/snapshot"
	access_control_enforcer "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-enforcer"
	access_control_policy "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/access/access-control-policy-translator"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/decider"
	traffic_policy_translator "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/routing/traffic-policy-translator"
	cert_manager "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/security/cert-manager"
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
	MeshWorkloadEventWatcher          smh_discovery_controller.MeshWorkloadEventWatcher
	MeshServiceEventWatcher           smh_discovery_controller.MeshServiceEventWatcher
	VirtualMeshEventWatcher           smh_networking_controller.VirtualMeshEventWatcher
	SnapshotValidator                 networking_snapshot.MeshNetworkingSnapshotValidator
	VMCSRSnapshotListener             cert_manager.VMCSRSnapshotListener
	FederationDeciderSnapshotListener decider.FederationDeciderSnapshotListener
}

func MeshNetworkingSnapshotContextProvider(
	meshWorkloadEventWatcher smh_discovery_controller.MeshWorkloadEventWatcher,
	meshServiceEventWatcher smh_discovery_controller.MeshServiceEventWatcher,
	virtualMeshEventWatcher smh_networking_controller.VirtualMeshEventWatcher,
	snapshotValidator networking_snapshot.MeshNetworkingSnapshotValidator,
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
	listenerGenerator, err := networking_snapshot.NewMeshNetworkingSnapshotGenerator(
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
