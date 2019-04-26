package linkerd

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/utils"
	"go.uber.org/zap"
)

const (
	linkerd    = "linkerd"
	controller = "controller"
	// Hardcoded name from helm chart for linkerd controller
	linkerdController = linkerd + "-" + controller

	linkerdSelector = "linkerd-mesh-discovery"
)

var (
	DiscoverySelector = map[string]string{
		utils.SelectorPrefix: linkerdSelector,
	}
)

type linkerdDiscoverySyncer struct{}

func NewLinkerdDiscoverySyncer() *linkerdDiscoverySyncer {
	return &linkerdDiscoverySyncer{}
}

func (s *linkerdDiscoverySyncer) DiscoverMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("linkerd-mesh-discovery-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)

	pods := snap.Pods.List()
	installs := snap.Installs.List()

	fields := []interface{}{
		zap.Int("pods", len(pods)),
		zap.Int("installs", len(installs)),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	existingInstalls := utils.GetActiveInstalls(installs, utils.LinkerdInstallFilterFunc)

	var newMeshes v1.MeshList
	for _, linkerdPod := range utils.SelectRunningPods(pods) {
		if strings.Contains(linkerdPod.Name, linkerdController) {
			mesh, err := constructDiscoveredMesh(ctx, linkerdPod, existingInstalls)
			if err != nil {
				return nil, err
			}
			logger.Debugf("successfully discovered mesh data for %v", mesh)
			newMeshes = append(newMeshes, mesh)
		}
	}

	return newMeshes, nil
}

func constructDiscoveredMesh(ctx context.Context, mainPod *v1.Pod, existingInstalls v1.InstallList) (*v1.Mesh, error) {
	logger := contextutils.LoggerFrom(ctx)

	linkerdVersion, err := utils.GetVersionFromPodWithMatchers(mainPod, []string{linkerd, controller})
	if err != nil {
		logger.Debugf("unable to find version from pod %v", mainPod)
		return nil, err
	}

	mesh := utils.BasicMeshInfo(mainPod, DiscoverySelector, linkerd)

	mesh.MeshType = &v1.Mesh_Linkerd{
		Linkerd: &v1.LinkerdMesh{
			Version:               linkerdVersion,
			InstallationNamespace: mainPod.Namespace,
		},
	}
	// If install crd exists, overwrite discovery data
	for _, install := range existingInstalls {
		meshInstall := install.GetMesh()
		if meshInstall == nil {
			continue
		}
		linkerdMeshInstall := meshInstall.GetLinkerd()
		if linkerdMeshInstall == nil {
			continue
		}
		// This install refers to the current mesh
		if install.InstallationNamespace == mainPod.Namespace {
			mesh.MtlsConfig = &v1.MtlsConfig{
				MtlsEnabled: linkerdMeshInstall.EnableMtls,
			}
			mesh.Metadata = install.Metadata
			// Set label to be aware of discovered nature
			mesh.Metadata.Labels = DiscoverySelector
			// Need to explicitly set this to "" so it doesn't attempt to overwrite it.
			mesh.Metadata.ResourceVersion = ""
		}
	}

	return mesh, nil
}
