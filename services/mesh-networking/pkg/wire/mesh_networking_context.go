package wire

import (
	"context"
	"time"

	"github.com/solo-io/mesh-projects/services/common/multicluster"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/controllers"
	access_control_poilcy "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/access/access-control-poilcy"
	controller_factories "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/controllers"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	traffic_policy_translator "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator"
	cert_manager "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/security/cert-manager"
)

// just used to package everything up for wire
type MeshNetworkingContext struct {
	MultiClusterDeps              multicluster.MultiClusterDependencies
	MeshNetworkingClusterHandler  mc_manager.AsyncManagerHandler
	TrafficPolicyTranslator       traffic_policy_translator.TrafficPolicyTranslator
	MeshNetworkingSnapshotContext *MeshNetworkingSnapshotContext
	AccessControlPolicyTranslator access_control_poilcy.AccessControlPolicyTranslator
}

func MeshNetworkingContextProvider(
	multiClusterDeps multicluster.MultiClusterDependencies,
	meshNetworkingClusterHandler mc_manager.AsyncManagerHandler,
	trafficPolicyTranslator traffic_policy_translator.TrafficPolicyTranslator,
	meshNetworkingSnapshotContext *MeshNetworkingSnapshotContext,
	accessControlPolicyTranslator access_control_poilcy.AccessControlPolicyTranslator,
) MeshNetworkingContext {
	return MeshNetworkingContext{
		MultiClusterDeps:              multiClusterDeps,
		MeshNetworkingClusterHandler:  meshNetworkingClusterHandler,
		TrafficPolicyTranslator:       trafficPolicyTranslator,
		MeshNetworkingSnapshotContext: meshNetworkingSnapshotContext,
		AccessControlPolicyTranslator: accessControlPolicyTranslator,
	}
}

type MeshNetworkingSnapshotContext struct {
	MeshWorkloadControllerFactory controllers.MeshWorkloadControllerFactory
	MeshServiceControllerFactory  controllers.MeshServiceControllerFactory
	VirtualMeshControllerFactory  controller_factories.VirtualMeshControllerFactory
	SnapshotValidator             snapshot.MeshNetworkingSnapshotValidator
	VMCSRSnapshotListener         cert_manager.VMCSRSnapshotListener
}

func MeshNetworkingSnapshotContextProvider(
	meshWorkloadControllerFactory controllers.MeshWorkloadControllerFactory,
	meshServiceControllerFactory controllers.MeshServiceControllerFactory,
	virtualMeshControllerFactory controller_factories.VirtualMeshControllerFactory,
	snapshotValidator snapshot.MeshNetworkingSnapshotValidator,
	vmcsrSnapshotListener cert_manager.VMCSRSnapshotListener,
) *MeshNetworkingSnapshotContext {
	return &MeshNetworkingSnapshotContext{
		MeshWorkloadControllerFactory: meshWorkloadControllerFactory,
		MeshServiceControllerFactory:  meshServiceControllerFactory,
		VirtualMeshControllerFactory:  virtualMeshControllerFactory,
		SnapshotValidator:             snapshotValidator,
		VMCSRSnapshotListener:         vmcsrSnapshotListener,
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
	mgCtrl, err := m.VirtualMeshControllerFactory(mgr, "virtual-mesh-controller")
	if err != nil {
		return err
	}
	listenerGenerator, err := snapshot.NewMeshNetworkingSnapshotGenerator(ctx, m.SnapshotValidator, msCtrl, mgCtrl, mwCtrl)
	if err != nil {
		return err
	}
	listenerGenerator.RegisterListener(m.VMCSRSnapshotListener)
	go func() { listenerGenerator.StartPushingSnapshots(ctx, time.Second) }()
	return nil
}
