package setup

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/registration"
)

func StartSuperglooConfigLoop(ctx context.Context, cs *clientset.Clientset, pubSub *registration.PubSub) {
	sgConfigLoop := &superglooConfigLoop{cs: cs}
	sgListener := registration.NewSubscriber(ctx, pubSub, sgConfigLoop)
	sgListener.Listen(ctx)
}

type superglooConfigLoop struct {
	cs *clientset.Clientset
}

func (scl *superglooConfigLoop) Enabled(enabled registration.EnabledConfigLoops) bool {
	return enabled.IstioSmi || enabled.Linkerd || enabled.Istio || enabled.AppMesh
}

func (scl *superglooConfigLoop) Start(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
	var syncers v1.ConfigSyncers

	if enabled.Istio {
		if enabled.IstioSmi {
			istioSyncer, err := createSmiConfigSyncer(ctx, scl.cs)
			if err != nil {
				return nil, err
			}
			syncers = append(syncers, istioSyncer)
		} else {
			istioSyncer, err := createIstioConfigSyncer(ctx, scl.cs)
			if err != nil {
				return nil, err
			}
			syncers = append(syncers, istioSyncer)
		}
	}

	if enabled.Linkerd {
		linkerdSyncer, err := createLinkerdConfigSyncer(ctx, scl.cs)
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, linkerdSyncer)
	}

	if enabled.AppMesh {
		syncers = append(syncers, createAppmeshConfigSyncer(scl.cs))
	}

	ctx = contextutils.WithLogger(ctx, "config-event-loop")

	configEmitter := v1.NewConfigEmitter(
		scl.cs.Supergloo.Mesh,
		scl.cs.Supergloo.MeshIngress,
		scl.cs.Supergloo.MeshGroup,
		scl.cs.Supergloo.RoutingRule,
		scl.cs.Supergloo.SecurityRule,
		scl.cs.Supergloo.TlsSecret,
		scl.cs.Supergloo.Upstream,
		scl.cs.Discovery.Pod,
		scl.cs.Discovery.Service,
	)
	configEventLoop := v1.NewConfigEventLoop(configEmitter, syncers)

	return configEventLoop, nil
}
