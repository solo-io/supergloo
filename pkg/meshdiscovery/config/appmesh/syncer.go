package appmesh

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/istio"
	"github.com/solo-io/supergloo/pkg/registration"
	"go.uber.org/zap"
)

func StartAppmeshDiscoveryConfigLoop(ctx context.Context, cs *clientset.Clientset, pubSub *registration.PubSub) {
	configLoop := newAppmeshDiscoveryConfigLoop(cs)
	listener := registration.NewSubscriber(ctx, pubSub, configLoop)
	listener.Listen(ctx)
}

type appmeshDiscoveryConfigLoop struct {
	cs *clientset.Clientset
}

func newAppmeshDiscoveryConfigLoop(cs *clientset.Clientset) *appmeshDiscoveryConfigLoop {
	return &appmeshDiscoveryConfigLoop{cs: cs}
}

func (cl *appmeshDiscoveryConfigLoop) Enabled(enabled registration.EnabledConfigLoops) bool {
	return enabled.AppMesh
}

func (cl *appmeshDiscoveryConfigLoop) Start(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
	emitter := v1.NewAppmeshDiscoveryEmitter(
		cl.cs.Discovery.Mesh,
		cl.cs.Input.Pod,
		cl.cs.Input.Upstream,
	)
	syncer := newAppmeshDiscoveryConfigSyncer(cl.cs)
	el := v1.NewAppmeshDiscoveryEventLoop(emitter, syncer)

	return el, nil
}

func newAppmeshDiscoveryConfigSyncer(c *clientset.Clientset) *appmeshDiscoveryConfigSyncer {
	return &appmeshDiscoveryConfigSyncer{c: c}
}

type appmeshDiscoveryConfigSyncer struct {
	c *clientset.Clientset
}

func (*appmeshDiscoveryConfigSyncer) Sync(ctx context.Context, snap *v1.AppmeshDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("appmesh-config-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("pods", len(snap.Pods.List())),
		zap.Int("upstreams", len(snap.Upstreams.List())),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	meshReconciler := v1.NewMeshReconciler(s.cs.Discovery.Mesh)
	listOpts := clients.ListOpts{
		Ctx:      ctx,
		Selector: istio.DiscoverySelector,
	}
}
