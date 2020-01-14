package appmesh

import (
	"context"
	"regexp"
	"strings"

	"github.com/rotisserie/eris"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/common"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/mesh-projects/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type appmeshDiscoveryPlugin struct {
	clientBuilder ClientBuilder
}

func NewAppmeshDiscoverySyncer(writeNamespace string, meshReconciler v1.MeshReconciler,
	meshIngressReconciler v1.MeshIngressReconciler, clientBuilder ClientBuilder) v1.DiscoverySyncer {
	return common.NewDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		meshIngressReconciler,
		&appmeshDiscoveryPlugin{
			clientBuilder: clientBuilder,
		},
	)
}

func (p *appmeshDiscoveryPlugin) MeshType() string {
	return common.AppmeshMeshID
}

var discoveryLabels = map[string]string{
	"discovered_by": "appmesh-mesh-discovery",
}

func (p *appmeshDiscoveryPlugin) DiscoveryLabels() map[string]string {
	return discoveryLabels
}

func (p *appmeshDiscoveryPlugin) DesiredMeshes(ctx context.Context, writeNamespace string, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	isEksCluster, awsRegion := detectEksCluster(snap.Pods)

	if !isEksCluster {
		return nil, nil
	}

	logger := contextutils.LoggerFrom(ctx)

	var uniqueMeshes []string
	awsClient, err := p.clientBuilder.GetClientInstance(awsRegion)
	if err != nil {
		return nil, eris.Wrapf(err, "failed getting aws client instance")
	}
	meshNames, err := awsClient.ListMeshes(ctx)
	if err != nil {
		return nil, eris.Wrapf(err, "failed listing appmesh meshes")
	}
	for _, meshName := range meshNames {
		uniqueMeshes = append(uniqueMeshes, meshName)
	}

	var appmeshMeshes v1.MeshList
	for _, meshName := range uniqueMeshes {

		injectedUpstreams, err := detectInjectedAppmeshUpstreams(meshName, snap.Pods, snap.Upstreams)
		if err != nil {
			logger.Errorf("failed to detect appmesh injected pods: %v", err)
			continue
		}

		meshUpstreams := func() []*core.ResourceRef {
			var usRefs []*core.ResourceRef
			for _, us := range injectedUpstreams {
				ref := us.Metadata.Ref()
				usRefs = append(usRefs, &ref)
			}
			return usRefs
		}()

		appmeshMesh := &v1.Mesh{
			Metadata: core.Metadata{
				Name:      "appmesh-" + meshName,
				Namespace: writeNamespace,
				Labels:    discoveryLabels,
			},
			MeshType: &v1.Mesh_AwsAppMesh{
				AwsAppMesh: &v1.AwsAppMesh{
					Region: awsRegion,
				},
			},
			DiscoveryMetadata: &v1.DiscoveryMetadata{
				Upstreams: meshUpstreams,
			},
		}
		appmeshMeshes = append(appmeshMeshes, appmeshMesh)
	}

	return appmeshMeshes.Sort(), nil
}

const (
	awsPodPrefix    = "aws-node"
	awsPodNamespace = "kube-system"
)

func detectEksCluster(pods kubernetes.PodList) (bool, string) {
	for _, pod := range pods {
		if pod.Namespace != awsPodNamespace || !strings.HasPrefix(pod.Name, awsPodPrefix) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			awsRegion := detectAwsRegionFromImage(container.Image)
			if awsRegion != "" {
				return true, awsRegion
			}
		}
	}
	return false, ""
}

var awsImageRegionRegex = regexp.MustCompile("([.][a-z]+[-][a-z]+[-][0-9][.])")

func detectAwsRegionFromImage(image string) string {
	imageTag := awsImageRegionRegex.FindString(image)
	return strings.ReplaceAll(imageTag, ".", "")
}

func detectInjectedAppmeshUpstreams(meshName string, pods kubernetes.PodList, upstreams gloov1.UpstreamList) (gloov1.UpstreamList, error) {
	appmeshConfig, err := NewAwsAppMeshConfiguration(meshName, pods, upstreams)
	if err != nil {
		return nil, err
	}
	return appmeshConfig.InjectedUpstreams(), nil
}
