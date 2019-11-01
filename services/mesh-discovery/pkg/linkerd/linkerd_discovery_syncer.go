package linkerd

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/mesh-projects/pkg/utils"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common/injectedpods"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type linkerdDiscoveryPlugin struct{}

func (p *linkerdDiscoveryPlugin) MeshType() string {
	return "linkerd"
}

var discoveryLabels = map[string]string{
	"discovered_by": "linkerd-mesh-discovery",
}

func (p *linkerdDiscoveryPlugin) DiscoveryLabels() map[string]string {
	return discoveryLabels
}

func NewLinkerdDiscoverySyncer(writeNamespace string, meshReconciler v1.MeshReconciler) v1.DiscoverySyncer {
	return common.NewDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		&linkerdDiscoveryPlugin{},
	)
}

type linkerdControllerDeployment struct {
	version, namespace, cluster string
}

func (p *linkerdDiscoveryPlugin) DesiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	linkerdControllers := detectLinkerdControllers(ctx, snap.Deployments)

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

		meshUpstreams := func() []*core.ResourceRef {
			injectedUpstreams := utils.UpstreamsForPods(injectedPods[linkerdController.cluster][linkerdController.namespace], upstreamsByCluster[linkerdController.cluster])
			var usRefs []*core.ResourceRef
			for _, us := range injectedUpstreams {
				ref := us.Metadata.Ref()
				usRefs = append(usRefs, &ref)
			}
			return usRefs
		}()

		// mtls always enabled with linkerd
		mtlsConfig := &v1.MtlsConfig{MtlsEnabled: true}

		linkerdMesh := &v1.Mesh{
			Metadata: core.Metadata{
				Name:   linkerdController.name(),
				Labels: discoveryLabels,
			},
			MeshType: &v1.Mesh_Linkerd{
				Linkerd: &v1.LinkerdMesh{
					InstallationNamespace: linkerdController.namespace,
					Version:               linkerdController.version,
				},
			},
			MtlsConfig: mtlsConfig,
			DiscoveryMetadata: &v1.DiscoveryMetadata{
				Cluster:          linkerdController.cluster,
				EnableAutoInject: autoInjectionEnabled,
				MtlsConfig:       mtlsConfig,
				Upstreams:        meshUpstreams,
			},
		}
		linkerdMeshes = append(linkerdMeshes, linkerdMesh)
	}

	return linkerdMeshes, nil
}

func detectLinkerdControllers(ctx context.Context, deployments kubernetes.DeploymentList) []linkerdControllerDeployment {
	var linkerdControllers []linkerdControllerDeployment
	for _, deployment := range deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if strings.Contains(container.Image, "linkerd-io/controller") {
				// TODO there can be > 1 controller image per pod, do we care?
				split := strings.Split(container.Image, ":")
				if len(split) != 2 {
					contextutils.LoggerFrom(ctx).Errorf("invalid or unexpected image format for linkerd controller: %v", container.Image)
					continue
				}
				linkerdControllers = append(linkerdControllers, linkerdControllerDeployment{version: split[1], namespace: deployment.Namespace, cluster: deployment.ClusterName})
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
