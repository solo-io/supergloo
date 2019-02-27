package flagutils

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/spf13/pflag"
)

func AddIstioInstallFlags(set *pflag.FlagSet, in *options.InputInstall) {
	set.StringVar(&in.IstioInstall.InstallationNamespace,
		"installation-namespace",
		"istio-system",
		"which namespace to install Istio into?")

	set.StringVar(&in.IstioInstall.IstioVersion,
		"version",
		istio.IstioVersion105,
		fmt.Sprintf("version of istio to install? available: %v", []string{
			istio.IstioVersion103,
			istio.IstioVersion105,
		}))

	set.BoolVar(&in.IstioInstall.EnableMtls,
		"mtls",
		true,
		"enable mtls?")

	set.BoolVar(&in.IstioInstall.EnableAutoInject,
		"auto-inject",
		true,
		"enable auto-injection?")

	set.BoolVar(&in.IstioInstall.InstallGrafana,
		"grafana",
		true,
		"add grafana to the install?")

	set.BoolVar(&in.IstioInstall.InstallPrometheus,
		"prometheus",
		true,
		"add prometheus to the install?")

	set.BoolVar(&in.IstioInstall.InstallJaeger,
		"jaeger",
		true,
		"add jaeger to the install?")

}
