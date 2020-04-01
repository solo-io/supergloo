package wire

import (
	"context"
	"time"

	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/controller"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/controllers"
	access_control_enforcer "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-enforcer"
	access_control_policy "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-policy-translator"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/decider"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/federation/resolver"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	traffic_policy_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator"
	cert_manager "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-manager"
)

// just used to package everything up for wire
type MeshNetworkingContext struct {
	MultiClusterDeps              multicluster.MultiClusterDependencies
	MeshNetworkingClusterHandler  mc_manager.AsyncManagerHandler
	TrafficPolicyTranslator       traffic_policy_translator.TrafficPolicyTranslatorLoop
	MeshNetworkingSnapshotContext *MeshNetworkingSnapshotContext
	AccessControlPolicyTranslator access_control_policy.AcpTranslatorLoop
	GlobalAccessPolicyEnforcer    access_control_enforcer.AccessPolicyEnforcerLoop
	FederationResolver            resolver.FederationResolver
}

func MeshNetworkingContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
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
	MeshWorkloadControllerFactory     controllers.MeshWorkloadControllerFactory
	MeshServiceControllerFactory      controllers.MeshServiceControllerFactory
	VirtualMeshController             controller.VirtualMeshController
	SnapshotValidator                 snapshot.MeshNetworkingSnapshotValidator
	VMCSRSnapshotListener             cert_manager.VMCSRSnapshotListener
	FederationDeciderSnapshotListener decider.FederationDeciderSnapshotListener
}

func MeshNetworkingSnapshotContextProvider(
	meshWorkloadControllerFactory controllers.MeshWorkloadControllerFactory,
	meshServiceControllerFactory controllers.MeshServiceControllerFactory,
	virtualMeshController controller.VirtualMeshController,
	snapshotValidator snapshot.MeshNetworkingSnapshotValidator,
	vmcsrSnapshotListener cert_manager.VMCSRSnapshotListener,
	federationDeciderSnapshotListener decider.FederationDeciderSnapshotListener,
) *MeshNetworkingSnapshotContext {
	return &MeshNetworkingSnapshotContext{
		MeshWorkloadControllerFactory:     meshWorkloadControllerFactory,
		MeshServiceControllerFactory:      meshServiceControllerFactory,
		VirtualMeshController:             virtualMeshController,
		SnapshotValidator:                 snapshotValidator,
		VMCSRSnapshotListener:             vmcsrSnapshotListener,
		FederationDeciderSnapshotListener: federationDeciderSnapshotListener,
	}
}

func (m *MeshNetworkingSnapshotContext) StartListening(ctx context.Context, mgr mc_manager.AsyncManager) error {
	msCtrl, err := m.MeshServiceControllerFactory.Build(mgr, "mesh-service-controller")
	if err != nil {
		return err
	}
	mwCtrl, err := m.MeshWorkloadControllerFactory.Build(mgr, "mesh-workload-controller")
	if err != nil {
		return err
	}
	listenerGenerator, err := snapshot.NewMeshNetworkingSnapshotGenerator(ctx, m.SnapshotValidator, msCtrl, m.VirtualMeshController, mwCtrl)
	if err != nil {
		return err
	}
	listenerGenerator.RegisterListener(m.VMCSRSnapshotListener)
	listenerGenerator.RegisterListener(m.FederationDeciderSnapshotListener)
	go func() { listenerGenerator.StartPushingSnapshots(ctx, time.Second) }()
	return nil
}
