package linkerd

import (
	"context"
	"strings"

	"github.com/solo-io/supergloo/pkg/meshdiscovery/common"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/utils"
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
	version, namespace string
}

func (p *linkerdDiscoveryPlugin) DesiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	linkerdControllers := detectLinkerdControllers(ctx, snap.Deployments)

	if len(linkerdControllers) == 0 {
		return nil, nil
	}

	injectedPods := detectInjectedLinkerdPods(snap.Pods)

	var linkerdMeshes v1.MeshList
	for _, linkerdController := range linkerdControllers {
		var autoInjectionEnabled bool
		sidecarInjector, err := snap.Deployments.Find(linkerdController.namespace, "linkerd-sidecar-injector")
		if err == nil && (sidecarInjector.Spec.Replicas == nil || *sidecarInjector.Spec.Replicas > 0) {
			autoInjectionEnabled = true
		}

		meshUpstreams := func() []*core.ResourceRef {
			injectedUpstreams := utils.UpstreamsForPods(injectedPods[linkerdController.namespace], snap.Upstreams)
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
				Name:   linkerdController.namespace + "-linkerd",
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
				split := strings.Split(container.Image, ":")
				if len(split) != 2 {
					contextutils.LoggerFrom(ctx).Errorf("invalid or unexpected image format for linkerd controller: %v", container.Image)
					continue
				}
				linkerdControllers = append(linkerdControllers, linkerdControllerDeployment{version: split[1], namespace: deployment.Namespace})
			}
		}
	}
	return linkerdControllers
}

func detectInjectedLinkerdPods(pods kubernetes.PodList) map[string]kubernetes.PodList {
	injectedPods := make(map[string]kubernetes.PodList)
	for _, pod := range pods {
		discoveryNamespace, ok := detectInjectedLinkerdPod(pod)
		if ok {
			injectedPods[discoveryNamespace] = append(injectedPods[discoveryNamespace], pod)
		}
	}
	return injectedPods
}

func detectInjectedLinkerdPod(pod *kubernetes.Pod) (string, bool) {
	for _, container := range pod.Spec.Containers {
		if container.Name == "linkerd-proxy" {
			for _, envVar := range container.Env {
				if envVar.Name == "LINKERD2_PROXY_DESTINATION_SVC_ADDR" {
					controlPlaneAddress := envVar.Value
					controlPlaneService := strings.Split(controlPlaneAddress, ":")[0]
					controlPlaneNamespace := strings.TrimPrefix(controlPlaneService, "linkerd-destination.")
					controlPlaneNamespace = strings.TrimSuffix(controlPlaneNamespace, ".svc.cluster.local")

					return controlPlaneNamespace, true
				}
			}
		}
	}
	return "", false
}
