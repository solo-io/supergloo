package helm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"k8s.io/helm/pkg/manifest"

	"github.com/solo-io/go-utils/errors"
)

type Options interface {
	// type of install
	InstallName() string
	// uri of chart
	ChartUri() string
	// version to install
	Version() string
	// namespace to install into
	Namespace() string
	// validator function for opts
	Validate() error
	// previous install
	PreviousInstall() Manifests
	// installer which should be called
	Installer() Installer
	// template which is called with opts, and passed into render as values
	HelmValuesTemplate() string
}

type ManifestFilterFunc func(input []manifest.Manifest) (output []manifest.Manifest, err error)

// Returns only non-empty manifests
var ExcludeEmptyManifests ManifestFilterFunc = func(input []manifest.Manifest) ([]manifest.Manifest, error) {
	var output []manifest.Manifest
	for _, manifest := range input {
		if !isEmptyManifest(manifest.Content) {
			output = append(output, manifest)
		}

	}
	return output, nil
}

func ReplaceHardcodedNamespace(hardcoded, override string) ManifestFilterFunc {
	var ReplaceFunc ManifestFilterFunc = func(input []manifest.Manifest) ([]manifest.Manifest, error) {
		for i, m := range input {
			// replace all instances of gloo-system with the target namespace just in case
			m.Content = strings.Replace(m.Content, hardcoded, override, -1)
			input[i] = m
		}
		return input, nil
	}
	return ReplaceFunc
}

func InstallOrUpdate(ctx context.Context, opts Options, filterFuncs ...ManifestFilterFunc) (Manifests, error) {
	if err := opts.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid install options")
	}
	version := opts.Version()
	namespace := opts.Namespace()

	helmValueOverrides, err := template.New(fmt.Sprintf("%s-%s", opts.InstallName(), version)).Parse(opts.HelmValuesTemplate())
	if err != nil {
		return nil, errors.Wrapf(err, "")
	}

	valuesBuf := &bytes.Buffer{}
	if err := helmValueOverrides.Execute(valuesBuf, opts); err != nil {
		return nil, errors.Wrapf(err, "internal error: rendering helm values")
	}

	manifests, err := RenderManifests(
		ctx,
		opts.ChartUri(),
		valuesBuf.String(),
		releaseName(opts.InstallName(), namespace, version),
		namespace,
		"",
		true,
	)

	// nothing to do if the manifest hasn't changed
	if opts.PreviousInstall().CombinedString() == manifests.CombinedString() {
		return manifests, nil
	}

	if filterFuncs != nil {
		for _, filterFunc := range filterFuncs {
			manifests, err = filterFunc(manifests)
			if err != nil {
				return nil, err
			}
		}
	}

	// perform upgrade instead
	if len(opts.PreviousInstall()) > 0 {
		if err := opts.Installer().UpdateFromManifests(ctx, namespace, opts.PreviousInstall(), manifests, true); err != nil {
			return nil, errors.Wrapf(err, "creating %s from manifests", opts.InstallName())
		}
	} else {
		if err := opts.Installer().CreateFromManifests(ctx, namespace, manifests); err != nil {
			return nil, errors.Wrapf(err, "creating %s from manifests", opts.InstallName())
		}
	}

	return manifests, nil
}

func releaseName(installType, namespace, version string) string {
	return fmt.Sprintf("%s-%s%s", installType, namespace, version)
}
