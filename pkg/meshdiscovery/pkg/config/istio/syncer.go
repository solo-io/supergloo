package istio

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/clientset"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/utils"
	"go.uber.org/zap"
)

const (
	istio      = "istio"
	pilot      = "pilot"
	istioPilot = istio + "-" + pilot

	istioSelector = "istio-mesh-discovery"

	injectionLabel = "istio-injection"
)

var (
	DiscoverySelector = map[string]string{
		common.SelectorPrefix: istioSelector,
	}
)

type istioConfigSyncer struct {
	ctx context.Context
	cs  *clientset.Clientset

	reconciler v1.MeshReconciler
}

func newIstioConfigSyncer(ctx context.Context, cs *clientset.Clientset) *istioConfigSyncer {
	meshReconciler := v1.NewMeshReconciler(cs.Discovery.Mesh)
	return &istioConfigSyncer{ctx: ctx, cs: cs, reconciler: meshReconciler}
}

func (s *istioConfigSyncer) Sync(ctx context.Context, snap *v1.IstioDiscoverySnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("istio-translation-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)

	pods := snap.Pods.List()
	installs := snap.Installs.List()
	meshes := snap.Meshes.List()

	fields := []interface{}{
		zap.Int("pods", len(pods)),
		zap.Int("meshes", len(meshes)),
		zap.Int("installs", len(installs)),
		// zap.Int("meshpolicies", len(snap.Meshpolicies)),
	}
	logger.Infow("begin sync", fields...)
	defer logger.Infow("end sync", fields...)
	logger.Debugf("full snapshot: %v", snap)

	existingMeshes := utils.GetMeshes(meshes, utils.IstioMeshFilterFunc)
	existingInstalls := utils.GetInstalls(installs, utils.IstioInstallFilterFunc)

	pilotPods := utils.FilerPodsByNamePrefix(pods, istio)
	if len(pilotPods) == 0 {
		logger.Debugf("no pilot pods found in istio pod list")
		return nil
	}

	var discoveredMeshes v1.MeshList
	for _, pilotPod := range pilotPods {
		if strings.Contains(pilotPod.Name, istioPilot) {
			mesh, err := constructDiscoveryData(ctx, pilotPod, existingInstalls)
			if err != nil {
				return err
			}
			logger.Debugf("successfully discovered mesh data for %v", mesh)
			discoveredMeshes = append(discoveredMeshes, mesh)
		}
	}

	mergedMeshes, err := mergeMeshes(discoveredMeshes, existingMeshes)
	if err != nil {
		return err
	}

	reconcileOpts := clients.ListOpts{
		Ctx:      s.ctx,
		Selector: DiscoverySelector,
	}
	return s.reconciler.Reconcile("", mergedMeshes, nil, reconcileOpts)
}

func mergeMeshes(discoveredMeshes, existingMeshes v1.MeshList) (v1.MeshList, error) {
	var mergedMeshes v1.MeshList
	for _, discoveredMesh := range discoveredMeshes {
		meshExists := false
		for _, existingMesh := range existingMeshes {
			istioMesh := existingMesh.GetIstio()
			if istioMesh == nil {
				continue
			}

			// This discovered mesh already exists, update the discovery data
			if istioMesh.InstallationNamespace == discoveredMesh.DiscoveryMetadata.InstallationNamespace {
				existingMesh.DiscoveryMetadata = discoveredMesh.DiscoveryMetadata
				meshExists = true
				break
			}
		}
		if meshExists {
			continue
		}

		discoveredMesh.MeshType = &v1.Mesh_Istio{
			Istio: &v1.IstioMesh{
				InstallationNamespace: discoveredMesh.DiscoveryMetadata.InstallationNamespace,
				IstioVersion:          discoveredMesh.DiscoveryMetadata.MeshVersion,
			},
		}
		discoveredMesh.MtlsConfig = discoveredMesh.DiscoveryMetadata.MtlsConfig
		if discoveredMesh.Metadata.Name == "" || discoveredMesh.Metadata.Namespace == "" {
			discoveredMesh.Metadata = core.Metadata{
				Namespace: getWriteNamespace(),
				Name:      fmt.Sprintf("istio-%s", discoveredMesh.DiscoveryMetadata.InstallationNamespace),
				// Very important to add this selector or the reconcile will not pick it up later
				Labels: DiscoverySelector,
			}
		}

		mergedMeshes = append(mergedMeshes, discoveredMesh)
	}
	mergedMeshes = append(mergedMeshes, existingMeshes...)
	return mergedMeshes, nil
}

func getWriteNamespace() string {
	if writeNamespace := os.Getenv("POD_NAMESPACE"); writeNamespace != "" {
		return writeNamespace
	}
	return "supergloo-system"
}

func constructDiscoveryData(ctx context.Context, istioPilotPod *v1.Pod, existingInstalls v1.InstallList) (*v1.Mesh, error) {
	logger := contextutils.LoggerFrom(ctx)
	mesh := &v1.Mesh{}

	istioVersion, err := getVersionFromPod(istioPilotPod)
	if err != nil {
		logger.Debugf("unable to find version from pod %v", istioPilotPod)
		return nil, err
	}

	discoveryData := &v1.DiscoveryMetadata{
		InstallationNamespace:  istioPilotPod.Namespace,
		MeshVersion:            istioVersion,
		InjectedNamespaceLabel: injectionLabel,
	}
	mesh.DiscoveryMetadata = discoveryData

	// If install crd exists, overwrite discovery data
	for _, install := range existingInstalls {
		meshInstall := install.GetMesh()
		if meshInstall == nil {
			continue
		}
		istioMeshInstall := meshInstall.GetIstioMesh()
		if istioMeshInstall == nil {
			continue
		}
		// This install refers to the current mesh
		if install.InstallationNamespace == istioPilotPod.Namespace {
			if istioMeshInstall.EnableMtls {
				mesh.DiscoveryMetadata.MtlsConfig = &v1.MtlsConfig{
					MtlsEnabled:     true,
					RootCertificate: istioMeshInstall.CustomRootCert,
				}
			}
			mesh.Metadata = install.Metadata
			// Need to explicitly set this to "" so it doesn't attempt to overwrite it.
			mesh.Metadata.ResourceVersion = ""
		}
	}

	return mesh, nil
}

func getVersionFromPod(pod *v1.Pod) (string, error) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		if strings.Contains(container.Image, istio) && strings.Contains(container.Image, pilot) {
			return utils.ImageVersion(container.Image)
		}
	}
	return "", errors.Errorf("unable to find pilot container from pod")
}
