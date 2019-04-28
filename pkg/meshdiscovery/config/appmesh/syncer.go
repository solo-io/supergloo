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
	"github.com/solo-io/supergloo/pkg/translator/appmesh"
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

func newAppmeshDiscoveryConfigSyncer(cs *clientset.Clientset) *appmeshDiscoveryConfigSyncer {
	return &appmeshDiscoveryConfigSyncer{cs: cs}
}

type appmeshDiscoveryConfigSyncer struct {
	cs *clientset.Clientset
}

func (s *appmeshDiscoveryConfigSyncer) Sync(ctx context.Context, snap *v1.AppmeshDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("appmesh-config-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	pods, meshes, upstreams := snap.Pods.List(), snap.Meshes.List(), snap.Upstreams.List()
	fields := []interface{}{
		zap.Int("meshes", len(meshes)),
		zap.Int("pods", len(pods)),
		zap.Int("upstreams", len(upstreams)),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	var updatedMeshes v1.MeshList
	for _, mesh := range snap.Meshes.List() {
		config, err := appmesh.NewAwsAppMeshConfiguration(mesh.Metadata.Name, pods, upstreams)
		if err != nil {
			return err
		}

	}

	meshReconciler := v1.NewMeshReconciler(s.cs.Discovery.Mesh)
	listOpts := clients.ListOpts{
		Ctx:      ctx,
		Selector: istio.DiscoverySelector,
	}
	return meshReconciler.Reconcile("", updatedMeshes, nil, listOpts)
}

func updatedMesh(config *appmesh.AwsAppMeshConfiguration) *v1.Mesh {
	return nil
}
