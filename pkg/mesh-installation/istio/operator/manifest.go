package operator

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/rotisserie/eris"
	operator_manifests "github.com/solo-io/service-mesh-hub/pkg/mesh-installation/istio/operator/manifests"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	validProfiles = sets.NewString(
		"demo",
		"default",
		"minimal",
		"sds",
		"remote",
	)

	InvalidProfileFound = func(profile string) error {
		return eris.Errorf(
			"invalid profile (%s) found, valid options are: [%s]",
			profile,
			strings.Join(validProfiles.List(), ","),
		)
	}
)

func NewInstallerManifestBuilder() InstallerManifestBuilder {
	return &installerManifestBuilder{}
}

type installerManifestBuilder struct{}

type operatorDeploymentManifestValues struct {
	CreateNamespace  bool
	InstallNamespace string
}

func (i *installerManifestBuilder) BuildOperatorDeploymentManifest(istioVersion IstioVersion, installNamespace string, createNamespace bool) (string, error) {
	tmpl := template.New("")

	var templateContents string
	switch istioVersion {
	case IstioVersion1_5:
		templateContents = operator_manifests.Operator1_5
	case IstioVersion1_6:
		templateContents = operator_manifests.Operator1_6
	default:
		return "", eris.Errorf("Unknown istio version: %s", string(istioVersion))
	}

	tmpl, err := tmpl.Parse(templateContents)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, &operatorDeploymentManifestValues{
		CreateNamespace:  createNamespace,
		InstallNamespace: installNamespace,
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (i *installerManifestBuilder) BuildOperatorConfigurationWithProfile(profile, namespace string) (string, error) {
	if !validProfiles.Has(profile) {
		return "", InvalidProfileFound(profile)
	}

	tmpl := template.New("")
	tmpl, err := tmpl.Parse(operator_manifests.OperatorWithProfile)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, &controlPlaneData{
		Profile:          profile,
		InstallNamespace: namespace,
	})
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

type controlPlaneData struct {
	Profile          string
	InstallNamespace string
}
