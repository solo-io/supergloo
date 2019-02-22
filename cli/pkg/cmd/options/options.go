package options

import (
	"context"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

type Options struct {
	// common
	Ctx         context.Context
	Interactive bool
	OutputType  string

	Create Create
}

type Create struct {
	Metadata     core.Metadata
	InputInstall InputInstall
}

type InputInstall struct {
	IstioInstall v1.Install_Istio
}
