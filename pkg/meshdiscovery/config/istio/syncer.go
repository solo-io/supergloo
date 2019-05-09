package istio

import (
	"context"
	"fmt"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/mesh/istio"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/utils"
	"github.com/solo-io/supergloo/pkg/registration"
	"go.uber.org/zap"
)

const (
	namespaceInjectionLabel = "istio-injection"
	enabled                 = "enabled"
	disabled                = "disabled"

	podInjectionAnnotation = "sidecar.istio.io/inject"
	podAnnotationEnabled   = "true"
	podAnnotationDisabled  = "false"

	proxyContainer = "istio-proxy"
)

var (
	injectedSelector = map[string]string{
		namespaceInjectionLabel: enabled,
	}
)

func StartIstioDiscoveryConfigLoop(ctx context.Context, cs *clientset.Clientset, pubSub *registration.PubSub) {
	configLoop := &istioDiscoveryConfigLoop{cs: cs}
	listener := registration.NewSubscriber(ctx, pubSub, configLoop)
	listener.Listen(ctx)
}

type istioDiscoveryConfigLoop struct {
	cs *clientset.Clientset
}

func (cl *istioDiscoveryConfigLoop) Enabled(enabled registration.EnabledConfigLoops) bool {
	return enabled.Istio
}

func (cl *istioDiscoveryConfigLoop) Start(ctx context.Context, enabled registration.EnabledConfigLoops) (eventloop.EventLoop, error) {
	istioClient, err := clientset.IstioClientsetFromContext(ctx)
	if err != nil {
		return nil, err
	}

	emitter := v1.NewIstioDiscoveryEmitter(
		cl.cs.Discovery.Mesh,
		cl.cs.Input.Install,
		cl.cs.Input.Pod,
		cl.cs.Input.Upstream,
		istioClient.MeshPolicies,
	)
	syncer := newIstioConfigDiscoverSyncer(cl.cs)
	el := v1.NewIstioDiscoveryEventLoop(emitter, syncer)

	return el, nil
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
		zap.Int("meshes", len(snap.Meshes)),
		zap.Int("installs", len(snap.Installs)),
		zap.Int("mesh-policies", len(snap.Meshpolicies)),
		zap.Int("pods", len(snap.Pods)),
		zap.Int("upstreams", len(snap.Upstreams)),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	istioMeshes := utils.GetMeshes(snap.Meshes, utils.IstioMeshFilterFunc, utils.FilterByLabels(istio.DiscoverySelector))
	istioInstalls := utils.GetActiveInstalls(snap.Installs, utils.IstioInstallFilterFunc)
	injectedPods := utils.InjectedPodsByNamespace(snap.Pods, proxyContainer)

	meshResources := organizeMeshes(
		istioMeshes,
		istioInstalls,
		snap.Meshpolicies,
		injectedPods,
		snap.Upstreams,
	)

	var updatedMeshes v1.MeshList
	for _, fullMesh := range meshResources {
		updatedMeshes = append(updatedMeshes, fullMesh.merge())
	}

	meshReconciler := v1.NewMeshReconciler(s.cs.Discovery.Mesh)
	listOpts := clients.ListOpts{
		Ctx:      ctx,
		Selector: istio.DiscoverySelector,
	}
	return meshReconciler.Reconcile("", updatedMeshes, nil, listOpts)
}

func organizeMeshes(meshes v1.MeshList, installs v1.InstallList, meshPolicies v1alpha1.MeshPolicyList,
	injectedPods kubernetes.PodList, upstreams gloov1.UpstreamList) meshResourceList {
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

		for _, policy := range meshPolicies {
			if policy.Metadata.Namespace == istioMesh.InstallationNamespace {
				fullMesh.MeshPolicy = policy
			}
		}
		// Currently injection is a constant so there's no way to distinguish between
		// multiple istio deployments in a single cluster
		fullMesh.Upstreams = utils.GetUpstreamsForInjectedPods(injectedPods, upstreams)

		result[i] = fullMesh
	}
	return result
}

type meshResourceList []*meshResources
type meshResources struct {
	Install    *v1.Install
	MeshPolicy *v1alpha1.MeshPolicy
	Mesh       *v1.Mesh
	Upstreams  gloov1.UpstreamList
}

// Main merge method for discovered info
// Priority of data is as such Install > MeshPolicy > Mesh
func (fm *meshResources) merge() *v1.Mesh {
	result := fm.Mesh
	istioMesh := fm.Mesh.GetIstio()
	if istioMesh == nil {
		return fm.Mesh
	}
	mtlsConfig := &v1.MtlsConfig{}
	if result.DiscoveryMetadata == nil {
		result.DiscoveryMetadata = &v1.DiscoveryMetadata{}
	}

	var meshUpstreams []*core.ResourceRef
	for _, upstream := range fm.Upstreams {
		ref := upstream.Metadata.Ref()
		meshUpstreams = append(meshUpstreams, &ref)
	}
	result.DiscoveryMetadata.Upstreams = meshUpstreams

	result.DiscoveryMetadata.InjectedNamespaceLabel = namespaceInjectionLabel
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
			istioMeshInstall := mesh.GetIstio()
			result.DiscoveryMetadata.MeshVersion = istioMeshInstall.GetVersion()
			result.DiscoveryMetadata.EnableAutoInject = istioMeshInstall.GetEnableAutoInject()
			mtlsConfig.MtlsEnabled = istioMeshInstall.GetEnableMtls()
			mtlsConfig.RootCertificate = istioMeshInstall.CustomRootCert
		}
		result.DiscoveryMetadata.InstallationNamespace = fm.Install.InstallationNamespace
	}
	result.DiscoveryMetadata.MtlsConfig = mtlsConfig
	return result
}
