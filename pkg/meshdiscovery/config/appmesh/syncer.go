package appmesh


import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/appmesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/utils"
	"github.com/solo-io/supergloo/pkg/registration"
	appmeshcfg "github.com/solo-io/supergloo/pkg/translator/appmesh"
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
	reconciler := v1.NewMeshReconciler(cl.cs.Discovery.Mesh)
	syncer := newAppmeshDiscoveryConfigSyncer(reconciler)
	el := v1.NewAppmeshDiscoveryEventLoop(emitter, syncer)

	return el, nil
}

func newAppmeshDiscoveryConfigSyncer(reconciler v1.MeshReconciler) *appmeshDiscoveryConfigSyncer {
	return &appmeshDiscoveryConfigSyncer{reconciler: reconciler}
}

type appmeshDiscoveryConfigSyncer struct {
	reconciler v1.MeshReconciler
}

func (s *appmeshDiscoveryConfigSyncer) Sync(ctx context.Context, snap *v1.AppmeshDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("appmesh-config-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	pods, meshes, upstreams := snap.Pods, snap.Meshes, snap.Upstreams
	fields := []interface{}{
		zap.Int("meshes", len(meshes)),
		zap.Int("pods", len(pods)),
		zap.Int("upstreams", len(upstreams)),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	appmeshMeshes := utils.GetMeshes(snap.Meshes, utils.AppmeshFilterFunc, utils.FilterByLabels(appmesh.DiscoverySelector))

	var updatedMeshes v1.MeshList
	for _, mesh := range appmeshMeshes {
		config, err := appmeshcfg.NewAwsAppMeshConfiguration(mesh.Metadata.Name, pods, upstreams)
		if err != nil {
			return err
		}
		if typedConfig, ok := config.(*appmeshcfg.AwsAppMeshConfigurationImpl); ok {
			updatedMesh, err := updateMesh(mesh, typedConfig)
			if err != nil {
				return err
			}
			updatedMeshes = append(updatedMeshes, updatedMesh)
		}
	}

	listOpts := clients.ListOpts{
		Ctx:      ctx,
		Selector: appmesh.DiscoverySelector,
	}
	return s.reconciler.Reconcile("", updatedMeshes, nil, listOpts)
}

func updateMesh(mesh *v1.Mesh, config *appmeshcfg.AwsAppMeshConfigurationImpl) (*v1.Mesh, error) {
	result := mesh
	awsMesh := mesh.GetAwsAppMesh()
	if awsMesh == nil {
		return nil, errors.Errorf("non aws app mesh CRD found in aws app mesh config discovery")
	}
	if result.DiscoveryMetadata == nil {
		result.DiscoveryMetadata = &v1.DiscoveryMetadata{}
	}

	var meshUpstreams []*core.ResourceRef
	for _, upstream := range config.UpstreamList {
		ref := upstream.Metadata.Ref()
		meshUpstreams = append(meshUpstreams, &ref)
	}
	result.DiscoveryMetadata.Upstreams = meshUpstreams
	result.DiscoveryMetadata.EnableAutoInject = awsMesh.EnableAutoInject

	return result, nil
}
