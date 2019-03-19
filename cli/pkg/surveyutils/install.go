package surveyutils

import (
	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/supergloo/cli/pkg/constants"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveyIstioInstall(in *options.Install) error {
	if err := cliutil.ChooseFromList("which namespace to install to? ", &in.InstallationNamespace.Istio, helpers.MustGetNamespaces()); err != nil {
		return err
	}
	if err := cliutil.ChooseFromList("which version of Istio to install? ", &in.IstioInstall.IstioVersion, constants.SupportedIstioVersions); err != nil {
		return err
	}

	if err := cliutil.GetBoolInput("enable mtls? ", &in.IstioInstall.EnableMtls); err != nil {
		return err
	}

	if err := cliutil.GetBoolInput("enable auto-injection? ", &in.IstioInstall.EnableAutoInject); err != nil {
		return err
	}

	if err := cliutil.GetBoolInput("add grafana to the install? ", &in.IstioInstall.InstallGrafana); err != nil {
		return err
	}

	if err := cliutil.GetBoolInput("add prometheus to the install? ", &in.IstioInstall.InstallPrometheus); err != nil {
		return err
	}

	if err := cliutil.GetBoolInput("add jaeger to the install? ", &in.IstioInstall.InstallJaeger); err != nil {
		return err
	}

	if err := cliutil.GetBoolInput("update an existing install? ", &in.Update); err != nil {
		return err
	}

	return nil
}

func SurveyGlooIngressInstall(in *options.Install) error {
	if err := cliutil.ChooseFromList("which namespace to install to? ", &in.InstallationNamespace.Gloo, helpers.MustGetNamespaces()); err != nil {
		return err
	}
	if err := cliutil.ChooseFromList("which version of Gloo to install? ", &in.GlooIngressInstall.GlooVersion, constants.SupportedGlooVersions); err != nil {
		return err
	}

	if err := cliutil.GetBoolInput("update an existing install? ", &in.Update); err != nil {
		return err
	}

	return nil
}
