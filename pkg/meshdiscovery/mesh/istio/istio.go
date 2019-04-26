package istio

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/utils"
	"go.uber.org/zap"
)

const (
	istio      = "istio"
	pilot      = "pilot"
	istioPilot = istio + "-" + pilot

	istioSelector = "istio-mesh-discovery"
)

var (
	DiscoverySelector = map[string]string{
		utils.SelectorPrefix: istioSelector,
	}
)

type istioDiscoverySyncer struct{}

func NewIstioDiscoverySyncer() *istioDiscoverySyncer {
	return &istioDiscoverySyncer{}
}

func (s *istioDiscoverySyncer) DiscoverMeshes(ctx context.Context, snap *v1.DiscoverySnapshot) (v1.MeshList, error) {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-mesh-discovery-sync-%v", snap.Hash()))
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

	existingInstalls := utils.GetActiveInstalls(installs, utils.IstioInstallFilterFunc)

	var newMeshes v1.MeshList
	for _, pod := range utils.SelectRunningPods(pods) {
		if strings.Contains(pod.Name, istioPilot) {
			mesh, err := constructDiscoveredMesh(ctx, pod, existingInstalls)
			if err != nil {
				return nil, err
			}
			logger.Debugf("successfully discovered mesh data for %v", mesh)
			newMeshes = append(newMeshes, mesh)
		}
	}

	return newMeshes, nil
}

func constructDiscoveredMesh(ctx context.Context, istioPilotPod *kubernetes.Pod, existingInstalls v1.InstallList) (*v1.Mesh, error) {
	logger := contextutils.LoggerFrom(ctx)

	istioVersion, err := utils.GetVersionFromPodWithMatchers(istioPilotPod, []string{istio, pilot})
	if err != nil {
		logger.Debugf("unable to find version from pod %v", istioPilotPod)
		return nil, err
	}

	mesh := utils.BasicMeshInfo(istioPilotPod, DiscoverySelector, istio)
	mesh.MeshType = &v1.Mesh_Istio{
		Istio: &v1.IstioMesh{
			InstallationNamespace: istioPilotPod.Namespace,
			Version:               istioVersion,
		},
	}
	// If install crd exists, overwrite discovery data
	for _, install := range existingInstalls {
		meshInstall := install.GetMesh()
		if meshInstall == nil {
			continue
		}
		istioMeshInstall := meshInstall.GetIstio()
		if istioMeshInstall == nil {
			continue
		}
		// This install refers to the current mesh
		if install.InstallationNamespace == istioPilotPod.Namespace {
			if istioMeshInstall.EnableMtls {
				mesh.MtlsConfig = &v1.MtlsConfig{
					MtlsEnabled:     true,
					RootCertificate: istioMeshInstall.CustomRootCert,
				}
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
