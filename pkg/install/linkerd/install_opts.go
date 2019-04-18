package linkerd

import (
	"bytes"
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/installutils/kubeinstall"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	Version_stable230 = "stable-2.3.0"
)

var supportedVersions = []string{Version_stable230}

type versionedInstallOpts interface {
	install(ctx context.Context, installer kubeinstall.Installer, withLabels map[string]string) error
}

type installOpts struct {
	installVersion   string
	installNamespace string
	enableMtls       bool
	enableAutoInject bool
}

func newInstallOpts(installVersion string, installNamespace string, enableMtls bool, enableAutoInject bool) *installOpts {
	return &installOpts{installVersion: installVersion, installNamespace: installNamespace, enableMtls: enableMtls, enableAutoInject: enableAutoInject}
}

func (o *installOpts) install(ctx context.Context, installer kubeinstall.Installer, withLabels map[string]string) error {
	uri, err := o.chartURI()
	if err != nil {
		return err
	}

	injector, values, err := o.values()
	if err != nil {
		return err
	}

	manifests, err := helmchart.RenderManifests(ctx, uri, values, "linkerd2", o.installNamespace, "")
	if err != nil {
		return err
	}

	manifests, err = injectManifests(injector, manifests)
	if err != nil {
		return err
	}

	resources, err := manifests.ResourceList()
	if err != nil {
		return err
	}

	// filter out the install namespace, it's created by the custom installer
	resources = resources.Filter(func(resource *unstructured.Unstructured) bool {
		return resource.GroupVersionKind().Kind == "Namespace"
	})

	contextutils.LoggerFrom(ctx).Infof("installing linkerd with options: %#v", o)

	if err := installer.ReconcileResources(ctx, o.installNamespace, resources, withLabels); err != nil {
		return err
	}

	return nil
}

func injectManifests(injector injector, in helmchart.Manifests) (helmchart.Manifests, error) {
	input := bytes.NewBufferString(in.CombinedString())
	out := &bytes.Buffer{}
	err := processYAML(input, out, injector)
	if err != nil {
		return nil, err
	}

	return helmchart.Manifests{{Name: "linkerd-injected-manifest", Content: out.String()}}, nil
}
