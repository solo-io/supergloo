package linkerd

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/discovery/linkerd"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/utils"
	"go.uber.org/zap"
)

const (
	injectionAnnotation = "linkerd.io/inject"
)

func NewLinkerdConfigDiscoveryRunner(ctx context.Context, cs *clientset.Clientset) (eventloop.EventLoop, error) {
	emitter := v1.NewLinkerdDiscoveryEmitter(
		cs.Discovery.Mesh,
		cs.Input.Install,
	)
	syncer := newLinkerdConfigDiscoverSyncer(cs)
	el := v1.NewLinkerdDiscoveryEventLoop(emitter, syncer)

	return el, nil
}

type linkerdConfigDiscoverSyncer struct {
	cs *clientset.Clientset
}

func newLinkerdConfigDiscoverSyncer(cs *clientset.Clientset) *linkerdConfigDiscoverSyncer {
	return &linkerdConfigDiscoverSyncer{cs: cs}
}

func (lcds *linkerdConfigDiscoverSyncer) Sync(ctx context.Context, snap *v1.LinkerdDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-config-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("installs", len(snap.Installs.List())),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	linkerdMeshes := utils.GetMeshes(snap.Meshes.List(), utils.LinkerdMeshFilterFunc)
	linkerdInstalls := utils.GetInstalls(snap.Installs.List(), utils.LinkerdInstallFilterFunc)

	meshResources := organizeMeshes(linkerdMeshes, linkerdInstalls)

	var updatedMeshes v1.MeshList
	for _, fullMesh := range meshResources {
		updatedMeshes = append(updatedMeshes, fullMesh.merge())
	}

	meshReconciler := v1.NewMeshReconciler(lcds.cs.Discovery.Mesh)
	listOpts := clients.ListOpts{
		Ctx:      ctx,
		Selector: linkerd.DiscoverySelector,
	}

	return meshReconciler.Reconcile("", updatedMeshes, nil, listOpts)
}

func organizeMeshes(meshes v1.MeshList, installs v1.InstallList) meshResourceList {
	result := make(meshResourceList, len(meshes))
	for i, mesh := range meshes {
		istioMesh := mesh.GetIstio()
		if istioMesh == nil {
			continue
		}
		fullMesh := &meshResources{
			Mesh: mesh,
		}
		for _, install := range installs {
			if install.InstallationNamespace == istioMesh.InstallationNamespace {
				fullMesh.Install = install
				break
			}
		}

		result[i] = fullMesh
	}
	return result
}

type meshResourceList []*meshResources
type meshResources struct {
	Install *v1.Install
	Mesh    *v1.Mesh
}

// Main merge method for discovered info
// Priority of data is as such Install > Mesh
func (fm *meshResources) merge() *v1.Mesh {
	result := fm.Mesh
	linkerdMesh := fm.Mesh.GetLinkerdMesh()
	if linkerdMesh == nil {
		return fm.Mesh
	}
	mtlsConfig := &v1.MtlsConfig{}
	if result.DiscoveryMetadata == nil {
		result.DiscoveryMetadata = &v1.DiscoveryMetadata{}
	}
	result.DiscoveryMetadata.InjectedNamespaceLabel = injectionAnnotation

	if fm.Install != nil {
		mesh := fm.Install.GetMesh()
		if mesh != nil {
			istioMeshInstall := mesh.GetIstioMesh()
			result.DiscoveryMetadata.MeshVersion = istioMeshInstall.GetIstioVersion()
			mtlsConfig.MtlsEnabled = istioMeshInstall.GetEnableMtls()
			mtlsConfig.RootCertificate = istioMeshInstall.CustomRootCert
		}
		result.DiscoveryMetadata.InstallationNamespace = fm.Install.InstallationNamespace
	}
	result.DiscoveryMetadata.MtlsConfig = mtlsConfig
	return result
}
