package linkerd

import (
	"context"
	"fmt"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	coreapi "github.com/solo-io/mesh-projects/pkg/api/v1/core"
	globalcommon "github.com/solo-io/mesh-projects/services/common"
	"github.com/solo-io/mesh-projects/services/internal/utils"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common/injectedpods"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"go.uber.org/zap"
)

type linkerdDiscoveryPlugin struct{}

var (
	discoveryLabels = map[string]string{
		"discovered_by": "linkerd-mesh-discovery",
	}

	// sidecar proxy upstreams should have this as a substring in their name
	sidecarUpstreamNameSubstring = fmt.Sprintf("-%s-", common.LinkerdMeshID)
)

func NewLinkerdDiscoverySyncer(writeNamespace string, meshReconciler v1.MeshReconciler,
	reconciler v1.MeshIngressReconciler) v1.DiscoverySyncer {

	return common.NewDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		reconciler,
		&linkerdDiscoveryPlugin{},
	)
}

type linkerdControllerDeployment struct {
	version, namespace, cluster string
}

func (p *linkerdDiscoveryPlugin) MeshType() string {
	return common.LinkerdMeshID
}

func (p *linkerdDiscoveryPlugin) DiscoveryLabels() map[string]string {
	return discoveryLabels
}

func (p *linkerdDiscoveryPlugin) DesiredMeshes(ctx context.Context, writeNamespace string, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	linkerdControllers := detectLinkerdControllers(ctx, snap.Deployments)
	contextutils.LoggerFrom(ctx).Infow("linkerd controllers", zap.Any("length", len(linkerdControllers)))

	if len(linkerdControllers) == 0 {
		return nil, nil
	}

	injectedPods := injectedpods.NewDetector(detectInjectedLinkerdPod).DetectInjectedPods(ctx, snap.Pods)

	upstreamsByCluster := utils.UpstreamsByCluster(snap.Upstreams)
	deploymentsByCluster := utils.DeploymentsByCluster(snap.Deployments)

	var linkerdMeshes v1.MeshList
	for _, linkerdController := range linkerdControllers {
		var autoInjectionEnabled bool
		sidecarInjector, err := deploymentsByCluster[linkerdController.cluster].Find(linkerdController.namespace, "linkerd-proxy-injector")
		if err == nil && (sidecarInjector.Spec.Replicas == nil || *sidecarInjector.Spec.Replicas > 0) {
			autoInjectionEnabled = true
		}

		meshUpstreams := findUpstreamsToReport(injectedPods, linkerdController, upstreamsByCluster)

		// mtls always enabled with linkerd
		mtlsConfig := &v1.MtlsConfig{MtlsEnabled: true}

		meshMetadata := core.Metadata{
			Name:      linkerdController.name(),
			Namespace: writeNamespace,
			Labels:    discoveryLabels,
		}
		meshRef := meshMetadata.Ref()
		ingressRef := v1.BuildMeshIngress(&meshRef, meshMetadata.Labels).Metadata.Ref()

		linkerdMesh := &v1.Mesh{
			Metadata: meshMetadata,
			MeshType: &v1.Mesh_Linkerd{
				Linkerd: &v1.LinkerdMesh{
					Installation: &v1.MeshInstallation{
						InstallationNamespace: linkerdController.namespace,
						Version:               linkerdController.version,
					},
				},
			},
			MtlsConfig: mtlsConfig,
			DiscoveryMetadata: &v1.DiscoveryMetadata{
				Cluster:          linkerdController.cluster,
				EnableAutoInject: autoInjectionEnabled,
				MtlsConfig:       mtlsConfig,
				Upstreams:        meshUpstreams,
			},
			EntryPoint: &coreapi.ClusterResourceRef{
				Resource: ingressRef,
			},
		}
		linkerdMeshes = append(linkerdMeshes, linkerdMesh)
	}

	contextutils.LoggerFrom(ctx).Infow("linkerd desired meshes", zap.Any("count", len(linkerdMeshes)))
	return linkerdMeshes, nil
}

// report all injected upstreams that are not sidecar proxies
func findUpstreamsToReport(injectedPods injectedpods.InjectedPods,
	linkerdController linkerdControllerDeployment,
	upstreamsByCluster map[string]gloov1.UpstreamList,
) []*core.ResourceRef {
	injectedUpstreams := utils.UpstreamsForPods(injectedPods[linkerdController.cluster][linkerdController.namespace], upstreamsByCluster[linkerdController.cluster])
	var usRefs []*core.ResourceRef
	for _, us := range injectedUpstreams {
		if strings.Contains(us.Metadata.GetName(), sidecarUpstreamNameSubstring) {
			continue
		}
		ref := us.Metadata.Ref()
		usRefs = append(usRefs, &ref)
	}
	return usRefs
}

func detectLinkerdControllers(ctx context.Context, deployments kubernetes.DeploymentList) []linkerdControllerDeployment {
	var linkerdControllers []linkerdControllerDeployment
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if strings.Contains(container.Image, "linkerd-io/controller") {
				// TODO there can be > 1 controller image per pod, do we care?
				parsedImage, err := globalcommon.NewImageNameParser().Parse(container.Image)
				if err != nil {
					contextutils.LoggerFrom(ctx).Errorf("invalid or unexpected image format for linkerd controller: %v", container.Image)
					continue
				}

				version := parsedImage.Tag
				if parsedImage.Digest != "" {
					version = parsedImage.Digest
				}
				linkerdControllers = append(linkerdControllers, linkerdControllerDeployment{version: version, namespace: deployment.Namespace, cluster: deployment.ClusterName})
			}
		}
	}
	return linkerdControllers
}

func detectInjectedLinkerdPod(ctx context.Context, pod *kubernetes.Pod) (string, string, bool) {
	for _, container := range pod.Spec.Containers {
		if container.Name == "linkerd-proxy" {
			for _, envVar := range container.Env {
				if envVar.Name == "LINKERD2_PROXY_DESTINATION_SVC_ADDR" {
					controlPlaneAddress := envVar.Value
					controlPlaneService := strings.Split(controlPlaneAddress, ":")[0]
					// linkerd 2.3.0
					controlPlaneNamespace := strings.TrimPrefix(controlPlaneService, "linkerd-destination.")
					// linkerd 2.6.0
					controlPlaneNamespace = strings.TrimPrefix(controlPlaneNamespace, "linkerd-dst.")
					controlPlaneNamespace = strings.TrimSuffix(controlPlaneNamespace, ".svc.cluster.local")

					// special case for the controller itself
					if controlPlaneNamespace == "localhost." {
						controlPlaneNamespace = pod.Namespace
					}

					return pod.ClusterName, controlPlaneNamespace, true
				}
			}
		}
	}
	return "", "", false
}

func (c linkerdControllerDeployment) name() string {
	if c.cluster == "" {
		return "linkerd-" + c.namespace
	}
	// TODO cluster is not restricted to kube name scheme, kebab it
	return "linkerd-" + c.namespace + "-" + c.cluster
}
