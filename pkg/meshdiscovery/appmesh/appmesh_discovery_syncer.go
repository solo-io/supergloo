package appmesh

import (
	"context"
	"regexp"
	"strings"

	appmeshtranslator "github.com/solo-io/supergloo/pkg/translator/appmesh"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"

	"github.com/solo-io/supergloo/pkg/config/appmesh"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/common"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type appmeshDiscoveryPlugin struct {
	clientBuilder appmesh.ClientBuilder
	secrets       gloov1.SecretClient
}

func NewAppmeshDiscoverySyncer(writeNamespace string, meshReconciler v1.MeshReconciler, clientBuilder appmesh.ClientBuilder, secrets gloov1.SecretClient) v1.DiscoverySyncer {
	return common.NewDiscoverySyncer(
		writeNamespace,
		meshReconciler,
		&appmeshDiscoveryPlugin{
			clientBuilder: clientBuilder,
			secrets:       secrets,
		},
	)
}

func (p *appmeshDiscoveryPlugin) MeshType() string {
	return "appmesh"
}

var discoveryLabels = map[string]string{
	"discovered_by": "appmesh-mesh-discovery",
}

func (p *appmeshDiscoveryPlugin) DiscoveryLabels() map[string]string {
	return discoveryLabels
}

const (
	DefaultVirtualNodeLabel = "virtual-node"
)

func (p *appmeshDiscoveryPlugin) DesiredMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	isEksCluster, awsRegion := detectEksCluster(snap.Pods)

	if !isEksCluster {
		return nil, nil
	}

	awsSecrets, err := detectAwsSecrets(ctx, p.secrets)
	if err != nil {
		return nil, err
	}
	if len(awsSecrets) == 0 {
		return nil, nil
	}

	logger := contextutils.LoggerFrom(ctx)

	uniqueMeshes := make(map[string]*core.ResourceRef)
	for _, secret := range awsSecrets {
		awsClient, err := p.clientBuilder.GetClientInstance(secret, awsRegion)
		if err != nil {
			logger.Errorf("failed getting aws client instance: %v", err)
			continue
		}
		meshNames, err := awsClient.ListMeshes(ctx)
		if err != nil {
			logger.Errorf("failed listing appmesh meshes: %v", err)
			continue
		}
		for _, meshName := range meshNames {
			uniqueMeshes[meshName] = secret
		}
	}

	var appmeshMeshes v1.MeshList
	for meshName, secret := range uniqueMeshes {

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
				Name:   "appmesh-" + meshName,
				Labels: discoveryLabels,
			},
			MeshType: &v1.Mesh_AwsAppMesh{
				AwsAppMesh: &v1.AwsAppMesh{
					AwsSecret:        secret,
					Region:           awsRegion,
					VirtualNodeLabel: DefaultVirtualNodeLabel,
				},
			},
			DiscoveryMetadata: &v1.DiscoveryMetadata{
				Upstreams: meshUpstreams,
			},
		}
		appmeshMeshes = append(appmeshMeshes, appmeshMesh)
	}

	return appmeshMeshes, nil
}

func detectAwsSecrets(ctx context.Context, client gloov1.SecretClient) ([]*core.ResourceRef, error) {
	var awsSecrets []*core.ResourceRef

	secrets, err := client.List("", clients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		awsSecret := secret.GetAws()
		if awsSecret == nil {
			continue
		}
		ref := secret.Metadata.Ref()
		awsSecrets = append(awsSecrets, &ref)
	}

	return awsSecrets, nil
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
	appmeshConfig, err := appmeshtranslator.NewAwsAppMeshConfiguration(meshName, pods, upstreams)
	if err != nil {
		return nil, err
	}
	return appmeshConfig.InjectedUpstreams(), nil
}
