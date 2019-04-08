package gloo

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/solo-io/supergloo/pkg/install/utils/helmchart"
	"github.com/solo-io/supergloo/pkg/install/utils/kubeinstall"
	"github.com/solo-io/supergloo/pkg/util"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

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
	EnsureGlooInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList, meshIngresses v1.MeshIngressList) (*v1.MeshIngress, error)
}

type glooInstaller struct {
	kubeInstaller kubeinstall.Installer
}

func newGlooInstaller(kubeInstaller kubeinstall.Installer) *glooInstaller {
	return &glooInstaller{kubeInstaller: kubeInstaller}
}

func (i *glooInstaller) EnsureGlooInstall(ctx context.Context, install *v1.Install, meshes v1.MeshList, meshIngresses v1.MeshIngressList) (*v1.MeshIngress, error) {
	ctx = contextutils.WithLogger(ctx, "gloo-ingress-installer")
	logger := contextutils.LoggerFrom(ctx)

	installIngress := install.GetIngress()
	if installIngress == nil {
		return nil, errors.Errorf("non ingress install detected in ingress install, %v", install.Metadata.Ref())
	}

	glooInstall := installIngress.GetGloo()
	if glooInstall == nil {
		return nil, errors.Errorf("%v: invalid install type, only gloo ingress supported currently", install.Metadata.Ref())
	}

	logger.Infof("syncing gloo install %v with config %v", install.Metadata.Ref(), glooInstall)

	if install.Disabled {
		logger.Infof("purging resources for disabled install %v", install.Metadata.Ref())
		if err := i.kubeInstaller.PurgeResources(ctx, util.LabelsForResource(install)); err != nil {
			return nil, errors.Wrapf(err, "uninstalling istio")
		}
		installIngress.InstalledIngress = nil
		return nil, nil
	}

	manifests, err := makeManifestsForInstall(ctx, install, glooInstall)
	if err != nil {
		return nil, err
	}

	rawResources, err := manifests.ResourceList()
	if err != nil {
		return nil, err
	}

	// filter out upstreams, supergloo installs them
	rawResources = rawResources.Filter(func(resource *unstructured.Unstructured) bool {
		if resource.GroupVersionKind().Kind == customResourceDefinition {
			return resource.GetName() == upstreamCrdName || resource.GetName() == settingsCrdName
		}
		return false
	})

	installNamespace := install.InstallationNamespace

	var meshIngress *v1.MeshIngress
	if installIngress.InstalledIngress != nil {
		var err error
		meshIngress, err = meshIngresses.Find(installIngress.InstalledIngress.Strings())
		if err != nil {
			return nil, errors.Wrapf(err, "installed ingress not found")
		}
	}

	var meshRefs []*core.ResourceRef
	for _, glooMesh := range glooInstall.Meshes {
		mesh, err := meshes.Find(glooMesh.Namespace, glooMesh.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "target mesh not found")
		}
		ref := mesh.Metadata.Ref()
		meshRefs = append(meshRefs, &ref)
	}

	logger.Infof("installing gloo with options: %#v", glooInstall)
	if err := i.kubeInstaller.ReconcilleResources(ctx, installNamespace, rawResources, util.LabelsForResource(install)); err != nil {
		return nil, errors.Wrapf(err, "reconciling install resources failed")
	}

	meshIngress = createOrUpdateMeshIngress(meshIngress, install, glooInstall, meshRefs)

	// caller should expect the install to have been modified
	ref := meshIngress.Metadata.Ref()
	installIngress.InstalledIngress = &ref

	return meshIngress, nil
}

func createOrUpdateMeshIngress(meshIngress *v1.MeshIngress, install *v1.Install, glooInstall *v1.GlooInstall, meshRefs []*core.ResourceRef) *v1.MeshIngress {

	if meshIngress != nil {
		meshIngress.Meshes = meshRefs
		meshIngress.InstallationNamespace = install.InstallationNamespace
		meshIngress.MeshIngressType = &v1.MeshIngress_Gloo{
			Gloo: &v1.GlooMeshIngress{},
		}
		return meshIngress
	}
	return &v1.MeshIngress{
		Metadata: core.Metadata{
			Namespace: install.Metadata.Namespace,
			Name:      install.Metadata.Name,
		},
		InstallationNamespace: install.InstallationNamespace,
		MeshIngressType: &v1.MeshIngress_Gloo{
			Gloo: &v1.GlooMeshIngress{},
		},
		Meshes: meshRefs,
	}
}

func makeManifestsForInstall(ctx context.Context, install *v1.Install, gloo *v1.GlooInstall) (helmchart.Manifests, error) {
	if install.InstallationNamespace == "" {
		return nil, errors.Errorf("must provide installation namespace")
	}
	if gloo.GlooVersion == "" {
		return nil, errors.Errorf("must provide gloo version")
	}

	manifests, err := helmchart.RenderManifests(ctx,
		glooManifestUrl(gloo.GlooVersion),
		helmValues,
		"gloo", // release name used in some manifests for rendering
		install.InstallationNamespace,
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
