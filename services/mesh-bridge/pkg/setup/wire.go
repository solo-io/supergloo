// +build wireinject

package setup

import (
	"context"

	"github.com/google/wire"
	"github.com/solo-io/go-utils/configutils"
	"github.com/solo-io/go-utils/envutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/internal/kube"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/setup/config"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/syncer"
	"github.com/solo-io/mesh-projects/services/mesh-bridge/pkg/translator"
)

func MustInitializeMeshBridge(ctx context.Context) (LoopSet, error) {
	wire.Build(
		// kube related configs.
		envutils.MustGetPodNamespace,
		kube.MustGetKubeConfig,
		kube.MustGetClient,
		configutils.NewConfigMapClient,

		// settings
		config.GetInitialSettings,
		config.GetOperatorConfig,

		// clientsets
		config.MustGetMeshBridgeClient,
		config.MustGetClientSet,

		// Needed by operator (installer) loop
		config.GetWatchOpts,
		config.GetWatchNamespaces,
		translator.NewMeshBridgeTranslator,

		v1.NewNetworkBridgeEmitter,
		NewNetworkBridgeSnapshotEmitter,
		syncer.NewMeshBridgeSyncer,
		v1.NewNetworkBridgeEventLoop,

		NewLoopSet,
	)

	return LoopSet{}, nil
}
