package istio

import (
	"context"
	"fmt"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/discovery/istio"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/utils"
	"go.uber.org/zap"
)

type istioConfigDiscovery struct {
	el v1.IstioDiscoveryEventLoop
}

func NewIstioConfigDiscovery(ctx context.Context, cs *clientset.Clientset) (*istioConfigDiscovery, error) {
	istioClient, err := clientset.IstioFromContext(ctx)
	if err != nil {
		return nil, err
	}

	emitter := v1.NewIstioDiscoveryEmitter(
		cs.Discovery.Mesh,
		cs.Input.Install,
		istioClient.MeshPolicies,
	)
	syncer := newIstioConfigDiscoverSyncer(cs)
	el := v1.NewIstioDiscoveryEventLoop(emitter, syncer)

	return &istioConfigDiscovery{el: el}, nil
}

func (s *istioConfigDiscovery) Run(ctx context.Context) (<-chan error, error) {
	watchOpts := clients.WatchOpts{
		Ctx:         ctx,
		RefreshRate: time.Minute * 1,
	}
	return s.el.Run(nil, watchOpts)
}

func (s *istioConfigDiscovery) HandleError(ctx context.Context, err error) {
	if err != nil {
		contextutils.LoggerFrom(ctx).With(zap.Error(err)).Info("istio config discovery failure")
	}
}

type istioConfigDiscoverSyncer struct {
	cs *clientset.Clientset
}

func newIstioConfigDiscoverSyncer(cs *clientset.Clientset) *istioConfigDiscoverSyncer {
	return &istioConfigDiscoverSyncer{cs: cs}
}

func (s *istioConfigDiscoverSyncer) Sync(ctx context.Context, snap *v1.IstioDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-config-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	fields := []interface{}{
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("installs", len(snap.Installs.List())),
		zap.Int("mesh-policies", len(snap.Meshpolicies)),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	istioMeshes := utils.GetMeshes(snap.Meshes.List(), utils.IstioMeshFilterFunc)
	istioInstalls := utils.GetInstalls(snap.Installs.List(), utils.IstioInstallFilterFunc)

	fullMeshes := organizeMeshes(istioMeshes, istioInstalls, snap.Meshpolicies)

	var updatedMeshes v1.MeshList
	for _, fullMesh := range fullMeshes {
		updatedMeshes = append(updatedMeshes, fullMesh.merge())
	}

	meshReconciler := v1.NewMeshReconciler(s.cs.Discovery.Mesh)
	listOpts := clients.ListOpts{
		Ctx:      ctx,
		Selector: istio.DiscoverySelector,
	}
	return meshReconciler.Reconcile("", updatedMeshes, nil, listOpts)
}

func organizeMeshes(meshes v1.MeshList, installs v1.InstallList, meshPolicies v1alpha1.MeshPolicyList) FullMeshList {
	result := make(FullMeshList, len(meshes))
	for i, mesh := range meshes {
		istioMesh := mesh.GetIstio()
		if istioMesh == nil {
			continue
		}
		fullMesh := &FullMesh{
			Mesh: mesh,
		}
		for _, install := range installs {
			if install.InstallationNamespace == istioMesh.InstallationNamespace {
				fullMesh.Install = install
				break
			}
		}

		for _, policy := range meshPolicies {
			if policy.Metadata.Namespace == istioMesh.InstallationNamespace {
				fullMesh.MeshPolicy = policy
			}
		}
		result[i] = fullMesh
	}
	return result
}

type FullMeshList []*FullMesh
type FullMesh struct {
	Install    *v1.Install
	MeshPolicy *v1alpha1.MeshPolicy
	Mesh       *v1.Mesh
}

// Main merge method for discovered info
// Priority of data is as such Install > MeshPolicy > Mesh
func (fm *FullMesh) merge() *v1.Mesh {
	result := fm.Mesh
	istioMesh := fm.Mesh.GetIstio()
	if istioMesh == nil {
		return fm.Mesh
	}
	mtlsConfig := &v1.MtlsConfig{}
	if fm.MeshPolicy != nil {
		for _, peers := range fm.MeshPolicy.GetPeers() {
			mtls := peers.GetMtls()
			if mtls == nil {
				continue
			}

			// if mtls is strict and peer is optional mtls is enabled
			if mtls.Mode == v1alpha1.MutualTls_STRICT && !fm.MeshPolicy.GetPeerIsOptional() {
				mtlsConfig.MtlsEnabled = true
			}
		}
	}

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
