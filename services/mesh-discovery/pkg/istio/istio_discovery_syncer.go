package istio

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/utils"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"go.uber.org/zap"

	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"

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
	// TODO bring this into s-k for multicluster
	jobGetter        batchv1client.JobsGetter
	resourceDetector IstioResourceDetector
}

func NewIstioDiscoverySyncer(writeNamespace string, meshReconciler v1.MeshReconciler, meshPolicyClient v1alpha1.MeshPolicyClient, crdGetter CrdGetter, jobGetter batchv1client.JobsGetter) v1.DiscoverySyncer {
	return common.NewDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		&istioDiscoveryPlugin{meshPolicyClient: meshPolicyClient, crdGetter: crdGetter, jobGetter: jobGetter, resourceDetector: NewIstioResourceDetector()},
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
	// for every pilot, do the following
	pilots := p.resourceDetector.DetectPilotDeployments(ctx, snap.Deployments)

	if len(pilots) == 0 {
		return nil, nil
	}

	injectedPods := p.resourceDetector.DetectInjectedIstioPods(ctx, snap.Pods)

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
		// TODO joekelley punting on jobs for now, need multicluster jobs client in solo-kit
		// postInstallComplete, err := p.resourceDetector.DetectPostInstallJobComplete(p.jobGetter, pilot.Namespace)
		// if err != nil {
		//	contextutils.LoggerFrom(ctx).Errorf("failed to detect if post-install jobs have finished: %v", err)
		//	continue
		// }
		//
		// if !postInstallComplete {
		//	continue
		// }

		var autoInjectionEnabled bool
		sidecarInjector, err := snap.Deployments.Find(pilot.Namespace, "istio-sidecar-injector")
		if err == nil && (sidecarInjector.Spec.Replicas == nil || *sidecarInjector.Spec.Replicas > 0) {
			autoInjectionEnabled = true
		}

		var smiEnabled bool
		smiAdapter, err := snap.Deployments.Find(pilot.Namespace, "smi-adapter-istio")
		if err == nil && (smiAdapter.Spec.Replicas == nil || *smiAdapter.Spec.Replicas > 0) {
			smiEnabled = true
		}

		// https://istio.io/docs/tasks/security/plugin-ca-cert/#plugging-in-the-existing-certificate-and-key
		var rootCa *core.ResourceRef
		// TODO joekelley support TLS secrets
		// customRootCa, err := snap.Tlssecrets.Find(pilot.Namespace, "cacerts")
		// if err == nil {
		//	root := customRootCa.Metadata.Ref()
		//	rootCa = &root
		// }

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

	return istioMeshes, nil
}
