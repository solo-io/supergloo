package surveyutils

import (
	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/gloo/projects/gloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/install/istio"
)

func SurveyIstioInstall(in *options.InputInstall) error {
	if err := cliutil.ChooseFromList("which namespace to install to? ", &in.IstioInstall.InstallationNamespace, helpers.MustGetNamespaces()); err != nil {
		return err
	}
	if err := cliutil.ChooseFromList("which version of Istio to install? ", &in.IstioInstall.IstioVersion, []string{
		istio.IstioVersion103,
		istio.IstioVersion105,
	}); err != nil {
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

	return nil
}
