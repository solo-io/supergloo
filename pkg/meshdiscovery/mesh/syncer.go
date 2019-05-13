package mesh

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/cliutils"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/utils"
	"go.uber.org/zap"

	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
)

type meshDiscoverySyncer struct {
	meshClient     v1.MeshClient
	plugins        MeshDiscoveryPlugins
	meshReconciler v1.MeshReconciler
}

// calling this function with nil is valid and expected outside of tests
func NewMeshDiscoverySyncer(meshClient v1.MeshClient, plugins ...MeshDiscovery) v1.DiscoverySyncer {
	meshReconciler := v1.NewMeshReconciler(meshClient)
	return &meshDiscoverySyncer{
		meshClient:     meshClient,
		plugins:        plugins,
		meshReconciler: meshReconciler,
	}
}

func (s *meshDiscoverySyncer) Sync(ctx context.Context, snap *v1.DiscoverySnapshot) error {
	multierr := &multierror.Error{}
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("mesh-discovery-syncer-%v", snap.Hash()))
	fields := []interface{}{
		zap.Int("installs", len(snap.Installs)),
		zap.Int("pods", len(snap.Pods)),
	}
	logger := contextutils.LoggerFrom(ctx)
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)

	filteredSnap := s.filterOutNamespacesWithInstalls(snap)

	var discoveredMeshes v1.MeshList
	for _, meshDiscoveryPlugin := range s.plugins {
		meshes, err := meshDiscoveryPlugin.DiscoverMeshes(ctx, filteredSnap)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			logger.Errorf(err.Error())
		}
		discoveredMeshes = append(discoveredMeshes, meshes...)
	}

	// reconcile all discovered meshes
	err := s.meshReconciler.Reconcile("", discoveredMeshes, func(original, desired *v1.Mesh) (b bool, e error) {
		return false, nil
	}, clients.ListOpts{
		Ctx:      ctx,
		Selector: map[string]string{utils.SelectorCreatedByPrefix: utils.SelectorCreatedByValue},
	})
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}

	return multierr.ErrorOrNil()
}

func (s *meshDiscoverySyncer) filterOutNamespacesWithInstalls(snap *v1.DiscoverySnapshot) *v1.DiscoverySnapshot {
	var namespaces []string
	for _, install := range snap.Installs {
		if !cliutils.Contains(namespaces, install.Metadata.Namespace) {
			namespaces = append(namespaces, install.Metadata.Namespace)
		}
	}

	var newPodList skkube.PodList
	snap.Pods.Each(func(pod *skkube.Pod) {
		if !cliutils.Contains(namespaces, pod.Namespace) {
			newPodList = append(newPodList, pod)
		}
	})

	var newConfigMapList skkube.ConfigMapList
	snap.Configmaps.Each(func(configMap *skkube.ConfigMap) {
		if !cliutils.Contains(namespaces, configMap.Namespace) {
			newConfigMapList = append(newConfigMapList, configMap)
		}
	})
	return &v1.DiscoverySnapshot{
		Pods:       newPodList,
		Configmaps: newConfigMapList,
		Installs:   snap.Installs,
	}

}
