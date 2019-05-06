package linkerd

import (
	"context"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"go.uber.org/zap"
)

type glooLinkerdMtlsPlugin struct {
	cs *clientset.Clientset
}

func NewGlooLinkerdMtlsPlugin(cs *clientset.Clientset) *glooLinkerdMtlsPlugin {
	return &glooLinkerdMtlsPlugin{cs: cs}
}

func (pl *glooLinkerdMtlsPlugin) HandleMeshes(ctx context.Context, ingress *v1.MeshIngress, meshes v1.MeshList) error {
	if ingress == nil {
		return nil
	}
	ctx = contextutils.WithLoggerValues(ctx,
		zap.String("plugin", "linkerd-gloo-mtls"),
		zap.String("mesh-ingress", ingress.Metadata.Ref().Key()),
	)
	logger := contextutils.LoggerFrom(ctx)

	var linkerdMeshes v1.MeshList
	for _, mesh := range meshes {
		if linkerdMesh := mesh.GetLinkerd(); linkerdMesh != nil {
			linkerdMeshes = append(linkerdMeshes, mesh)
		}
	}

	linkerdEnabled := len(linkerdMeshes) > 0
	logger.Debugf("linkerd mtls enabled")

	glooSettings, err := pl.cs.Supergloo.Settings.List(ingress.InstallationNamespace, clients.ListOpts{})
	if err != nil {
		return errors.Wrapf(err, "unable to find setting for gloo %v", ingress.Metadata.Ref().Key())
	}

	for _, settings := range glooSettings {
		settings.Linkerd = linkerdEnabled
	}

	settingsReconciler := gloov1.NewSettingsReconciler(pl.cs.Supergloo.Settings)
	return settingsReconciler.Reconcile(ingress.InstallationNamespace, glooSettings, nil, clients.ListOpts{})
}
