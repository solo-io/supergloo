package gloo

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kubeinstall"
	"github.com/solo-io/supergloo/pkg/util"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

const (
	customResourceDefinition = "CustomResourceDefinition"
	upstreamCrdName          = "upstreams.gloo.solo.io"
	settingsCrdName          = "settings.gloo.solo.io"
)

type Installer interface {
	EnsureGlooInstall(ctx context.Context, mesh *v1.MeshIngress) error
}

type glooInstaller struct {
	kubeInstaller kubeinstall.Installer
}

func newGlooInstaller(kubeInstaller kubeinstall.Installer) *glooInstaller {
	return &glooInstaller{kubeInstaller: kubeInstaller}
}

func (i *glooInstaller) EnsureGlooInstall(ctx context.Context, ingress *v1.MeshIngress) error {
	ctx = contextutils.WithLogger(ctx, "gloo-ingress-installer")
	logger := contextutils.LoggerFrom(ctx)

	gloo := ingress.GetGloo()
	if gloo == nil {
		return errors.Errorf("%v: invalid install type, only gloo ingress supported currently", ingress.Metadata.Ref().Key())
	}

	logger.Infof("syncing gloo install %v with config %v", ingress.Metadata.Ref().Key(), gloo)

	if gloo.Install.Options.Disabled {
		logger.Infof("purging resources for disabled install %v", ingress.Metadata.Ref().Key())
		if err := i.kubeInstaller.PurgeResources(ctx, util.LabelsForResource(ingress)); err != nil {
			return errors.Wrapf(err, "uninstalling gloo")
		}
		return nil
	}

	manifests, err := makeManifestsForInstall(ctx, ingress, gloo.Install)
	if err != nil {
		return err
	}

	rawResources, err := manifests.ResourceList()
	if err != nil {
		return err
	}

	// filter out upstreams, supergloo installs them
	rawResources = rawResources.Filter(func(resource *unstructured.Unstructured) bool {
		if resource.GroupVersionKind().Kind == customResourceDefinition {
			return resource.GetName() == upstreamCrdName || resource.GetName() == settingsCrdName
		}
		return false
	})

	installNamespace := gloo.Install.Options.InstallationNamespace

	logger.Infof("installing gloo with options: %#v", gloo)
	if err := i.kubeInstaller.ReconcileResources(ctx, installNamespace, rawResources, util.LabelsForResource(ingress)); err != nil {
		return errors.Wrapf(err, "reconciling install resources failed")
	}

	return nil
}

func makeManifestsForInstall(ctx context.Context, ingress *v1.MeshIngress, gloo *v1.GlooInstall) (helmchart.Manifests, error) {
	if gloo.Options.InstallationNamespace == "" {
		return nil, errors.Errorf("must provide installation namespace")
	}
	if gloo.Options.InstallationNamespace == "" {
		return nil, errors.Errorf("must provide gloo version")
	}

	manifests, err := helmchart.RenderManifests(ctx,
		glooManifestUrl(gloo.Options.Version),
		helmValues,
		"gloo", // release name used in some manifests for rendering
		gloo.Options.InstallationNamespace,
		"", // use default kube version
	)
	if err != nil {
		return nil, errors.Wrapf(err, "rendering install manifests")
	}

	return manifests, nil
}

func glooManifestUrl(version string) string {
	var url = "https://storage.googleapis.com/solo-public-helm/charts/gloo-%s.tgz"
	return fmt.Sprintf(url, version)
}
