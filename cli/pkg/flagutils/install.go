package flagutils

import (
	"fmt"
	"time"

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

	set.DurationVar(&in.InstallTimeout,
		"update",
		time.Minute*5,
		"maximum time to wait for a mesh installation to complete")

}

func AddIstioInstallFlags(set *pflag.FlagSet, in *options.Install) {
	set.StringVar(&in.InstallationNamespace.Istio,
		"installation-namespace",
		"istio-system",
		"which namespace to install Istio into?")

	set.StringVar(&in.IstioInstall.Version,
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

	set.BoolVar(&in.IstioInstall.EnableIngress,
		"ingress",
		false,
		"enable ingress gateway?",
	)

	set.BoolVar(&in.IstioInstall.EnableEgress,
		"egress",
		false,
		"enable egress gateway?",
	)

	set.BoolVar(&in.IstioInstall.InstallGrafana,
		"grafana",
		false,
		"add grafana to the install?")

	set.BoolVar(&in.IstioInstall.InstallPrometheus,
		"prometheus",
		false,
		"add prometheus to the install?")

	set.BoolVar(&in.IstioInstall.InstallJaeger,
		"jaeger",
		false,
		"add jaeger to the install?")

	set.BoolVar(&in.IstioInstall.InstallSmiAdapter,
		"smi-install",
		false,
		"add the SMI adapter to the install?")

}

func AddLinkerdInstallFlags(set *pflag.FlagSet, in *options.Install) {
	set.StringVar(&in.InstallationNamespace.Linkerd,
		"installation-namespace",
		"linkerd",
		"which namespace to install Linkerd into?")

	set.StringVar(&in.LinkerdInstall.Version,
		"version",
		linkerd.Version_stable230,
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

	set.StringVar(&in.GlooIngressInstall.Version,
		"version",
		"latest",
		fmt.Sprintf("version of gloo to install? available: %v", constants.SupportedGlooVersions))

	set.VarP(&in.MeshIngress.Meshes, "target-meshes", "t", "Which meshes to target (comma seperated list) <namespace>.<name>")
}
