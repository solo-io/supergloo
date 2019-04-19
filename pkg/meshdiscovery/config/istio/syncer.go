package istio

import (
	"context"
	"fmt"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/eventloop"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/discovery/istio"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/utils"
	selectorutils "github.com/solo-io/supergloo/pkg/translator/utils"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	injectionLabel = "istio-injection"
	enabled        = "enabled"
)

var (
	injectedSelector = map[string]string{
		injectionLabel: enabled,
	}
)

func NewIstioConfigDiscoveryRunner(ctx context.Context, cs *clientset.Clientset) (eventloop.EventLoop, error) {
	istioClient, err := clientset.IstioClientsetFromContext(ctx)
	if err != nil {
		return nil, err
	}

	emitter := v1.NewIstioDiscoveryEmitter(
		cs.Discovery.Mesh,
		cs.Input.Install,
		cs.Input.Namespace,
		cs.Input.Pod,
		cs.Input.Upstream,
		istioClient.MeshPolicies,
	)
	syncer := newIstioConfigDiscoverSyncer(cs)
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
		zap.Int("meshes", len(snap.Meshes.List())),
		zap.Int("installs", len(snap.Installs.List())),
		zap.Int("mesh-policies", len(snap.Meshpolicies)),
		zap.Int("namespaces", len(snap.Kubenamespaces.List())),
		zap.Int("pods", len(snap.Pods.List())),
		zap.Int("upstreams", len(snap.Upstreams.List())),
	}

	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	istioMeshes := utils.GetMeshes(snap.Meshes.List(), utils.IstioMeshFilterFunc)
	istioInstalls := utils.GetInstalls(snap.Installs.List(), utils.IstioInstallFilterFunc)
	injectedNamespaces := utils.FilterNamespaces(snap.Kubenamespaces.List(), func(namespace *v1.KubeNamespace) bool {
		return labels.SelectorFromSet(injectedSelector).Matches(labels.Set(namespace.Labels))
	})
	injectedPodsByNamespace := utils.GetInjectedPods(injectedNamespaces, snap.Pods.List(),
		func(pod *v1.Pod) bool {
			return true
		},
	)

	meshResources := organizeMeshes(
		istioMeshes,
		istioInstalls,
		snap.Meshpolicies,
		injectedPodsByNamespace,
		snap.Upstreams.List(),
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
	injectedPods v1.PodsByNamespace, upstreams gloov1.UpstreamList) meshResourceList {
	result := make(meshResourceList, len(meshes))

	selector := utils.InjectionNamespaceSelector(injectedPods)

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
		if upstreams, err := selectorutils.UpstreamsForSelector(selector, upstreams); err == nil {
			fullMesh.Upstreams = upstreams
		}

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
	result.DiscoveryMetadata.MeshUpstreams = meshUpstreams

	result.DiscoveryMetadata.InjectedNamespaceLabel = injectionLabel
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
			result.DiscoveryMetadata.EnableAutoInject = istioMeshInstall.GetEnableAutoInject()
			mtlsConfig.MtlsEnabled = istioMeshInstall.GetEnableMtls()
			mtlsConfig.RootCertificate = istioMeshInstall.CustomRootCert
		}
		result.DiscoveryMetadata.InstallationNamespace = fm.Install.InstallationNamespace
	}
	result.DiscoveryMetadata.MtlsConfig = mtlsConfig
	return result
}
