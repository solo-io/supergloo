package istio

import (
	"context"

	"github.com/solo-io/mesh-projects/services/internal/utils"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"go.uber.org/zap"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
)

// the subsets of the kube api that we need
type CrdGetter interface {
	Read(name string, opts clients.ReadOpts) (*kubernetes.CustomResourceDefinition, error)
}

type istioDiscoveryPlugin struct {
	meshPolicyClient v1alpha1.MeshPolicyClient
	crdGetter        CrdGetter
	jobClient        kubernetes.JobClient
	resourceDetector IstioResourceDetector
}

func NewIstioDiscoverySyncer(writeNamespace string, meshReconciler v1.MeshReconciler, meshPolicyClient v1alpha1.MeshPolicyClient, crdGetter CrdGetter, jobClient kubernetes.JobClient) v1.DiscoverySyncer {
	return common.NewDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		&istioDiscoveryPlugin{meshPolicyClient: meshPolicyClient, crdGetter: crdGetter, jobClient: jobClient, resourceDetector: NewIstioResourceDetector()},
	)
}

func (p *istioDiscoveryPlugin) MeshType() string {
	return "istio"
}

var discoveryLabels = map[string]string{
	"discovered_by": "istio-mesh-discovery",
}

func (p *istioDiscoveryPlugin) DiscoveryLabels() map[string]string {
	return discoveryLabels
}

func (p *istioDiscoveryPlugin) DesiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	pilots := p.resourceDetector.DetectPilotDeployments(ctx, snap.Deployments)
	contextutils.LoggerFrom(ctx).Infow("pilots", zap.Any("length", len(pilots)))

	if len(pilots) == 0 {
		return nil, nil
	}

	injectedPods := p.resourceDetector.DetectInjectedIstioPods(ctx, snap.Pods)

	byCluster := snapShotByCluster(snap)

	var istioMeshes v1.MeshList
	for _, pilot := range pilots {
		meshPolicyCrdRegistered, err := p.resourceDetector.DetectMeshPolicyCrd(p.crdGetter, pilot.Cluster)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorw("Failed to detect mesh policy CRD", zap.Error(err), zap.Any("cluster", pilot.Cluster))
			continue
		}
		if !meshPolicyCrdRegistered {
			continue
		}

		globalMtlsEnabled, mtlsPermissive := func() (bool, bool) {
			// https://istio.io/docs/tasks/security/authn-policy/#globally-enabling-istio-mutual-tls
			defaultMeshPolicy, err := p.meshPolicyClient.Read("default", clients.ReadOpts{Ctx: ctx, Cluster: pilot.Cluster})
			if err != nil {
				return false, false
			}
			for _, peer := range defaultMeshPolicy.GetPeers() {
				if peer.GetMtls() != nil {
					return true, peer.GetMtls().GetMode() == v1alpha1.MutualTls_PERMISSIVE
				}
			}
			return false, false
		}()

		// ensure that the post-install jobs have run for this pilot,
		// if not, we're not ready to detect it
		postInstallComplete, err := p.resourceDetector.DetectPostInstallJobComplete(p.jobClient, pilot.Namespace, pilot.Cluster)
		if err != nil {
			contextutils.LoggerFrom(ctx).Errorf("failed to detect if post-install jobs have finished: %v", err)
			continue
		}

		if !postInstallComplete {
			continue
		}

		var autoInjectionEnabled bool
		sidecarInjector, err := byCluster[pilot.Cluster].Deployments.Find(pilot.Namespace, "istio-sidecar-injector")
		if err == nil && (sidecarInjector.Spec.Replicas == nil || *sidecarInjector.Spec.Replicas > 0) {
			autoInjectionEnabled = true
		}

		var smiEnabled bool
		smiAdapter, err := byCluster[pilot.Cluster].Deployments.Find(pilot.Namespace, "smi-adapter-istio")
		if err == nil && (smiAdapter.Spec.Replicas == nil || *smiAdapter.Spec.Replicas > 0) {
			smiEnabled = true
		}

		// https://istio.io/docs/tasks/security/plugin-ca-cert/#plugging-in-the-existing-certificate-and-key
		var rootCa *core.ResourceRef
		customRootCa, err := byCluster[pilot.Cluster].Tlssecrets.Find(pilot.Namespace, "cacerts")
		if err == nil {
			root := customRootCa.Metadata.Ref()
			rootCa = &root
		}

		var mtlsConfig *v1.MtlsConfig
		if globalMtlsEnabled {
			mtlsConfig = &v1.MtlsConfig{
				MtlsEnabled:     true,
				MtlsPermissive:  mtlsPermissive,
				RootCertificate: rootCa,
			}
		}

		meshUpstreams := func() []*core.ResourceRef {
			injectedUpstreams := utils.UpstreamsForPods(injectedPods[pilot.Cluster][pilot.Namespace], snap.Upstreams)
			var usRefs []*core.ResourceRef
			for _, us := range injectedUpstreams {
				ref := us.Metadata.Ref()
				usRefs = append(usRefs, &ref)
			}
			return usRefs
		}()

		istioMesh := &v1.Mesh{
			Metadata: core.Metadata{
				Name:   pilot.Name(),
				Labels: discoveryLabels,
			},
			MeshType: &v1.Mesh_Istio{
				Istio: &v1.IstioMesh{
					InstallationNamespace: pilot.Namespace,
					Version:               pilot.Version,
				},
			},
			MtlsConfig: mtlsConfig,
			SmiEnabled: smiEnabled,
			DiscoveryMetadata: &v1.DiscoveryMetadata{
				Cluster:          pilot.Cluster,
				EnableAutoInject: autoInjectionEnabled,
				MtlsConfig:       mtlsConfig,
				Upstreams:        meshUpstreams,
			},
		}
		istioMeshes = append(istioMeshes, istioMesh)
	}

	contextutils.LoggerFrom(ctx).Infow("istio desired meshes", zap.Any("count", len(istioMeshes)))
	return istioMeshes, nil
}

// TODO extract utils like this into solo-kit and test them
func snapShotByCluster(snap *v1.DiscoverySnapshot) map[string]*v1.DiscoverySnapshot {
	snapByCluster := make(map[string]*v1.DiscoverySnapshot)

	createIfNeedBe := func(cluster string) {
		if snapByCluster[cluster] == nil {
			snapByCluster[cluster] = &v1.DiscoverySnapshot{}
		}
	}

	for _, p := range snap.Pods {
		createIfNeedBe(p.GetMetadata().Cluster)
		snapByCluster[p.GetMetadata().Cluster].Pods = append(snapByCluster[p.GetMetadata().Cluster].Pods, p)
	}
	for _, u := range snap.Upstreams {
		createIfNeedBe(u.GetMetadata().Cluster)
		snapByCluster[u.GetMetadata().Cluster].Upstreams = append(snapByCluster[u.GetMetadata().Cluster].Upstreams, u)
	}
	for _, d := range snap.Deployments {
		createIfNeedBe(d.GetMetadata().Cluster)
		snapByCluster[d.GetMetadata().Cluster].Deployments = append(snapByCluster[d.GetMetadata().Cluster].Deployments, d)
	}
	for _, t := range snap.Tlssecrets {
		createIfNeedBe(t.GetMetadata().Cluster)
		snapByCluster[t.GetMetadata().Cluster].Tlssecrets = append(snapByCluster[t.GetMetadata().Cluster].Tlssecrets, t)
	}
	return snapByCluster
}
