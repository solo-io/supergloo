package appmesh

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/config/appmesh"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/utils"
	"go.uber.org/zap"
)

const (
	aws          = "aws"
	auth         = "auth"
	node         = "node"
	awsConfigMap = aws + "-" + auth
	awsNode      = aws + "-" + node
	kubeSysyem   = "kube-system"

	appmeshSelector = "appmesh-mesh-discovery"
)

var (
	awsPod = []string{aws, node}

	DiscoverySelector = map[string]string{
		utils.SelectorDiscoveredByPrefix: appmeshSelector,
		utils.SelectorCreatedByPrefix:    utils.SelectorCreatedByValue,
	}
)

type appmeshDiscoverySyncer struct {
	cb      appmesh.ClientBuilder
	secrets gloov1.SecretClient
}

func NewAppmeshDiscoverySyncer(cb appmesh.ClientBuilder, secrets gloov1.SecretClient) *appmeshDiscoverySyncer {
	return &appmeshDiscoverySyncer{cb: cb, secrets: secrets}
}

func (s *appmeshDiscoverySyncer) DiscoverMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("appmesh-mesh-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)

	pods := snap.Pods.List()
	installs := snap.Installs.List()
	configMaps := snap.Configmaps.List()

	fields := []interface{}{
		zap.Int("pods", len(pods)),
		zap.Int("installs", len(installs)),
		zap.Int("configmaps", len(configMaps)),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	amd := newAppmeshDiscoveryData(ctx, configMaps, pods)
	if !amd.eksExists() {
		return nil, nil
	}

	// For the next 2 errors return nil so it doesn't crash other discovery not having this
	// Maybe pull this logic out, but better it start working later if it's added
	glooSecret, err := s.getAwsSecret()
	if err != nil {
		logger.Errorw("could not find secret to authenticate with aws", zap.Error(err))
		return nil, nil
	}
	secretRef := glooSecret.Metadata.Ref()

	awsClient, err := s.cb.GetClientInstance(&secretRef, amd.region)
	if err != nil {
		logger.Errorw("authentication with aws failed for the given credentials",
			zap.Error(err), zap.String("secret", secretRef.Key()),
		)
		return nil, nil
	}

	newMeshes, err := constructAwsMeshes(ctx, awsClient, amd.region, &secretRef)
	if err != nil {
		return nil, errors.Wrapf(err, "could not construct meshes for appmesh")
	}

	return newMeshes, nil
}

func (s *appmeshDiscoverySyncer) getAwsSecret() (*gloov1.Secret, error) {
	secrets, err := s.secrets.List("", clients.ListOpts{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list secrets")
	}
	for _, secret := range secrets {
		if secret.GetAws() != nil {
			return secret, nil
		}
	}
	return nil, errors.Errorf("could not find an AWS secret to use for discovery")
}

type appmeshDiscoveryData struct {
	configMaps skkube.ConfigMapList
	pods       skkube.PodList
	region     string
}

func newAppmeshDiscoveryData(ctx context.Context, configMaps skkube.ConfigMapList, pods skkube.PodList) *appmeshDiscoveryData {
	logger := contextutils.LoggerFrom(ctx)

	amd := &appmeshDiscoveryData{}
	for _, cm := range configMaps {
		if cm.Namespace == kubeSysyem && cm.Name == awsConfigMap {
			logger.Debugw("found possible aws config map", zap.Any("configmap", cm))
			amd.configMaps = append(amd.configMaps, cm)
		}
	}

	for _, pod := range pods {
		if pod.Namespace == kubeSysyem && utils.StringContainsAll(awsPod, pod.Name) {
			logger.Debugw("found possible aws node pod", zap.Any("pod", pod))
			region, err := getAwsRegionFromPod(pod)
			if err != nil {
				logger.Debugw("could not find aws from node pod", zap.Any("pod", pod))
			}
			amd.pods = append(amd.pods, pod)
			amd.region = region
		}
	}
	return amd
}

func (amd *appmeshDiscoveryData) eksExists() bool {
	return len(amd.configMaps) > 0 && len(amd.pods) > 0
}

func getAwsRegionFromPod(pod *skkube.Pod) (string, error) {
	for _, container := range pod.Spec.Containers {
		if container.Name == awsNode {
			region, err := awsRegionFromImage(container.Image)
			if err != nil {
				return "", err
			}
			return region, nil
		}
	}
	return "", errors.Errorf("unable get region from container in pod %v", pod.GetMetadata().Ref().Key())
}

func awsRegionFromImage(image string) (string, error) {
	regex := regexp.MustCompile("([.][a-z]+[-][a-z]+[-][0-9][.])")
	imageTag := regex.FindString(image)
	if imageTag == "" {
		return "", errors.Errorf("unable to find image version for image: %s", image)
	}
	return strings.ReplaceAll(imageTag, ".", ""), nil
}

func constructAwsMeshes(ctx context.Context, client appmesh.Client, region string, secret *core.ResourceRef) (v1.MeshList, error) {
	var meshes v1.MeshList

	apiMeshes, err := client.ListMeshes(ctx)
	if err != nil {
		return nil, err
	}
	for _, apiMesh := range apiMeshes {
		metadata := meshMeta(apiMesh, DiscoverySelector)
		mesh := &v1.Mesh{
			Metadata: metadata,
			MeshType: &v1.Mesh_AwsAppMesh{
				AwsAppMesh: &v1.AwsAppMesh{
					EnableAutoInject: true,
					AwsSecret:        secret,
					Region:           region,
				},
			},
			DiscoveryMetadata: &v1.DiscoveryMetadata{},
		}
		meshes = append(meshes, mesh)
	}

	return meshes, nil
}

func meshMeta(appmeshName string, discoverySelector map[string]string) core.Metadata {
	metadata := core.Metadata{
		Namespace: utils.MeshWriteNamespace(),
		Name:      appmeshName,
		Labels:    discoverySelector,
	}
	return metadata
}
