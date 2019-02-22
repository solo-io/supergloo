package flagutils

import (
	"fmt"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/pkg/install/istio"
	"github.com/spf13/pflag"
)

func AddIstioInstallFlags(set *pflag.FlagSet, in *options.InputInstall) error {
	set.StringVar(&in.IstioInstall.InstallationNamespace,
		"installation-namespace",
		"istio-system",
		"which namespace to install Istio into?")

	set.StringVar(&in.IstioInstall.IstioVersion,
		"version",
		"istio-system",
		fmt.Sprintf("version of istio to install? available: %v", []string{
			istio.IstioVersion103,
			istio.IstioVersion105,
		}))

	set.BoolVar(&in.IstioInstall.EnableMtls,
		"version",
		true,
		"enable mtls?")

	set.BoolVar(&in.IstioInstall.EnableAutoInject,
		"version",
		true,
		"enable auto-injection?")

	set.BoolVar(&in.IstioInstall.InstallGrafana,
		"version",
		true,
		"add grafana to the install?")


	set.BoolVar(&in.IstioInstall.InstallPrometheus,
		"version",
		true,
		"add prometheus to the install?")


	set.BoolVar(&in.IstioInstall.InstallJaeger,
		"version",
		true,
		"add jaeger to the install?")

	return nil
}
