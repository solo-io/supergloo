package istio

import (
	"context"
	"strings"

	"github.com/solo-io/go-utils/contextutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/config/common"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/pkg/utils"
)

const (
	istio      = "istio"
	pilot      = "pilot"
	istioPilot = istio + "-" + pilot
)

type istioMeshDiscovery struct {
}

func NewIstioMeshDiscovery() *istioMeshDiscovery {
	return &istioMeshDiscovery{}
}

func (imd *istioMeshDiscovery) DiscoverMeshes(ctx context.Context, snapshot *v1.DiscoverySnapshot, enabled *common.EnabledConfigLoops) error {
	pods := snapshot.Pods.List()
	existingMeshes := utils.GetMeshes(snapshot.Meshes.List(), utils.IstioMeshFilterFunc)
	logger := contextutils.LoggerFrom(ctx)

	if len(existingMeshes) > 0 {
		enabled.SetIstio(true)
		logger.Info("discovered istio mesh")
		return nil
	}

	pilotPods := utils.FilerPodsByNamePrefix(pods, istio)
	if len(pilotPods) == 0 {
		logger.Debugf("no pilot pods found in istio pod list")
		return nil
	}

	for _, pilotPod := range pilotPods {
		if strings.Contains(pilotPod.Name, istioPilot) {
			enabled.SetIstio(true)
			logger.Info("discovered istio mesh")
			break
		}
	}

	return nil
}
