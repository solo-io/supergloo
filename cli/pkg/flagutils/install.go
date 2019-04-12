package flagutils

import (
	"fmt"

	"github.com/solo-io/supergloo/pkg/install/linkerd"

	"github.com/solo-io/supergloo/pkg/install/istio"

	"github.com/solo-io/supergloo/cli/pkg/constants"

	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddInstallFlags(set *pflag.FlagSet, in *options.Install) {

	set.BoolVar(&in.Update,
		"update",
		false,
		"update an existing install?")

}

func AddIstioInstallFlags(set *pflag.FlagSet, in *options.Install) {
	set.StringVar(&in.InstallationNamespace.Istio,
		"installation-namespace",
		"istio-system",
		"which namespace to install Istio into?")

	set.StringVar(&in.IstioInstall.IstioVersion,
		"version",
		istio.IstioVersion106,
		fmt.Sprintf("version of istio to install? available: %v", constants.SupportedIstioVersions))

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

func AddLinkerdInstallFlags(set *pflag.FlagSet, in *options.Install) {
	set.StringVar(&in.InstallationNamespace.Linkerd,
		"installation-namespace",
		"linkerd-system",
		"which namespace to install Linkerd into?")

	set.StringVar(&in.LinkerdInstall.LinkerdVersion,
		"version",
		linkerd.Version_stable221,
		fmt.Sprintf("version of linkerd to install? available: %v", constants.SupportedLinkerdVersions))

	set.BoolVar(&in.LinkerdInstall.EnableMtls,
		"mtls",
		true,
		"enable mtls?")

	set.BoolVar(&in.LinkerdInstall.EnableAutoInject,
		"auto-inject",
		true,
		"enable auto-injection?")
}

func AddGlooIngressInstallFlags(set *pflag.FlagSet, in *options.Install) {

	set.StringVar(&in.InstallationNamespace.Gloo,
		"installation-namespace",
		"gloo-system",
		"which namespace to install Gloo into?")

	set.StringVar(&in.GlooIngressInstall.GlooVersion,
		"version",
		"latest",
		fmt.Sprintf("version of gloo to install? available: %v", constants.SupportedGlooVersions))

	set.VarP(&in.MeshIngress.Meshes, "target-meshes", "t", "Which meshes to target (comma seperated list) <namespace>.<name>")
}
