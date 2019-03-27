package webhook

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/webhook/patch"
	"github.com/solo-io/supergloo/pkg/webhook/plugins"
	corev1 "k8s.io/api/core/v1"
)

type SidecarInjectionHandler interface {
	// Generate a JSONPatch that injects the given pod with the proxy sidecar(s)
	GetSidecarPatch(ctx context.Context, candidatePod *corev1.Pod) (patchRequired bool, patch []byte, err error)
}

type injectionHandler struct {
	plugins []plugins.InjectionPlugin
}

var (
	handler SidecarInjectionHandler
	mutex   sync.Mutex
)

func RegisterSidecarInjectionHandler() {
	mutex.Lock()
	defer mutex.Unlock()
	handler = injectionHandler{plugins: plugins.GetPlugins()}
}

func GetInjectionHandler() SidecarInjectionHandler {
	mutex.Lock()
	defer mutex.Unlock()
	if handler == nil {
		panic("SidecarInjectionHandler is nil. Make sure to register it before trying too use it.")
	}
	return handler
}

// NOTE: The plugins are currently unaware of each other. This is fine, as we currently only handle auto-injection for
// AWS App Mesh. If and when we add support for other meshes, we might want to split the plugin loop so that in case of
// multiple matches for different meshes we can check for and handle potential conflicts between the patches.
func (h injectionHandler) GetSidecarPatch(ctx context.Context, candidatePod *corev1.Pod) (patchRequired bool, patches []byte, err error) {
	logger := contextutils.LoggerFrom(ctx)

	var jsonPatches []patch.JSONPatchOperation
	for _, plugin := range h.plugins {

		// Get the meshes that have auto-injection enabled
		autoInjectionMeshes, err := plugin.GetAutoInjectMeshes(ctx)
		if err != nil {
			return false, nil, errors.Wrapf(err, "failed to list auto-injection enabled meshes for plugin %s", plugin.Name())
		}

		// Check whether the pod matches any of the pod selection criteria defined for the meshes
		matchingMeshes, err := plugin.CheckMatch(ctx, candidatePod, autoInjectionMeshes)
		if err != nil {
			return false, nil, errors.Wrapf(err, "failed to check whether candidate pod matches auto-injection "+
				"criteria for meshes in plugin %s", plugin.Name())
		}

		patches, err := plugin.GetSidecarPatch(ctx, candidatePod, matchingMeshes)
		if err != nil {
			return false, nil, errors.Wrapf(err, "failed to create sidecar patch in plugin %s", plugin.Name())
		}
		jsonPatches = append(jsonPatches, patches...)
	}

	if len(jsonPatches) == 0 {
		logger.Info("pod does not match any mesh auto-injection selector, admit it without patching")
		return false, nil, nil
	}

	logger.Infof("the following patches will be applied to the pod: %v", jsonPatches)
	patchesBytes, err := json.Marshal(jsonPatches)
	if err != nil {
		return false, nil, errors.Wrapf(err, "failed to marshal patches to JSON")
	}
	return true, patchesBytes, nil
}
