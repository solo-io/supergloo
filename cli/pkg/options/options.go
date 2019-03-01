package options

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type Options struct {
	// common
	Ctx         context.Context
	Interactive bool
	OutputType  string

	Init      Init
	Install   Create
	Uninstall Uninstall
}

type Init struct {
	HelmChartOverride string
	HelmValues        string
	InstallNamespace  string
	ReleaseVersion    string
	DryRun            bool
}

type Create struct {
	Metadata     core.Metadata
	InputInstall InputInstall
}

type InputInstall struct {
	IstioInstall v1.Install_Istio
}

type Uninstall struct {
	Metadata core.Metadata
}
