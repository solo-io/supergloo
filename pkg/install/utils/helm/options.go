package helm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/solo-io/go-utils/errors"
)

type Options interface {
	// type of install
	Type() string
	// uri of chart
	Uri() string
	// version to install
	Version() string
	// namespace to install into
	Namespace() string
	// Default namespace to override in helm template
	NamespaceOverride() string
	// validator function for opts
	Validate() error
	// previous install
	PreviousInstall() Manifests
	// installer which should be called
	Installer() Installer
	// template which is called with opts, and passed into render as values
	HelmValuesTemplate() string
}

func InstallOrUpdate(ctx context.Context, opts Options) (Manifests, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid install options")
	}
	version := opts.Version()
	namespace := opts.Namespace()

	helmValueOverrides, err := template.New(fmt.Sprintf("%s-%s", opts.Type(), version)).Parse(opts.HelmValuesTemplate())
	if err != nil {
		return nil, errors.Wrapf(err, "")
	}

	valuesBuf := &bytes.Buffer{}
	if err := helmValueOverrides.Execute(valuesBuf, opts); err != nil {
		return nil, errors.Wrapf(err, "internal error: rendering helm values")
	}

	manifests, err := RenderManifests(
		ctx,
		opts.Uri(),
		valuesBuf.String(),
		releaseName(opts.Type(), namespace, version),
		namespace,
		"",
		true,
	)

	// nothing to do if the manifest hasn't changed
	if opts.PreviousInstall().CombinedString() == manifests.CombinedString() {
		return manifests, nil
	}

	for i, m := range manifests {
		// replace all instances of gloo-system with the target namespace just in case
		m.Content = strings.Replace(m.Content, opts.NamespaceOverride(), namespace, -1)
		manifests[i] = m
	}

	// perform upgrade instead
	if len(opts.PreviousInstall()) > 0 {
		if err := opts.Installer().UpdateFromManifests(ctx, namespace, opts.PreviousInstall(), manifests, true); err != nil {
			return nil, errors.Wrapf(err, "creating %s from manifests", opts.Type())
		}
	} else {
		if err := opts.Installer().CreateFromManifests(ctx, namespace, manifests); err != nil {
			return nil, errors.Wrapf(err, "creating %s from manifests", opts.Type())
		}
	}

	return manifests, nil
}

func releaseName(installType, namespace, version string) string {
	return fmt.Sprintf("%s-%s%s", installType, namespace, version)
}
